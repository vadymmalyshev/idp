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
	a.authBoss = initAuthBoss(a.conf.ServerConfig.Addr, a.db, sessionStore, cookieStore)

	a.authBoss.Events.After(authboss.EventRegister, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		challenge := r.Header.Get("Challenge")
		return a.handleLogin(challenge, w, r)
	})

	a.authBoss.Events.After(authboss.EventAuth, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		challenge := r.Header.Get("Challenge")
		return a.handleLogin(challenge, w, r)
	})

	mux := chi.NewRouter()

	mux.Use(a.authBoss.LoadClientStateMiddleware, remember.Middleware(a.authBoss))
	mux.Use(a.handleUserSession)
	mux.Use(a.dataInjector)

	mux.Get("/api/userinfo", a.getUserInfo)
	mux.Get("/api/login", a.challengeCode)
	mux.Get("/api/callback", a.callbackToken)
	mux.Get("/api/consent", a.acceptConsent)
	mux.Get("/api/users/email/{email}", a.getUserByEmail)
	mux.Get("/api/loginchallenge", a.loginChallenge)
	mux.Get("/api/token/refresh/{email}", a.refreshTokenByEmail)

	mux.Post("/api/register", a.checkRegistrationCredentials)

	mux.Group(func(mux chi.Router) {
		mux.Use(authboss.ModuleListMiddleware(a.authBoss))
		mux.Mount("/api", http.StripPrefix("/api", a.authBoss.Config.Core.Router))
	})

	a.r.Any("/*resources", gin.WrapH(mux))
}
