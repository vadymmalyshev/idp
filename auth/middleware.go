package auth

import (
	"bytes"
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
		if r.URL.Path == "/api/login" && r.Method == "POST" && *flagAPI {
			fromURL, challenge := getChallengeFromURL(r, w)
			authboss.PutSession(w, "Challenge", challenge)
			authboss.PutSession(w, "fromURL", fromURL)
			//k, _ := r.Cookie("oauth2_authentication_csrf");
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
				/*data := layoutData(w, &r, url)
				r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))
				h.ServeHTTP(w, r)*/
				setRedirectURL(url, w)
				h.ServeHTTP(w, r)
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
			challenge := r.URL.Query().Get("Challenge")
			if len(challenge) == 0 { // obtain login challenge

				hydraConfig,_ := config.GetHydraConfig()
				oauthClient = InitClient(hydraConfig.ClientID, hydraConfig.ClientSecret)
				redirectUrl := oauthClient.AuthCodeURL("state123")
				// return link to hydra
				data := layoutData(w, &r, redirectUrl)
				r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))
				h.ServeHTTP(w, r)

				if !*flagAPI {
					ro := authboss.RedirectOptions{
						Code:         http.StatusTemporaryRedirect,
						RedirectPath: redirectUrl,
						Success:      "Obtaining login challenge",
					}
					ab.Core.Redirector.Redirect(w, r, ro)
				}

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

			if !*flagAPI {
				authboss.PutSession(w, "Challenge", challengeCode)
			}

			// put login_challenge in cookies
			if *flagAPI {
				c2 := http.Cookie{
					Name:  "Challenge",
					Value: challengeCode,
					//Domain: "localhost",
					Path: "/",
				}
				http.SetCookie(w, &c2)
				h.ServeHTTP(w, r)
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
				/*
				data := layoutData(w, &r, fromURL)
				r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))*/
				setRedirectURL(fromURL, w)
				h.ServeHTTP(w, r)

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
		data := layoutData(w, &r, "")
		r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))
		handler.ServeHTTP(w, r)
	})
}

// layoutData is passing pointers to pointers be able to edit the current pointer
// to the request. This is still safe as it still creates a new request and doesn't
// modify the old one, it just modifies what we're pointing to in our methods so
// we're able to skip returning an *http.Request everywhere
func layoutData(w http.ResponseWriter, r **http.Request, redirect string) authboss.HTMLData {
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
		"redirectURL":       redirect,
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

func getChallengeFromURL(r *http.Request, w http.ResponseWriter) (string, string) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	b:=bodyBytes
	fromURLString := ""
	chalengeString := ""

	var t map[string]string
	json.Unmarshal(b, &t)

	if t["fromURL"] != "" {
		fromURLString = t["fromURL"]

	}
	if t["Challenge"] != "" {
		chalengeString = t["Challenge"]

	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return fromURLString, chalengeString
}

func setRedirectURL(redirectURL string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"redirectURL": %q}`, redirectURL)
	w.WriteHeader(http.StatusOK)
}
