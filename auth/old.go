package auth

// import (
// 	"context"
// 	"encoding/base64"
// 	"encoding/json"
// 	"flag"
// 	"fmt"

// 	"log"

// 	"git.tor.ph/hiveon/idp/config"
// 	"git.tor.ph/hiveon/idp/models/users"

// 	"github.com/go-chi/chi"

// 	"net/http"
// 	"regexp"

// 	"github.com/davecgh/go-spew/spew"
// 	"github.com/gin-gonic/gin"

// 	"github.com/volatiletech/authboss"
// 	"github.com/volatiletech/authboss/auth"
// 	"github.com/volatiletech/authboss/defaults"
// 	"github.com/volatiletech/authboss/logout"

// 	_ "github.com/volatiletech/authboss/logout"
// 	_ "github.com/volatiletech/authboss/recover"
// 	"github.com/volatiletech/authboss/register"

// 	"github.com/volatiletech/authboss-clientstate"
// 	"github.com/volatiletech/authboss-renderer"

// 	"github.com/aarondl/tpl"
// 	"github.com/gorilla/schema"
// 	"github.com/gorilla/sessions"
// 	"github.com/justinas/nosurf"

// 	"github.com/gwatts/gin-adapter"
// )

// var (
// 	flagDebug    = flag.Bool("debug", false, "output debugging information")
// 	flagDebugDB  = flag.Bool("debugdb", false, "output database on each request")
// 	flagDebugCTX = flag.Bool("debugctx", false, "output specific authboss related context keys on each request")
// 	flagAPI      = flag.Bool("api", false, "configure the app to be an api instead of an html app")
// )

// var (
// 	ab = authboss.New()

// 	schemaDec = schema.NewDecoder()

// 	sessionStore abclientstate.SessionStorer
// 	cookieStore  abclientstate.CookieStorer

// 	templates tpl.Templates
// )

// const (
// 	sessionCookieName = "hiveon_idp"
// )

// func setupAuthboss() {
// 	ab.Config.Paths.RootURL = config.ServerURL
// 	// ab.Config.Paths.Mount = "/auth"

// 	ab.Config.Storage.Server = users.NewUserStorer()
// 	ab.Config.Storage.SessionState = sessionStore
// 	ab.Config.Storage.CookieState = cookieStore

// 	if !*flagAPI {
// 		// Prevent us from having to use Javascript in our basic HTML
// 		// to create a delete method, but don't override this default for the API
// 		// version
// 		ab.Config.Modules.LogoutMethod = "GET"
// 	}

// 	if *flagAPI {
// 		ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
// 	} else {
// 		ab.Config.Core.ViewRenderer = abrenderer.NewHTML("/", "views/auth")
// 	}

// 	ab.Config.Core.MailRenderer = abrenderer.NewEmail("/", "views/auth")
// 	ab.Config.Core.Router = defaults.NewRouter()

// 	ab.Config.Modules.RegisterPreserveFields = []string{"email", "username"}

// 	ab.Config.Modules.TOTP2FAIssuer = "HiveonID"
// 	ab.Config.Modules.RoutesRedirectOnUnauthed = true

// 	ab.Config.Modules.TwoFactorEmailAuthRequired = false

// 	defaults.SetCore(&ab.Config, *flagAPI, false)

// 	emailRule := defaults.Rules{
// 		FieldName: "email", Required: true,
// 		MatchError: "Must be a valid e-mail address",
// 		MustMatch:  regexp.MustCompile(`.*@.*\.[a-z]{1,}`),
// 	}
// 	passwordRule := defaults.Rules{
// 		FieldName: "password", Required: true,
// 		MinLength: 4,
// 	}
// 	nameRule := defaults.Rules{
// 		FieldName: "name", Required: true,
// 		MinLength: 2,
// 	}

// 	ab.Config.Core.BodyReader = defaults.HTTPBodyReader{
// 		ReadJSON: *flagAPI,
// 		Rulesets: map[string][]defaults.Rules{
// 			"register":    {emailRule, passwordRule, nameRule},
// 			"recover_end": {passwordRule},
// 		},
// 		// Confirms: map[string][]string{
// 		// 	"register":    {"password", authboss.ConfirmPrefix + "password"},
// 		// 	"recover_end": {"password", authboss.ConfirmPrefix + "password"},
// 		// },
// 		Whitelist: map[string][]string{
// 			"register": []string{"email", "name", "password"},
// 		},
// 	}

// 	modAuth := auth.Auth{}
// 	if err := modAuth.Init(ab); err != nil {
// 		panic(err)
// 	}

// 	modRegister := register.Register{}
// 	if err := modRegister.Init(ab); err != nil {
// 		panic(err)
// 	}

// 	modLogout := logout.Logout{}
// 	if err := modLogout.Init(ab); err != nil {
// 		panic(err)
// 	}

// }

// func Init(r *gin.Engine) {

// 	cookieStoreKey, _ := base64.StdEncoding.DecodeString(`NpEPi8pEjKVjLGJ6kYCS+VTCzi6BUuDzU0wrwXyf5uDPArtlofn2AG6aTMiPmN3C909rsEWMNqJqhIVPGP3Exg==`)
// 	cookieStore = abclientstate.NewCookieStorer(cookieStoreKey, nil)
// 	cookieStore.HTTPOnly = false
// 	cookieStore.Secure = false

// 	sessionStoreKey, _ := base64.StdEncoding.DecodeString(`AbfYwmmt8UCwUuhd9qvfNA9UCuN1cVcKJN1ofbiky6xCyyBj20whe40rJa3Su0WOWLWcPpO1taqJdsEI/65+JA==`)
// 	sessionStore = abclientstate.NewSessionStorer(sessionCookieName, sessionStoreKey, nil)

// 	cstore := sessionStore.Store.(*sessions.CookieStore)
// 	cstore.Options.HttpOnly = false
// 	cstore.Options.Secure = false

// 	setupAuthboss()

// 	schemaDec.IgnoreUnknownKeys(true)

// 	mux := chi.NewRouter()

// 	mux.Use(nosurfing, ab.LoadClientStateMiddleware, dataInjector)
// 	mux.Group(func(mux chi.Router) {
// 		mux.Use(authboss.ModuleListMiddleware(ab))
// 		mux.Mount("/auth", http.StripPrefix("/auth", ab.Config.Core.Router))
// 	})

// 	r.Any("/auth/*resources", gin.WrapH(mux))

// 	r.GET("/current", func(c *gin.Context) {
// 		user, err := ab.CurrentUser(c.Request)
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}

// 		log.Println("User loaded:", user.GetPID())
// 	})

// 	if *flagAPI {
// 		// In order to have a "proper" API with csrf protection we allow
// 		// the options request to return the csrf token that's required to complete the request
// 		// when using post
// 		optionsHandler := func(w http.ResponseWriter, r *http.Request) {
// 			w.Header().Set("X-CSRF-TOKEN", nosurf.Token(r))
// 			w.WriteHeader(http.StatusOK)
// 		}

// 		// We have to add each of the authboss get/post routes specifically because
// 		// chi sees the 'Mount' above as overriding the '/*' pattern.
// 		routes := []string{"login", "logout", "recover", "recover/end", "register"}
// 		// r.OPTIONS("/*", gin.WrapF(optionsHandler))
// 		for _, route := range routes {
// 			r.OPTIONS("/auth/"+route, gin.WrapF(optionsHandler))
// 		}
// 	}

// 	ab.Events.Before(authboss.EventAuth, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
// 		beforeHasValues := r.Context().Value(authboss.CTXKeyValues) != nil
// 		return beforeHasValues, nil
// 	})

// }

// func dataInjector(handler http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		data := layoutData(w, &r)
// 		r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))
// 		handler.ServeHTTP(w, r)
// 	})
// }

// // layoutData is passing pointers to pointers be able to edit the current pointer
// // to the request. This is still safe as it still creates a new request and doesn't
// // modify the old one, it just modifies what we're pointing to in our methods so
// // we're able to skip returning an *http.Request everywhere
// func layoutData(w http.ResponseWriter, r **http.Request) authboss.HTMLData {
// 	currentUserName := ""
// 	userInter, err := ab.LoadCurrentUser(r)
// 	if userInter != nil && err == nil {
// 		currentUserName = userInter.(*users.User).Username
// 	}

// 	return authboss.HTMLData{
// 		"loggedin":          userInter != nil,
// 		"current_user_name": currentUserName,
// 		"csrf_token":        nosurf.Token(*r),
// 		"flash_success":     authboss.FlashSuccess(w, *r),
// 		"flash_error":       authboss.FlashError(w, *r),
// 	}
// }

// func addMW(r *gin.Engine, f func(http.Handler) http.Handler) {
// 	r.Use(adapter.Wrap(f))
// }

// func addMWtoGroup(r *gin.RouterGroup, f func(http.Handler) http.Handler) {
// 	r.Use(adapter.Wrap(f))
// }

// func mustRender(w http.ResponseWriter, r *http.Request, name string, data authboss.HTMLData) {
// 	// We've sort of hijacked the authboss mechanism for providing layout data
// 	// for our own purposes. There's nothing really wrong with this but it looks magical
// 	// so here's a comment.
// 	var current authboss.HTMLData
// 	dataIntf := r.Context().Value(authboss.CTXKeyData)
// 	if dataIntf == nil {
// 		current = authboss.HTMLData{}
// 	} else {
// 		current = dataIntf.(authboss.HTMLData)
// 	}

// 	current.MergeKV("csrf_token", nosurf.Token(r))
// 	current.Merge(data)

// 	if *flagAPI {
// 		w.Header().Set("Content-Type", "application/json")

// 		byt, err := json.Marshal(current)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			fmt.Println("failed to marshal json:", err)
// 			fmt.Fprintln(w, `{"error":"internal server error"}`)
// 		}

// 		w.Write(byt)
// 		return
// 	}

// 	err := templates.Render(w, name, current)
// 	if err == nil {
// 		return
// 	}

// 	w.Header().Set("Content-Type", "text/plain")
// 	w.WriteHeader(http.StatusInternalServerError)
// 	fmt.Fprintln(w, "Error occurred rendering template:", err)
// }

// func nosurfing(h http.Handler) http.Handler {
// 	surfing := nosurf.New(h)
// 	surfing.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Println("Failed to validate CSRF token:", nosurf.Reason(r))
// 		w.WriteHeader(http.StatusBadRequest)
// 	}))
// 	return surfing
// }

// func badRequest(w http.ResponseWriter, err error) bool {
// 	if err == nil {
// 		return false
// 	}

// 	if *flagAPI {
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusBadRequest)
// 		fmt.Fprintln(w, `{"error":"bad request"}`, err)
// 		return true
// 	}

// 	w.Header().Set("Content-Type", "text/plain")
// 	w.WriteHeader(http.StatusBadRequest)
// 	fmt.Fprintln(w, "Bad request:", err)
// 	return true
// }

// func redirect(w http.ResponseWriter, r *http.Request, path string) {
// 	if *flagAPI {
// 		w.Header().Set("Content-Type", "application/json")
// 		w.Header().Set("Location", path)
// 		w.WriteHeader(http.StatusFound)
// 		fmt.Fprintf(w, `{"path": %q}`, path)
// 		return
// 	}

// 	http.Redirect(w, r, path, http.StatusFound)
// }

// func logger(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Printf("\n%s %s %s\n", r.Method, r.URL.Path, r.Proto)

// 		if *flagDebug {
// 			session, err := sessionStore.Get(r, sessionCookieName)
// 			if err == nil {
// 				fmt.Print("Session: ")
// 				first := true
// 				for k, v := range session.Values {
// 					if first {
// 						first = false
// 					} else {
// 						fmt.Print(", ")
// 					}
// 					fmt.Printf("%s = %v", k, v)
// 				}
// 				fmt.Println()
// 			}
// 		}

// 		if *flagDebugCTX {
// 			if val := r.Context().Value(authboss.CTXKeyData); val != nil {
// 				fmt.Printf("CTX Data: %s", spew.Sdump(val))
// 			}
// 			if val := r.Context().Value(authboss.CTXKeyValues); val != nil {
// 				fmt.Printf("CTX Values: %s", spew.Sdump(val))
// 			}
// 		}

// 		h.ServeHTTP(w, r)
// 	})
// }
