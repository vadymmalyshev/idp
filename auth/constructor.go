package auth

import (
	renderPkg "github.com/unrolled/render"

	"net/http"

	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/models/users"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
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

	//Register authBoss recover post request
	//a.authBoss.Core.Router.Post(recoverSentURL, http.HandlerFunc(a.loginChallenge))
	a.authBoss.Config.Core.Router.Get(recoverSentURL, a.authBoss.Core.ErrorHandler.Wrap(func(w http.ResponseWriter, req *http.Request) error {
		challenge, err := a.getChallengeCodeFromHydra(req)
		if err != nil {
			logrus.Error("can't get challenge code after register", err)
			return err
		}
		_ ,err = a.handleLogin(challenge, w, req)

		if err != nil {
			logrus.Error("can't login", err)
			return err
		}
		return nil
	}))

	a.authBoss.Events.After(authboss.EventRegister, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		referalID, err := r.Cookie("refId")

		if err == nil && referalID != nil {
			abUser, err := a.authBoss.LoadCurrentUser(&r)
			if abUser != nil && err == nil {
				user := abUser.(*users.User)
				user.PutReferaL(referalID.Value)
			}
		}

		challenge, err := a.getChallengeCodeFromHydra(r)
		if err != nil {
			logrus.Error("can't get challenge code after register", err)
			return true, err
		}

		return a.handleLogin(challenge, w, r)
	})

	a.authBoss.Events.After(authboss.EventAuth, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		//TODO move challenge to back after front fix
		/*challenge, err := a.getChallengeCodeFromHydra(r)
		if err != nil {
			logrus.Error("can't get challenge code after register", err)
			return true, err
		}*/

		challenge := r.Header.Get("Challenge")

		return a.handleLogin(challenge, w, r)
	})

	mux := chi.NewRouter()

	mux.Use(a.authBoss.LoadClientStateMiddleware, remember.Middleware(a.authBoss))
	mux.Use(a.handleUserSession)
	mux.Use(a.checkRegistrationCredentials)
	mux.Use(a.dataInjector)

	mux.Get("/api/userinfo", a.getUserInfo)
	//TODO remove after line will :68 fixed
	mux.Get("/api/login", a.challengeCode)
	mux.Get("/api/callback", a.callbackToken)
	mux.Get("/api/consent", a.acceptConsent)
	mux.Get("/api/users/email/{email}", a.getUserByEmail)
	mux.Get("/api/loginchallenge", a.loginChallenge)
	mux.Get("/api/token/refresh/{email}", a.refreshTokenByEmail)

	mux.Group(func(mux chi.Router) {
		mux.Use(authboss.ModuleListMiddleware(a.authBoss))
		mux.Mount("/api", http.StripPrefix("/api", a.authBoss.Config.Core.Router))
	})

	a.r.Any("/*resources", gin.WrapH(mux))
}
