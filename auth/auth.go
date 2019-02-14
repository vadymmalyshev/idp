package auth

import (
	"encoding/base32"
	"flag"

	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/models/users"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/volatiletech/authboss"
	"github.com/volatiletech/authboss/auth"
	"github.com/volatiletech/authboss/defaults"

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

	cookieStore := clientState.NewCookieStorer(cookieAuthenticationKey, cookieEncryptionKey)
	cookieStore.Domain = config.ServerHost
	cookieStore.MaxAge = SessionCookieMaxAge
	cookieStore.HTTPOnly = SessionCookieHTTPOnly
	cookieStore.Secure = SessionCookieSecure

	sessionStore := clientState.NewSessionStorer(IDPSessionName, cookieAuthenticationKey, cookieEncryptionKey)

	ab = authboss.New()

	ab.Config.Paths.Mount = "/"
	ab.Config.Paths.RootURL = config.ServerHost

	ab.Config.Storage.Server = users.NewUserStorer()
	ab.Config.Storage.SessionState = sessionStore
	ab.Config.Storage.CookieState = cookieStore

	if *flagAPI {
		ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
	} else {
		ab.Config.Core.ViewRenderer = abrenderer.NewHTML("/", "views/auth")
	}

	defaults.SetCore(&ab.Config, *flagAPI, false)

	modAuth := auth.Auth{}
	err := modAuth.Init(ab)

	if err != nil {
		log.Panicf("can't initialize authboss's auth mod", err)
	}

	// modRegister := register.Register{}
	// if err := modRegister.Init(ab); err != nil {
	// 	panic(err)
	// }

	// modLogout := logout.Logout{}
	// if err := modLogout.Init(ab); err != nil {
	// 	panic(err)
	// }

}
