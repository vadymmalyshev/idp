package auth

import (
	"encoding/base32"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/volatiletech/authboss/otp/twofactor/totp2fa"
	"github.com/volatiletech/authboss/remember"
	"gopkg.in/resty.v1"

	"github.com/volatiletech/authboss/recover"

	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/internal/hydra"
	"git.tor.ph/hiveon/idp/models/users"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"github.com/volatiletech/authboss"
	"github.com/volatiletech/authboss/auth"
	"github.com/volatiletech/authboss/defaults"
	"github.com/volatiletech/authboss/register"

	renderPkg "github.com/unrolled/render"
	clientState "github.com/volatiletech/authboss-clientstate"
	abrenderer "github.com/volatiletech/authboss-renderer"
)

const IDPSessionName = "idp_session"

// SessionCookieMaxAge holds long an authenticated session should be valid in seconds
const SessionCookieMaxAge = 30 * 24 * 60 * 60

// SessionCookieHTTPOnly describes if the cookies should be accessible from HTTP requests only (no JS)
const SessionCookieHTTPOnly = false

const SessionCookieSecure = false

var (
	CookieDomain  string
	SessionCookie bool

	signingKey       string
	signingKeyBase32 string

	cookieStore  clientState.CookieStorer
	sessionStore clientState.SessionStorer

	cookieAuthenticationCSRFName = "oauth2_authentication_csrf"
	cookieConsentCSRFName        = "oauth2_consent_csrf"
)

var (
	ab *authboss.Authboss

	flagAPI = flag.Bool("api", true, "configure the app to be an api instead of an html app")
)

const (
	recoverSentURL = "/recover/sent"
	recoverSentTPL = "recover_sent"

	tplPath = "views/"
)

type ResponseError struct {
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Error   string `json:"errorMsg"`
}

func Init(r *gin.Engine, db *gorm.DB) {
	signingKey, _ := config.GetSignKey()

	signingKeyBytes := []byte(signingKey)
	signingKeyBase32 = base32.StdEncoding.EncodeToString(signingKeyBytes)

	// cookieAuthenticationKey := signingKeyBytes
	// cookieEncryptionKey := signingKeyBytes[:32]

	cookieStoreKey, _ := base64.StdEncoding.DecodeString(`NpEPi8pEjKVjLGJ6kYCS+VTCzi6BUuDzU0wrwXyf5uDPArtlofn2AG6aTMiPmN3C909rsEWMNqJqhIVPGP3Exg==`)
	sessionStoreKey, _ := base64.StdEncoding.DecodeString(`AbfYwmmt8UCwUuhd9qvfNA9UCuN1cVcKJN1ofbiky6xCyyBj20whe40rJa3Su0WOWLWcPpO1taqJdsEI/65+JA==`)

	cookieStore = clientState.NewCookieStorer(cookieStoreKey, nil)
	// cookieStore.MaxAge = SessionCookieMaxAge
	// cookieStore.HTTPOnly = SessionCookieHTTPOnly
	// cookieStore.Secure = SessionCookieSecure
	cookieStore.Domain = "localhost"
	cookieStore.HTTPOnly = false
	cookieStore.Secure = false

	sessionStore = clientState.NewSessionStorer(IDPSessionName, sessionStoreKey, nil)

	cstore := sessionStore.Store.(*sessions.CookieStore)
	// cstore.Options.HttpOnly = SessionCookieHTTPOnly
	// cstore.Options.Secure = SessionCookieSecure

	cstore.Options.HttpOnly = false
	cstore.Options.Secure = false

	ab = authboss.New()

	portalConfig, _ := config.GetPortalConfig()

	ab.Config.Paths.RootURL = portalConfig.Callback
	//ab.Config.Paths.Mount = "/"

	ab.Config.Storage.Server = users.NewUserStorer(db)
	ab.Config.Storage.SessionState = sessionStore
	ab.Config.Storage.CookieState = cookieStore
	ab.Config.Storage.SessionStateWhitelistKeys = []string{"Authorization", "oauth2_authentication_csrf", "access_token"}

	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}

	ab.Config.Modules.RegisterPreserveFields = []string{"email", "login", "name"}

	ab.Config.Modules.TOTP2FAIssuer = "HiveonID"
	ab.Config.Modules.TwoFactorEmailAuthRequired = false

	defaults.SetCore(&ab.Config, *flagAPI, false)

	ab.Config.Core.Mailer = NewMailer()
	ab.Config.Core.MailRenderer = abrenderer.NewEmail("/", tplPath)

	emailRule := defaults.Rules{
		FieldName: "email", Required: true,
		MatchError: "Must be a valid e-mail address",
		MustMatch:  regexp.MustCompile(`.*@.*\.[a-z]{1,}`),
	}
	passwordRule := defaults.Rules{
		FieldName: "password", Required: true,
		MinLength: 4,
	}
	nameRule := defaults.Rules{
		FieldName: "name", Required: false,
		AllowWhitespace: true,
	}
	loginRule := defaults.Rules{
		FieldName: "login", Required: true,
		MinLength: 2,
	}

	ab.Config.Core.BodyReader = defaults.HTTPBodyReader{
		ReadJSON: *flagAPI,
		Rulesets: map[string][]defaults.Rules{
			"register":      {emailRule, passwordRule, nameRule, loginRule},
			"recover_start": {emailRule},
			"recover_end":   {passwordRule},
		},
		Confirms: map[string][]string{
			// 	"register":    {"password", authboss.ConfirmPrefix + "password"},
			"recover_end": {"password", authboss.ConfirmPrefix + "password"},
		},
		Whitelist: map[string][]string{
			"register": []string{"email", "name", "login", "password"},
		},
	}

	ab.Config.Paths.RecoverOK = recoverSentURL
	// Load our template of recover sent message to AB renderer
	ab.Config.Core.ViewRenderer.Load(recoverSentTPL)
	// Handle recover sent
	ab.Config.Core.Router.Get(recoverSentURL, ab.Core.ErrorHandler.Wrap(func(w http.ResponseWriter, req *http.Request) error {
		return ab.Config.Core.Responder.Respond(w, req, http.StatusOK, recoverSentTPL, nil)
	}))

	modAuth := auth.Auth{}
	if err := modAuth.Init(ab); err != nil {
		logrus.Panicf("can't initialize authboss's auth mod", err)
	}

	modRegister := register.Register{}
	if err := modRegister.Init(ab); err != nil {
		logrus.Panicf("can't initialize authboss's register mod", err)
	}

	modRecover := recover.Recover{}
	if err := modRecover.Init(ab); err != nil {
		logrus.Panicf("can't initialize authboss's recover mod", err)
	}

	modTotp := &totp2fa.TOTP{Authboss: ab}
	if err := modTotp.Setup(); err != nil {
		logrus.Panicf("can't initialize authboss's totp2fa mod", err)
	}

	schemaDec := schema.NewDecoder()
	schemaDec.IgnoreUnknownKeys(true)
	render := renderPkg.New()

	mux := chi.NewRouter()

	mux.Use(ab.LoadClientStateMiddleware, remember.Middleware(ab))
	mux.Use(handleUserSession)
	mux.Use(dataInjector)
	mux.Use(checkRegistrationCredentials)

	mux.Get("/api/userinfo", func(w http.ResponseWriter, r *http.Request) {
		user, err := ab.LoadCurrentUser(&r)
		if err != nil {
			render.JSON(w, 401, err.Error())
			return
		}
		RefreshToken(w, r, user)
	})

	mux.Get("/api/login", challengeCode)
	mux.Get("/api/callback", callbackToken)
	mux.Get("/api/consent", acceptConsent)

	mux.Get("/api/users/email/{email}", func(w http.ResponseWriter, r *http.Request) {
		user, err := getAuthbossUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}

		render.JSON(w, 200, user)
	})

	mux.Get("/api/loginchallenge", func(w http.ResponseWriter, r *http.Request) {

		hydraConfig, _ := config.GetHydraConfig()
		oauthClient = InitClient(hydraConfig.ClientID, hydraConfig.ClientSecret)
		redirectURL := oauthClient.AuthCodeURL("state123")

		render.JSON(w, 200, map[string]string{"redirectURL": redirectURL})
	})

	mux.Get("/api/token/refresh/{email}", func(w http.ResponseWriter, r *http.Request) {
		email := chi.URLParam(r, "email")
		user, err := ab.Config.Storage.Server.Load(r.Context(), email)
		if err != nil {
			render.JSON(w, 500, map[string]string{"error": "user not found"})
			return
		}

		RefreshToken(w, r, user)
	})

	mux.Group(func(mux chi.Router) {
		mux.Use(authboss.ModuleListMiddleware(ab))
		mux.Mount("/api", http.StripPrefix("/api", ab.Config.Core.Router))
	})

	r.Any("/*resources", gin.WrapH(mux))

	// ab.Events.Before(authboss.EventGetUserSession, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
	// 	user, err := getUserFromHydraSession(w, r)
	// 	if err != nil {
	// 		return true, err
	// 	}

	// 	ab.Config.Storage.Server.Save(r.Context(), user)

	// 	return true, nil
	// })

	ab.Events.After(authboss.EventRegister, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		challenge := r.Header.Get("Challenge")
		return handleLogin(challenge, w, r)
	})

	ab.Events.After(authboss.EventAuth, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		challenge := r.Header.Get("Challenge")
		return handleLogin(challenge, w, r)
	})
}

func handleLogin(challenge string, w http.ResponseWriter, r *http.Request) (bool, error) {
	if challenge == "" {
		render.JSON(w, 422, &ResponseError{
			Status:  "error",
			Success: false,
			Error:   "no challenge code has been provided",
		})

		return true, nil
	}

	user, err := ab.LoadCurrentUser(&r)
	if user != nil && err == nil {

		user := user.(*users.User)
		val := user.GetArbitrary()
		fmt.Println(val)
		resp, errConfirm := hydra.ConfirmLogin(user.ID, false, challenge)

		if errConfirm != nil || resp.RedirectTo == "" {
			logrus.Debugf("probably challenge has been expired")
			render.JSON(w, 422, &ResponseError{
				Status:  "error",
				Success: false,
				Error:   "challenge code has been expired",
			})
			return true, nil
		}

		oauth2_auth_csrf, _ := r.Cookie(cookieAuthenticationCSRFName)
		cookieArray := []*http.Cookie{}
		resty.DefaultClient.Cookies = cookieArray

		res, err := resty.SetCookie(oauth2_auth_csrf).
			R().
			SetHeader("Accept", "application/json").
			Get(resp.RedirectTo)

		if err != nil {
			render.JSON(w, 422, &ResponseError{
				Status:  "error",
				Success: false,
				Error:   "no csrf token has been provided",
			})
			return true, nil
		}

		accessToken := res.RawResponse.Header.Get("access_token")
		if accessToken == "" {
			render.JSON(w, 422, &ResponseError{
				Status:  "error",
				Success: false,
				Error:   "No access token has been obtained",
			})
			return true, nil
		}

		SetAccessTokenCookie(w, accessToken)

		render.JSON(w, 200, map[string]string{
			"access_token": accessToken,
			"token_type":   "bearer",
		})
	}
	return true, nil
}

func getAuthbossUser(r *http.Request) (authboss.User, error) {
	email := chi.URLParam(r, "email")
	user, err := ab.Config.Storage.Server.Load(r.Context(), email)
	return user, err
}

func getAuthbossUserByEmail(r *http.Request, email string) (authboss.User, error) {
	user, err := ab.Config.Storage.Server.Load(r.Context(), email)
	return user, err
}

func formatToken(token string) string {
	token = strings.Replace(token, "Authorization=", "", 1)
	token = strings.Replace(token, "; Path=/", "", 1)
	return fmt.Sprintf("Bearer %s", token)
}
