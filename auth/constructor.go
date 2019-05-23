package auth

import (
	renderPkg "github.com/unrolled/render"
	"net/http"

	"git.tor.ph/hiveon/idp/config"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"github.com/volatiletech/authboss"
	"github.com/volatiletech/authboss/remember"
)

type Auth struct {
	r        *gin.Engine
	db       *gorm.DB
	conf     *config.CommonConfig
	render   *renderPkg.Render
	authBoss *authboss.Authboss
}

func NewAuth(r *gin.Engine, db *gorm.DB, conf *config.CommonConfig) *Auth {
	auth := &Auth{
		r:    r,
		db:   db,
		conf: conf,
	}
	auth.render = renderPkg.New()

	return auth
}

func (a *Auth) Init() {
	sessionStore := initSessionStorer()
	cookieStore := initCookieStorer()
	a.authBoss = initAuthBoss(a.conf.Portal.Callback, a.db, sessionStore, cookieStore)

	//Events
	a.authBoss.Events.After(authboss.EventRegister, a.AfterEventRegistration)
	//a.authBoss.Events.After(authboss.EventAuth, a.AfterEventAuth)
	//a.authBoss.Events.Before(authboss.EventAuthHijack, a.BeforeEventAuthHijack)

	mux := chi.NewRouter()

	//Middlewares
	mux.Use(a.authBoss.LoadClientStateMiddleware, remember.Middleware(a.authBoss))
	mux.Use(a.handleUserSession)
	mux.Use(a.checkRegistrationCredentials)
	mux.Use(a.check2FaSetupRequest)
	mux.Use(a.store2faCode)
	mux.Use(a.dataInjector)

	//IDP handlers
	mux.Get("/api/userinfo", a.getUserInfo)
	mux.Get("/api/callback", a.callbackToken)
	mux.Get("/api/consent", a.acceptConsent)
	mux.Get("/api/users/email/{email}", a.getUserByEmail)
	mux.Get("/api/token/refresh/{email}", a.refreshTokenByEmail)

	//AuthBoss handlers
	a.authBoss.Config.Core.Router.Get(recoverSentURL, a.authBoss.Core.ErrorHandler.Wrap(a.getRecoverSentURL))
	a.authBoss.Config.Core.Router.Post(loginPostURL, a.authBoss.Core.ErrorHandler.Wrap(a.LoginPost))

	mux.Group(func(mux chi.Router) {
		mux.Use(authboss.ModuleListMiddleware(a.authBoss))
		mux.Mount("/api", http.StripPrefix("/api", a.authBoss.Config.Core.Router))
	})

	a.r.Any("/*resources", gin.WrapH(mux))
}
