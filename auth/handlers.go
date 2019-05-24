package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
	"github.com/volatiletech/authboss/auth"
	"github.com/volatiletech/authboss/otp/twofactor"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"

	"git.tor.ph/hiveon/idp/internal/hydra"
	"git.tor.ph/hiveon/idp/models/users"
	"github.com/go-chi/chi"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/authboss"
	"golang.org/x/oauth2"
	"gopkg.in/resty.v1"
)

func (a Auth) getChallengeCodeFromHydra(request *http.Request) (string, *http.Cookie, error) {
	oauthClient := initOauthClient(a.conf.Hydra)

	state, err := stateTokenGenerator()
	if err != nil {
		logrus.Error("login token failed generation")
		logrus.Debugf("server err, can't generate auth token")
		return "", nil, err
	}

	c := http.Cookie{
		Name:     cookieLoginState,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
	}

	request.AddCookie(&c)

	redirectUrl := oauthClient.AuthCodeURL(state)

	fmt.Println("redirectURL", redirectUrl)

	resp, err := resty.New().SetCookie(&c).R().Get(redirectUrl)
	if err != nil && !strings.Contains(err.Error(), "auto redirect is disabled") {
		fmt.Println("Go to redirect error", err)
		return "", nil, err
	}

	locationWithCode := resp.Header().Get("Location")

	code := regexpForChalangeCode.FindStringSubmatch(locationWithCode)

	if len(code) > 1 {
		for _, v := range resp.Cookies() {
			request.AddCookie(v)
		}

		return code[1], &c, nil
	}

	return "", nil, errors.New("no challenge code")
}

func (a Auth) callbackToken(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	fmt.Println("Code: ", code)

	oauthClient := initOauthClient(a.conf.Hydra)

	stateToken, err := r.Cookie(cookieLoginState)
	if err != nil {
		logrus.Infoln("state token absent")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		logrus.Debugf("Can't obtain authorization token\n")
		return
	}

	if stateToken.Value != state {
		logrus.Infof("invalid oauth state, cookie: '%s', URL: '%s'\n", stateToken.Value, state)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		logrus.Debugf("Can't obtain authorization token\n")
		return
	}

	token, err := oauthClient.Exchange(oauth2.NoContext, code)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		logrus.Debugf("Can't obtain authorization token")
		return
	}

	var introToken swagger.OAuth2TokenIntrospection
	introspectUrl := a.conf.Hydra.Introspect

	res, err := resty.R().SetFormData(map[string]string{"token": token.AccessToken}).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").Post(introspectUrl)

	err = json.Unmarshal(res.Body(), &introToken)
	user, err := a.authBoss.Storage.Server.Load(context.TODO(), introToken.Sub)

	if user != nil && err == nil {
		user1 := user.(*users.User)
		user1.PutOAuth2AccessToken(token.AccessToken)
		user1.PutOAuth2RefreshToken(token.RefreshToken)
		user1.PutOAuth2Expiry(token.Expiry)

		a.authBoss.Config.Storage.Server.Save(r.Context(), user1)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	SetAccessTokenCookie(w, token.AccessToken)
}

func (a Auth) acceptConsent(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get("consent_challenge")

	if len(challenge) == 0 {
		ro := authboss.RedirectOptions{
			Code:         http.StatusTemporaryRedirect,
			RedirectPath: "/",
			Failure:      "You have no consent challenge",
		}
		a.authBoss.Core.Redirector.Redirect(w, r, ro)
		return
	}

	url, err := hydra.AcceptConsentChallengeCode(challenge, a.conf.Hydra)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		logrus.Debugf("consent challenge code isn't right")
		return
	}

	logrus.Debugf("Consent code accepted")

	var oauth2ConsentCSRF *http.Cookie
	oauth2AuthCSRF, _ := r.Cookie(cookieAuthenticationCSRFName)

	k := r.Cookies()
	for i, v := range k {
		if v.Name == cookieConsentCSRFName {
			oauth2ConsentCSRF = k[i]
		}
	}

	res, err := resty.
		SetCookie(oauth2ConsentCSRF).
		SetCookie(oauth2AuthCSRF).
		R().
		SetHeader("Accept", "application/json").
		Get(url)

	if err != nil {
		a.render.JSON(w, http.StatusUnprocessableEntity, &ResponseError{
			Status:  "error",
			Success: false,
			Error:   "no consent csrf token has been provided",
		})
		return
	}

	accessToken := res.RawResponse.Header.Get("Set-Cookie")
	splitToken := strings.Split(accessToken, " ")
	if len(splitToken) < 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		logrus.Error("Can't obtain access token!")
		return
	}

	w.Header().Set("access_token", splitToken[1])
}

func (a Auth) handleLogin(challenge string, w http.ResponseWriter, r *http.Request) (bool, error) {
	if challenge == "" {
		a.render.JSON(w, http.StatusUnprocessableEntity, &ResponseError{
			Status:  "error",
			Success: false,
			Error:   "no challenge code has been provided",
		})

		return true, nil
	}

	user, err := a.authBoss.LoadCurrentUser(&r)
	if user != nil && err == nil {

		user := user.(*users.User)
		resp, errConfirm := hydra.ConfirmLogin(user.ID, false, challenge, a.conf.Hydra)
		if errConfirm != nil || resp.RedirectTo == "" {
			logrus.Debugf("probably challenge has been expired")
			a.render.JSON(w, http.StatusUnprocessableEntity, &ResponseError{
				Status:  "error",
				Success: false,
				Error:   "challenge code has been expired",
			})
			return true, nil
		}

		oauth2AuthCSRF, oauth2Err := r.Cookie(cookieAuthenticationCSRFName)
		loginStateToken, loginStateErr := r.Cookie(cookieLoginState)

		k := r.Cookies() // last
		for i, v := range k {
			if v.Name == cookieAuthenticationCSRFName {
				oauth2AuthCSRF = k[i]
			}
			if v.Name == cookieLoginState {
				loginStateToken = k[i]
			}
		}

		cookieArray := []*http.Cookie{}
		resty.DefaultClient.Cookies = cookieArray

		if oauth2Err != nil || loginStateErr != nil {
			if oauth2Err != nil {
				logrus.Infof("%s token absent! login rejected\n", cookieAuthenticationCSRFName)
			}
			if loginStateErr != nil {
				logrus.Infof("%s token absent! login rejected\n", cookieLoginState)
			}
			a.render.JSON(w, http.StatusUnprocessableEntity, &ResponseError{
				Status:  "error",
				Success: false,
				Error:   "auth token absent",
			})
			return true, nil
		}

		res, err := resty.
			SetCookie(oauth2AuthCSRF).
			SetCookie(loginStateToken).
			R().
			SetHeader("Accept", "application/json").
			Get(resp.RedirectTo)

		if err != nil {
			a.render.JSON(w, http.StatusUnprocessableEntity, &ResponseError{
				Status:  "error",
				Success: false,
				Error:   "no csrf token has been provided",
			})
			return true, nil
		}

		accessToken := res.RawResponse.Header.Get("access_token")
		if accessToken == "" {
			a.render.JSON(w, http.StatusUnprocessableEntity, &ResponseError{
				Status:  "error",
				Success: false,
				Error:   "No access token has been obtained",
			})
			return true, nil
		}

		SetAccessTokenCookie(w, accessToken)

		a.render.JSON(w, http.StatusOK, map[string]string{
			"access_token": accessToken,
			"token_type":   "bearer",
		})
	}
	return true, nil
}

func (a Auth) getRecoverSentURL(w http.ResponseWriter, r *http.Request) error {
	challenge, cookie, err := a.getChallengeCodeFromHydra(r)

	if err != nil {
		logrus.Error("can't get challenge code after register", err)
		return err
	}
	http.SetCookie(w, cookie)

	_, err = a.handleLogin(challenge, w, r)

	if err != nil {
		logrus.Error("can't login", err)
		return err
	}

	return nil
}

func (a Auth) getUserByEmail(w http.ResponseWriter, r *http.Request) {
	user, err := a.getAuthbossUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	a.render.JSON(w, http.StatusOK, user)
}

func (a Auth) refreshTokenByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	user, err := a.authBoss.Config.Storage.Server.Load(r.Context(), email)
	if err != nil {
		a.render.JSON(w, http.StatusInternalServerError, map[string]string{"error": "user not found"})
		return
	}

	a.RefreshToken(w, r, user)
}

func (a Auth) getUserInfo(w http.ResponseWriter, r *http.Request) {
	user, err := a.getUserFromHydraSession(w, r) // also refresh token if needed
	if err != nil {
		a.render.JSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	userMap, er := ToMap(user, "json")
	if er != nil {
		a.render.JSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	a.render.JSON(w, http.StatusOK, userMap)
	return
}

func (a *Auth) LoginPost(w http.ResponseWriter, r *http.Request)  {
	logger := a.authBoss.RequestLogger(r)

	validatable, err := a.authBoss.Core.BodyReader.Read(auth.PageLogin, r)
	if err != nil {
		return
	}

	// Skip validation since all the validation happens during the database lookup and
	// password check.
	creds := authboss.MustHaveUserValues(validatable)

	pid := creds.GetPID()
	pidUser, err := a.authBoss.Storage.Server.Load(r.Context(), pid)
	if err == authboss.ErrUserNotFound {
		logger.Infof("failed to load user requested by pid: %s", pid)
		return
	} else if err != nil {
		return
	}

	authUser := authboss.MustBeAuthable(pidUser)
	password := authUser.GetPassword()

	r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyUser, pidUser))

	var handled bool
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(creds.GetPassword()))
	if err != nil {
		handled, err = a.authBoss.Events.FireAfter(authboss.EventAuthFail, w, r)
		if err != nil {
			return
		}

		logger.Infof("user %s failed to log in", pid)
		a.render.JSON(w, http.StatusUnauthorized, &ResponseError{
			Status:  "error",
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyValues, validatable))

	handled, err = a.authBoss.Events.FireBefore(authboss.EventAuth, w, r)
	if err != nil {
		return
	} else if handled {
		return
	}

	if _, err := a.checkTOTPWhenLogin(w, r); err != nil {
		logger.Errorf("TOTP error %s", err)
		return
	}

	logger.Infof("user %s logged in", pid)
	authboss.PutSession(w, authboss.SessionKey, pid)
	authboss.DelSession(w, authboss.SessionHalfAuthKey)

	handled, err = a.authBoss.Events.FireAfter(authboss.EventAuth, w, r)
	if err != nil {
		return
	} else if handled {
		return
	}

	//HandleLogin with hydra
	challenge, cookie, err := a.getChallengeCodeFromHydra(r)
	if err != nil {
		logrus.Error("can't get challenge code after register", err)
		a.render.JSON(w, http.StatusUnprocessableEntity, &ResponseError{
			Status:  "error",
			Success: false,
			Error:   "Can't get challenge code after register",
		})
	}
	http.SetCookie(w, cookie)

	_, err = a.handleLogin(challenge, w, r)
}

func (a Auth) checkTOTPWhenLogin(w http.ResponseWriter, r *http.Request) (bool, error) {
	abUser, _ := a.authBoss.LoadCurrentUser(&r)
	user := abUser.(*users.User)

	if len(user.GetTOTPSecretKey()) == 0 {
		return false, nil
	}

	totpSecret := user.GetTOTPSecretKey()
	recoveryCode := user.Code2FA

	var ok bool

	recoveryCodes := twofactor.DecodeRecoveryCodes(user.GetRecoveryCodes())
	recoveryCodes, ok = twofactor.UseRecoveryCode(recoveryCodes, recoveryCode)

	if ok {
		//logger.Infof("user %s used recovery code instead of sms2fa", user.GetPID())
		user.PutRecoveryCodes(twofactor.EncodeRecoveryCodes(recoveryCodes))
		if err := a.authBoss.Config.Storage.Server.Save(r.Context(), user); err != nil {
			return false, err
		}
	}
	res := totp.Validate(recoveryCode, totpSecret)

	if !res {
		a.render.JSON(w, http.StatusBadRequest, &ResponseError{
			Status:  "error",
			Success: false,
			Error:   "2FA code is incorrect",
		})
		return false, errors.New("2FA code is incorrect")
	}

	return true, nil
}
