package auth

import (
	"git.tor.ph/hiveon/idp/models/logs"
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

	//IDP handlers

	// swagger:route GET /userinfo
	//
	// Get user info and refresh auth token if needed
	//
	// responses:
	// "200":
	//    description: User info
	// "401":
	//    description: Authorization token missed
	//    "$ref": "#/responses/ResponseError"
	mux.Get(rootPath+"/userinfo", a.getUserInfo)

	// swagger:route GET /users/email/{email}
	//
	// Get user info by email
	//
	// parameters:
	// - name: email
	//   in: path
	//   description: User's email
	//   required: true
	//   type: string
	// responses:
	// "200":
	//    description: User info
	// "204":
	//    description: User not found
	//    "$ref": "#/responses/ResponseError"
	mux.Get(rootPath+"/users/email/{email}", a.getUserByEmail)

	// swagger:route GET /token/refresh/{email}
	//
	// Refresh auth token by email
	//
	// parameters:
	// - name: email
	//   in: path
	//   description: User's email
	//   required: true
	//   type: string
	// responses:
	// "200":
	//    description: access token
	// "500":
	//    description: User not found
	//    "$ref": "#/responses/ResponseError"
	// "401": ResponseError
	//    description: No refresh token
	//    "$ref": "#/responses/ResponseError"
	mux.Get(rootPath+"/token/refresh/{email}", a.refreshTokenByEmail)

	// swagger:route POST /login
	//
	// User's login
	//
	// parameters:
	// - name: email
	//   in: formData
	//   description: User's email/login
	//   required: true
	//   type: string
	// - name: password
	//   in: formData
	//   description: User's password
	//   required: true
	//   type: string
	// - name: code
	//   in: formData
	//   description: Promocode
	//   required: false
	//   type: string
	// - name: fromUrl
	//   in: formData
	//   description: From URL
	//   required: true
	//   type: string
	// - name: rm
	//   in: formData
	//   description: Remember me
	//   required: true
	//   type: bool
	// responses:
	// "200":
	//    description: login success
	// "400":
	//    description: TOTP error
	//    "$ref": "#/responses/ResponseError"
	// "401": ResponseError
	//    description: User failed to log in
	//    "$ref": "#/responses/ResponseError"
	// "422":
	//    description: Can't get challenge code after register
	//    "$ref": "#/responses/ResponseError"
	mux.Post(rootPath+"/login", a.loginPost)
	// Hydra endpoints
	mux.Get(rootPath+"/callback", a.callbackToken)
	mux.Get(rootPath+"/consent", a.acceptConsent)


	//AuthBoss handlers
	a.authBoss.Config.Core.Router.Get(recoverSentURL, a.authBoss.Core.ErrorHandler.Wrap(a.getRecoverSentURL))

	mux.Group(func(mux chi.Router) {
		mux.Use(authboss.ModuleListMiddleware(a.authBoss))
		mux.Mount(rootPath, http.StripPrefix(rootPath, a.authBoss.Config.Core.Router))
	})

	a.r.Any("/*resources", gin.WrapH(mux))
}

func initUserLogger(db *gorm.DB) *logs.UserLogger{
	return logs.NewUserLogger(db)
}
