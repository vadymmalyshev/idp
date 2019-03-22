package auth

import (
	"encoding/base32"
	"encoding/base64"
	"flag"
	"net/http"
	"regexp"

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
	"github.com/volatiletech/authboss-renderer"
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
)

var (
	ab *authboss.Authboss

	flagAPI = flag.Bool("api", false, "configure the app to be an api instead of an html app")
)

const (
	recoverSentURL = "/recover/sent"
	recoverSentTPL = "recover_sent"

	tplPath = "views/"
)

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

	serverConfig, _ := config.GetServerConfig()

	ab.Config.Paths.RootURL = serverConfig.Addr
	ab.Config.Paths.Mount = "/"

	ab.Config.Storage.Server = users.NewUserStorer(db)
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
		ab.Config.Core.ViewRenderer = abrenderer.NewHTML("/", tplPath)
	}

	ab.Config.Modules.RegisterPreserveFields = []string{"email", "username"}

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
		FieldName: "name", Required: true,
		MinLength: 2,
	}

	ab.Config.Core.BodyReader = defaults.HTTPBodyReader{
		ReadJSON: *flagAPI,
		Rulesets: map[string][]defaults.Rules{
			"register":      {emailRule, passwordRule, nameRule},
			"recover_start": {emailRule},
			"recover_end":   {passwordRule},
		},
		Confirms: map[string][]string{
			// 	"register":    {"password", authboss.ConfirmPrefix + "password"},
			"recover_end": {"password", authboss.ConfirmPrefix + "password"},
		},
		Whitelist: map[string][]string{
			"register": []string{"email", "name", "password"},
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

	schemaDec := schema.NewDecoder()
	schemaDec.IgnoreUnknownKeys(true)

	mux := chi.NewRouter()

	mux.Use(nosurfing, ab.LoadClientStateMiddleware, dataInjector)
	mux.Use(challengeCode)
	mux.Use(acceptConsent)
	mux.Use(callbackToken)

	mux.Group(func(mux chi.Router) {
		mux.Use(authboss.ModuleListMiddleware(ab))
		mux.Mount("/", http.StripPrefix("", ab.Config.Core.Router))
	})

	render := renderPkg.New()
	mux.Get("/api/users/email/{email}", func(w http.ResponseWriter, r *http.Request) {
		email := chi.URLParam(r, "email")
		user, err := ab.Config.Storage.Server.Load(r.Context(), email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}

		render.JSON(w, 200, user)
	})

	r.Any("/*resources", gin.WrapH(mux))

	ab.Events.After(authboss.EventAuth, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		challenge, _ := authboss.GetSession(r, "Challenge")

		if len(challenge) == 0 {
			ro := authboss.RedirectOptions{
				Code:         http.StatusTemporaryRedirect,
				RedirectPath: "/",
				Success:      "You have no login challenge",
			}
			ab.Core.Redirector.Redirect(w, r, ro)
			return true, nil
		}

		user, err := ab.LoadCurrentUser(&r)
		if user != nil && err == nil {
			user := user.(*users.User)

			resp, errConfirm := hydra.ConfirmLogin(user.ID, false, challenge)

			if errConfirm != nil {
				logrus.WithFields(logrus.Fields{
					"Email":     user.Email,
					"UserID":    user.ID,
					"Challenge": challenge,
				}).Error("hydra/login/accept request has been failed")
			}

			ro := authboss.RedirectOptions{
				Code:         http.StatusTemporaryRedirect,
				RedirectPath: resp.RedirectTo,
				Success:      "Hydra redirect",
			}
			logrus.Infof("user will be redirected to %s", resp.RedirectTo)
			ab.Core.Redirector.Redirect(w, r, ro)
		}
		return true, nil
	})
}

