package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/internal/hydra"
	"git.tor.ph/hiveon/idp/models/users"
	"github.com/davecgh/go-spew/spew"

	//"github.com/gorilla/csrf"
	"io/ioutil"
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	renderPkg "github.com/unrolled/render"
	"github.com/volatiletech/authboss"
	"golang.org/x/oauth2"
	"gopkg.in/resty.v1"
)

var oauthClient *oauth2.Config
var render *renderPkg.Render

func init() {
	render = renderPkg.New()
}

func acceptConsent(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get("consent_challenge")

	if len(challenge) == 0 {
		ro := authboss.RedirectOptions{
			Code:         http.StatusTemporaryRedirect,
			RedirectPath: "/",
			Failure:      "You have no consent challenge",
		}
		ab.Core.Redirector.Redirect(w, r, ro)
		return
	}

	url, err := hydra.AcceptConsentChallengeCode(challenge)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		logrus.Debugf("consent challenge code isn't right")
		return
	}

	logrus.Debugf("Consent code accepted")

	var oauth2_consent_csrf *http.Cookie
	oauth2_auth_csrf, _ := r.Cookie(cookieAuthenticationCSRFName)

	k := r.Cookies()
	for i, v := range k {
		if v.Name == cookieConsentCSRFName {
			oauth2_consent_csrf = k[i]
		}
	}

	res, err := resty.
		SetCookie(oauth2_consent_csrf).
		SetCookie(oauth2_auth_csrf).
		R().
		SetHeader("Accept", "application/json").
		Get(url)

	if err != nil {
		render.JSON(w, 422, &ResponseError{
			Status:  "error",
			Success: false,
			Error:   "no consent csrf token has been provided",
		})
	}

	accessToken := res.RawResponse.Header.Get("Set-Cookie")
	splitToken := strings.Split(accessToken, " ")
	if len(splitToken) < 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		logrus.Error("Can't obtain access token!")
		return
	}

	// accessToken = formatToken(accessToken)
	w.Header().Set("access_token", splitToken[1])
}

func getUserFromHydraSession(w http.ResponseWriter, r *http.Request) (authboss.User, error) {
	hydraConfig, _ := config.GetHydraConfig()
	reqTokenCookie, err := r.Cookie("Authorization")
	if err != nil {
		return nil, errors.New("Authorization token missed")
	}

	reqToken := reqTokenCookie.Value

	if len(reqToken) == 0 {
		return nil, errors.New("Authorization token missed")
	}

	splitToken := strings.Split(reqToken, " ")
	if len(splitToken) < 1 {
		return nil, errors.New("Token is wrong")
	}

	token := strings.TrimSpace(splitToken[1])
	introspectURL := fmt.Sprintf("%s/oauth2/introspect", hydraConfig.Admin)

	res, err := resty.R().SetFormData(map[string]string{"token": token}).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").Post(introspectURL)

	if err != nil {
		return nil, errors.New("Can't check token")
	}
	var introToken swagger.OAuth2TokenIntrospection

	err = json.Unmarshal(res.Body(), &introToken)
	if err != nil {
		return nil, errors.New("Can't unmarshall token")
	}

	user, err := getAuthbossUserByEmail(r, introToken.Sub)
	if err != nil {
		return nil, errors.New("can't find user")
	}

	// RefreshToken(w, r, user)

	return user, nil
}

// RefreshToken refreshing token via hydra for specified user
func RefreshToken(w http.ResponseWriter, r *http.Request, abUser authboss.User) {
	user := abUser.(*users.User)
	refreshToken := user.GetOAuth2RefreshToken()
	accessToken := user.GetOAuth2AccessToken()
	expiry := user.GetOAuth2Expiry()

	if refreshToken == "" {
		http.Error(w, "No refresh token", http.StatusForbidden)
		return
	}
	token := oauth2.Token{RefreshToken: refreshToken, AccessToken: accessToken, Expiry: expiry}
	updatedToken, _ := oauthClient.TokenSource(context.TODO(), &token).Token()

	if accessToken != updatedToken.AccessToken {
		user.PutOAuth2AccessToken(updatedToken.AccessToken)
		user.PutOAuth2RefreshToken(updatedToken.RefreshToken)
		user.PutOAuth2Expiry(updatedToken.Expiry)

		ab.Config.Storage.Server.Save(r.Context(), user)
	}

	SetAccessTokenCookie(w, updatedToken.AccessToken)

	render.JSON(w, 200, map[string]string{
		"access_token": updatedToken.AccessToken,
	})
}

func challengeCode(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get("login_challenge")
	k := r.Cookies()
	fmt.Println(k)
	if len(challenge) == 0 { // obtain login challenge
		// move to auth
		hydraConfig, _ := config.GetHydraConfig()
		oauthClient = InitClient(hydraConfig.ClientID, hydraConfig.ClientSecret)
		redirectUrl := oauthClient.AuthCodeURL("state123")

		render.JSON(w, 200, map[string]string{"redirectURL": redirectUrl})
		return
	}

	challengeResp, err := hydra.CheckChallengeCode(challenge)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		logrus.Debugf("wrong login challenge")
		return
	}
	challengeCode := challengeResp.Challenge
	authboss.PutSession(w, "Challenge", challengeCode)

	render.JSON(w, 200, map[string]string{"challenge": challengeCode})
	return
}

func callbackToken(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	fmt.Println("Code: ", code)
	token, err := oauthClient.Exchange(oauth2.NoContext, code)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		logrus.Debugf("Can't obtain authorization token")
		return
	}

	//user, err := ab.LoadCurrentUser(&r)

	var introToken swagger.OAuth2TokenIntrospection
	hydraConfig, _ := config.GetHydraConfig()
	introspectUrl := hydraConfig.Introspect

	res, err := resty.R().SetFormData(map[string]string{"token": token.AccessToken}).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").Post(introspectUrl)

	err = json.Unmarshal(res.Body(), &introToken)
	user, err := ab.Storage.Server.Load(context.TODO(), introToken.Sub)

	if user != nil && err == nil {
		user1 := user.(*users.User)
		user1.PutOAuth2AccessToken(token.AccessToken)
		user1.PutOAuth2RefreshToken(token.RefreshToken)
		user1.PutOAuth2Expiry(token.Expiry)

		ab.Config.Storage.Server.Save(r.Context(), user1)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	//portalConfig, err := config.GetPortalConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	SetAccessTokenCookie(w, token.AccessToken)

	return

	//http.Redirect(w, r, portalConfig.Callback, http.StatusPermanentRedirect)
}

//nosurfing is a more verbose wrapper around csrf handling
func nosurfing(h http.Handler) http.Handler {
	surfing := nosurf.New(h)
	surfing.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Println("Failed to validate CSRF token:", nosurf.Reason(r))
		w.WriteHeader(http.StatusBadRequest)
	}))
	return surfing
}

func handleUserSession(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user, err := getUserFromHydraSession(w, r)
		if err != nil || user == nil {
			authboss.DelAllSession(w, ab.Config.Storage.SessionStateWhitelistKeys)
			authboss.DelKnownSession(w)
			authboss.DelKnownCookie(w)

			logrus.Error(err.Error())
		} else {
			r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyPID, user.(*users.User).Email))
		}
		handler.ServeHTTP(w, r)
	})
}

func dataInjector(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := layoutData(w, &r, "")
		r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))
		handler.ServeHTTP(w, r)
	})
}

// layoutData is passing pointers to pointers be able to edit the current pointer
// to the request. This is still safe as it still creates a new request and doesn't
// modify the old one, it just modifies what we're pointing to in our methods so
// we're able to skip returning an *http.Request everywhere
func layoutData(w http.ResponseWriter, r **http.Request, redirect string) authboss.HTMLData {
	currentUserName := ""
	userInter, err := ab.LoadCurrentUser(r)
	if userInter != nil && err == nil {
		currentUserName = userInter.(*users.User).Login
	}

	return authboss.HTMLData{
		"loggedin":          userInter != nil,
		"current_user_name": currentUserName,
		//"csrf_token":        nosurf.Token(*r),
		"flash_success": authboss.FlashSuccess(w, *r),
		"flash_error":   authboss.FlashError(w, *r),
		"redirectURL":   redirect,
	}
}

func debugMw(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("\n%s %s %s\n", r.Method, r.URL.Path, r.Proto)

		session, err := sessionStore.Get(r, IDPSessionName)
		if err == nil {
			fmt.Print("Session: ")
			first := true
			for k, v := range session.Values {
				if first {
					first = false
				} else {
					fmt.Print(", ")
				}
				fmt.Printf("%s = %v", k, v)
			}
			fmt.Println()
		}
		// fmt.Println("Database:")
		// for _, u := range database.Users {
		// 	fmt.Printf("! %#v\n", u)
		// }
		if val := r.Context().Value(authboss.CTXKeyData); val != nil {
			fmt.Printf("CTX Data: %s", spew.Sdump(val))
		}
		if val := r.Context().Value(authboss.CTXKeyValues); val != nil {
			fmt.Printf("CTX Values: %s", spew.Sdump(val))
		}

		handler.ServeHTTP(w, r)
	})
}

func InitClient(clientId string, secret string) *oauth2.Config {

	hydraConfig, _ := config.GetHydraConfig()
	hydraAPI := hydraConfig.API
	client := GetClient(clientId)
	oauthConfig := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  hydraAPI + "/oauth2/auth",
			TokenURL: hydraAPI + "/oauth2/token",
		},
		RedirectURL: client.RedirectUris[0],
		Scopes:      getScopes(),
	}

	return oauthConfig
}

func GetClient(clientId string) swagger.OAuth2Client {
	hydraConfig, _ := config.GetHydraConfig()
	hydraAdmin := hydraConfig.Admin

	clientUrl := hydraAdmin + "/clients/" + clientId
	res, err := resty.R().Get(clientUrl)
	if err != nil {
		log.Info(err)
	}
	var client swagger.OAuth2Client
	json.Unmarshal(res.Body(), &client)
	return client

}

func getScopes() []string {
	return []string{"openid", "offline"}
}

func getChallengeFromURL(r *http.Request, w http.ResponseWriter) (string, string) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	b := bodyBytes
	fromURLString := ""
	chalengeString := ""

	var t map[string]string
	json.Unmarshal(b, &t)

	if t["fromURL"] != "" {
		fromURLString = t["fromURL"]

	}
	if t["login_challenge"] != "" {
		chalengeString = t["login_challenge"]

	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return fromURLString, chalengeString
}

func setRedirectURL(redirectURL string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"redirectURL": %q}`, redirectURL)
	w.WriteHeader(http.StatusOK)
}
