package auth

import (
	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/models/logs"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	renderPkg "github.com/unrolled/render"
	"github.com/volatiletech/authboss"
	"github.com/volatiletech/authboss/remember"
	"net/http"
)

type Auth struct {
	r          *gin.Engine
	db         *gorm.DB
	conf       *config.CommonConfig
	render     *renderPkg.Render
	authBoss   *authboss.Authboss
	userLogger *logs.UserLogger
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
	a.userLogger = initUserLogger(a.db)
	//Events
	a.authBoss.Events.After(authboss.EventRegister, a.AfterEventRegistration)
	a.authBoss.Events.After(authboss.EventAuth, a.AfterEventLogin)

	mux := chi.NewRouter()

	//Middlewares
	mux.Use(a.authBoss.LoadClientStateMiddleware, remember.Middleware(a.authBoss))
	mux.Use(a.handleUserSession)
	mux.Use(a.checkRegistrationCredentials)
	mux.Use(a.check2FaSetupRequest)
	mux.Use(a.store2faCode)
	mux.Use(a.dataInjector)
	mux.Use(a.deleteAuthorizationCookieAfterLogout)

	mux.Route(rootPath, func(r chi.Router) {
		//IDP handlers
		r.Get("/userinfo", a.getUserInfo)
		r.Get("/users/email/{email}", a.getUserByEmail)
		r.Get("/token/refresh/{email}", a.refreshTokenByEmail)
		r.Post("/login", a.loginPost)
		r.Get("/callback", a.callbackToken)
		r.Get("/consent", a.acceptConsent)

		r.Group(func(mux chi.Router) {
			mux.Use(authboss.ModuleListMiddleware(a.authBoss))
			mux.Mount("/", http.StripPrefix(rootPath, a.authBoss.Config.Core.Router))
		})
	})

	//AuthBoss handlers
	a.authBoss.Config.Core.Router.Get(recoverSentURL, a.authBoss.Core.ErrorHandler.Wrap(a.getRecoverSentURL))
	a.r.Any("/*resources", gin.WrapH(mux))
}

func initUserLogger(db *gorm.DB) *logs.UserLogger {
	return logs.NewUserLogger(db)
}
