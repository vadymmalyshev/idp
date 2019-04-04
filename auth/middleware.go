package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/internal/hydra"
	"git.tor.ph/hiveon/idp/models/users"
	"github.com/davecgh/go-spew/spew"
	"github.com/justinas/nosurf"
	. "github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/authboss"
	"golang.org/x/oauth2"
	"gopkg.in/resty.v1"
	"io/ioutil"
	"net/http"
)

var oauthClient *oauth2.Config

func ServeHTTP (w http.ResponseWriter, req *http.Request) {

}

func acceptPost(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer h.ServeHTTP(w, r)

		if r.URL.Path == "/api/login" && r.Method == "POST" && *flagAPI{
			//returnURL := getReturnURL(r, w)
			//authboss.PutSession(w, "fromURL", getReturnURL(r, w))
		}
	})
}

func acceptConsent(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer h.ServeHTTP(w, r)

		if r.URL.Path == "/api/consent" && r.Method == "GET" {
			challenge := r.URL.Query().Get("consent_challenge")

			if len(challenge) == 0 {
				ro := authboss.RedirectOptions{
					Code:         http.StatusTemporaryRedirect,
					RedirectPath: "/",
					Failure:      "You have no consent challenge",
				}
				ab.Core.Redirector.Redirect(w, r, ro)
				return
			}

			url, err := hydra.AcceptConsentChallengeCode(challenge)

			if err != nil {
				if *flagAPI {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNoContent)
					fmt.Fprintf(w, `{"consent challenge code isn't right"}`)

				} else {
					ro := authboss.RedirectOptions{
						Code:         http.StatusTemporaryRedirect,
						RedirectPath: "/",
						Failure:      "consent challenge code isn't right",
					}
					ab.Core.Redirector.Redirect(w, r, ro)
					}
				return
			}

			if *flagAPI {
				setRedirectURL(url, w)
				return
			}

			ro := authboss.RedirectOptions{
				Code:         http.StatusTemporaryRedirect,
				RedirectPath: url,
				Success:      "Consent code accepted",
			}
			ab.Core.Redirector.Redirect(w, r, ro)
			return
		}
	})
}

func challengeCode(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer h.ServeHTTP(w, r)

		if r.URL.Path == "/api/login" && r.Method == "GET" {
			challenge := r.URL.Query().Get("login_challenge")
			if len(challenge) == 0 { // obtain login challenge
				if *flagAPI {
					return
				}

				hydraConfig,_ := config.GetHydraConfig()
				oauthClient = InitClient(hydraConfig.ClientID, hydraConfig.ClientSecret)
				redirectUrl := oauthClient.AuthCodeURL("state123")

				ro := authboss.RedirectOptions{
					Code:         http.StatusTemporaryRedirect,
					RedirectPath: redirectUrl,
					Success:      "Obtaining login challenge",
				}
				ab.Core.Redirector.Redirect(w, r, ro)
				return
			}

			challengeResp, err := hydra.CheckChallengeCode(challenge)

			if err != nil {
				if *flagAPI {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNoContent)
					fmt.Fprintf(w, `{"You have wrong login challenge"}`)

				} else {
				ro := authboss.RedirectOptions{
					Code:         http.StatusTemporaryRedirect,
					RedirectPath: "/",
					Success:      "You have wrong login challenge",
				}
				ab.Core.Redirector.Redirect(w, r, ro)}
				return
			}

			challengeCode := challengeResp.Challenge
			authboss.PutSession(w, "Challenge", challengeCode)

			if *flagAPI {
				user, err := ab.LoadCurrentUser(&r)
				if user != nil && err == nil {
					user := user.(*users.User)

					resp, errConfirm := hydra.ConfirmLogin(user.ID, false, challenge)

					if errConfirm != nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNoContent)
						fmt.Fprintf(w, `{"hydra/login/accept request has been failed"}`)

						logrus.WithFields(logrus.Fields{
							"Email":     user.Email,
							"UserID":    user.ID,
							"Challenge": challenge,
						}).Error("hydra/login/accept request has been failed")
						return
					}
					setRedirectURL(resp.RedirectTo, w)
					return

				}
			}
		}
	})
}

func callbackToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer h.ServeHTTP(w, r)

		if r.URL.Path == "/api/callback" && r.Method == "GET" {
			code := r.URL.Query().Get("code")
			token, err := oauthClient.Exchange(oauth2.NoContext, code)

			if err != nil {
				if *flagAPI {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"Can't obtain authorization token"}`)

				} else {
					ro := authboss.RedirectOptions{
						Code:         http.StatusInternalServerError,
						RedirectPath: "/",
						Failure:      "Can't obtain authorization token",
					}
					ab.Core.Redirector.Redirect(w, r, ro)
				}
				return
			}

			user, err := ab.LoadCurrentUser(&r)

			if user != nil && err == nil {
				user1 := user.(*users.User)
				user1.PutOAuth2AccessToken(token.AccessToken)
				user1.PutOAuth2RefreshToken(token.RefreshToken)
				user1.PutOAuth2Expiry(token.Expiry)

				ab.Config.Storage.Server.Save(r.Context(),user1)
				}

			if err != nil {
				http.Error(w, err.Error(), http.StatusNoContent)
				return
			}

			c := http.Cookie{
				Name: "Authorization",
				Value: token.AccessToken,
				Domain: "localhost",
				Path:     "/",
			}
			http.SetCookie(w, &c)

			portalConfig, err := config.GetPortalConfig()
			if err != nil {
				http.Error(w, err.Error(), http.StatusNoContent)
				return
			}

			if *flagAPI {
				fromURL, _ := authboss.GetSession(r, "fromURL")

				if fromURL =="" {
					fromURL = portalConfig.Callback
				}

				setRedirectURL(fromURL, w)
				return
			}

			http.Redirect(w, r, portalConfig.Callback,http.StatusPermanentRedirect)
			return
		}
	})
}

//nosurfing is a more verbose wrapper around csrf handling
func nosurfing(h http.Handler) http.Handler {
	surfing := nosurf.New(h)
	surfing.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Println("Failed to validate CSRF token:", nosurf.Reason(r))
		w.WriteHeader(http.StatusBadRequest)
	}))
	return surfing
}

func dataInjector(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := layoutData(w, &r)
		r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))
		handler.ServeHTTP(w, r)
	})
}

// layoutData is passing pointers to pointers be able to edit the current pointer
// to the request. This is still safe as it still creates a new request and doesn't
// modify the old one, it just modifies what we're pointing to in our methods so
// we're able to skip returning an *http.Request everywhere
func layoutData(w http.ResponseWriter, r **http.Request) authboss.HTMLData {
	currentUserName := ""
	userInter, err := ab.LoadCurrentUser(r)
	if userInter != nil && err == nil {
		currentUserName = userInter.(*users.User).Username
	}

	return authboss.HTMLData{
		"loggedin":          userInter != nil,
		"current_user_name": currentUserName,
		//"csrf_token":        nosurf.Token(*r),
		"flash_success":     authboss.FlashSuccess(w, *r),
		"flash_error":       authboss.FlashError(w, *r),
	}
}

func debugMw(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("\n%s %s %s\n", r.Method, r.URL.Path, r.Proto)

		session, err := sessionStore.Get(r, IDPSessionName)
		if err == nil {
			fmt.Print("Session: ")
			first := true
			for k, v := range session.Values {
				if first {
					first = false
				} else {
					fmt.Print(", ")
				}
				fmt.Printf("%s = %v", k, v)
			}
			fmt.Println()
		}
		// fmt.Println("Database:")
		// for _, u := range database.Users {
		// 	fmt.Printf("! %#v\n", u)
		// }
		if val := r.Context().Value(authboss.CTXKeyData); val != nil {
			fmt.Printf("CTX Data: %s", spew.Sdump(val))
		}
		if val := r.Context().Value(authboss.CTXKeyValues); val != nil {
			fmt.Printf("CTX Values: %s", spew.Sdump(val))
		}

		h.ServeHTTP(w, r)
	})
}

func InitClient(clientId string, secret string) *oauth2.Config {

	hydraConfig,_ := config.GetHydraConfig()
	hydraAPI := hydraConfig.API
	client := GetClient(clientId)
	oauthConfig := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  hydraAPI + "/oauth2/auth",
			TokenURL: hydraAPI + "/oauth2/token",
		},
		RedirectURL: client.RedirectUris[0],
		Scopes:      getScopes(),
	}

	return oauthConfig
}

func GetClient(clientId string) OAuth2Client {
	hydraConfig,_ := config.GetHydraConfig()
	hydraAdmin := hydraConfig.Admin

	clientUrl := hydraAdmin + "/clients/" + clientId
	res, err := resty.R().Get(clientUrl)
	if err != nil {
		log.Info(err)
	}
	var client OAuth2Client
	json.Unmarshal(res.Body(), &client)
	return client

}

func getScopes() []string {
	return []string{"openid", "offline"}
}

func getReturnURL(r *http.Request, w http.ResponseWriter) string {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	f := map[string]interface{}{}

	err = json.Unmarshal([]byte(reqBody), &f)
	if err != nil {
		return ""
	}
	fromURL := f["fromURL"]
	fromURLString := ""

	if fromURL != nil {
		fromURLString = fromURL.(string)
	}

	return fromURLString
}

func setRedirectURL(redirectURL string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"redirectURL": %q}`, redirectURL)
	w.WriteHeader(http.StatusOK)
}
