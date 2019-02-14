package auth

import (
	"encoding/base32"
	"flag"
	"net/http"
	"regexp"

	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/models/users"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/volatiletech/authboss"
	"github.com/volatiletech/authboss/auth"
	"github.com/volatiletech/authboss/defaults"
	"github.com/volatiletech/authboss/register"

	clientState "github.com/volatiletech/authboss-clientstate"
	"github.com/volatiletech/authboss-renderer"
)

const IDPSessionName = "_idp_session"

// SessionCookieMaxAge holds long an authenticated session should be valid in seconds
const SessionCookieMaxAge = 30 * 24 * 60 * 60

// SessionCookieHTTPOnly describes if the cookies should be accessible from HTTP requests only (no JS)
const SessionCookieHTTPOnly = true

// SessionCookieName is the name of the token that is stored in the session cookie
const SessionCookieName = "Hiveon ID Session Token"

const SessionCookieSecure = false

var (
	CookieDomain  string
	SessionCookie bool

	signingKey       string
	signingKeyBase32 string

	cookieStore clientState.CookieStorer
	sessionStore clientState.SessionStorer
)

var (
	log *logrus.Logger

	ab *authboss.Authboss

	flagAPI = flag.Bool("api", false, "configure the app to be an api instead of an html app")
)

func init() {
	log = config.Logger()

}
func Init(r *gin.Engine) {
	signingKey := config.AuthSignKey

	signingKeyBytes := []byte(signingKey)
	signingKeyBase32 = base32.StdEncoding.EncodeToString(signingKeyBytes)

	cookieAuthenticationKey := signingKeyBytes
	cookieEncryptionKey := signingKeyBytes[:32]

	cookieStore = clientState.NewCookieStorer(cookieAuthenticationKey, cookieEncryptionKey)
	cookieStore.Domain = config.ServerHost
	cookieStore.MaxAge = SessionCookieMaxAge
	cookieStore.HTTPOnly = SessionCookieHTTPOnly
	cookieStore.Secure = SessionCookieSecure

	sessionStore = clientState.NewSessionStorer(IDPSessionName, cookieAuthenticationKey, cookieEncryptionKey)

	ab = authboss.New()

	ab.Config.Paths.Mount = "/"
	ab.Config.Paths.RootURL = config.ServerHost

	ab.Config.Storage.Server = users.NewUserStorer()
	ab.Config.Storage.SessionState = sessionStore
	ab.Config.Storage.CookieState = cookieStore

	if !*flagAPI {
		// Prevent us from having to use Javascript in our basic HTML
		// to create a delete method, but don't override this default for the API
		// version
		ab.Config.Modules.LogoutMethod = "GET"
	}

	if *flagAPI {
		ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
	} else {
		ab.Config.Core.ViewRenderer = abrenderer.NewHTML("/", "views/auth")
	}

	ab.Config.Core.MailRenderer = abrenderer.NewEmail("/", "views/auth")
	ab.Config.Core.Router = defaults.NewRouter()

	ab.Config.Modules.RegisterPreserveFields = []string{"email", "username"}

	ab.Config.Modules.TOTP2FAIssuer = "HiveonID"
	ab.Config.Modules.TwoFactorEmailAuthRequired = false

	defaults.SetCore(&ab.Config, *flagAPI, false)

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
		FieldName: "name", Required: true,
		MinLength: 2,
	}

	ab.Config.Core.BodyReader = defaults.HTTPBodyReader{
		ReadJSON: *flagAPI,
		Rulesets: map[string][]defaults.Rules{
			"register":    {emailRule, passwordRule, nameRule},
			"recover_end": {passwordRule},
		},
		// Confirms: map[string][]string{
		// 	"register":    {"password", authboss.ConfirmPrefix + "password"},
		// 	"recover_end": {"password", authboss.ConfirmPrefix + "password"},
		// },
		Whitelist: map[string][]string{
			"register": []string{"email", "name", "password"},
		},
	}

	modAuth := auth.Auth{}
	if err := modAuth.Init(ab); err != nil {
		log.Panicf("can't initialize authboss's auth mod", err)
	}

	modRegister := register.Register{}
	if err := modRegister.Init(ab); err != nil {
		log.Panicf("can't initialize authboss's register mod", err)
	}

	schemaDec := schema.NewDecoder()
	schemaDec.IgnoreUnknownKeys(true)

	mux := chi.NewRouter()

	mux.Use(challengeCode)
	mux.Use(nosurfing, ab.LoadClientStateMiddleware, dataInjector)
	mux.Group(func(mux chi.Router) {
		mux.Use(authboss.ModuleListMiddleware(ab))
		mux.Mount("/", http.StripPrefix("", ab.Config.Core.Router))
	})

	r.Any("/*resources", gin.WrapH(mux))

	ab.Events.After(authboss.EventAuthHijack, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		beforeHasValues := r.Context().Value(authboss.CTXKeyValues) != nil
		return beforeHasValues, nil
	})
}
