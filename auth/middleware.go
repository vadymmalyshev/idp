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
	"net/http"
)

var oauthClient *oauth2.Config

func ServeHTTP (w http.ResponseWriter, req *http.Request) {

}
func acceptConsent(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer h.ServeHTTP(w, r)

		if r.URL.Path == "/consent" && r.Method == "GET" {
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
				ro := authboss.RedirectOptions{
					Code:         http.StatusTemporaryRedirect,
					RedirectPath: "/",
					Failure:      "consent challenge code isn't right",
				}
				ab.Core.Redirector.Redirect(w, r, ro)
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

		chal, _ := authboss.GetSession(r, "Challenge")
		logrus.Info("current challenge:" + chal)

		if r.URL.Path == "/login" && r.Method == "GET" {
			challenge := r.URL.Query().Get("login_challenge")

			if len(challenge) == 0 { // obtain login challenge
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
				ro := authboss.RedirectOptions{
					Code:         http.StatusTemporaryRedirect,
					RedirectPath: "/",
					Success:      "You have wrong login challenge",
				}
				ab.Core.Redirector.Redirect(w, r, ro)
				return
			}

			challengeCode := challengeResp.Challenge

			authboss.PutSession(w, "Challenge", challengeCode)
		}
	})
}

func callbackToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer h.ServeHTTP(w, r)

		if r.URL.Path == "/callback" && r.Method == "GET" {
			code := r.URL.Query().Get("code")
			token, err := oauthClient.Exchange(oauth2.NoContext, code)

			if err != nil {
				ro := authboss.RedirectOptions{
					Code:         http.StatusInternalServerError,
					RedirectPath: "/",
					Failure:      "Can't obtain authorization token",
				}
				ab.Core.Redirector.Redirect(w, r, ro)
				return
			}
			var introToken OAuth2TokenIntrospection
			hydraConfig,_ := config.GetHydraConfig()
			introspectUrl := hydraConfig.IntrospectURL

			res, err := resty.R().SetFormData(map[string]string{"token": token.AccessToken}).
				SetHeader("Content-Type", "application/x-www-form-urlencoded").
				SetHeader("Accept", "application/json").Post(introspectUrl)

			err = json.Unmarshal(res.Body(), &introToken)
			user, err := ab.Config.Storage.Server.Load(r.Context(), introToken.Sub)

			user1 := user.(*users.User)
			user1.PutOAuth2AccessToken(token.AccessToken)
			user1.PutOAuth2RefreshToken(token.RefreshToken)
			ab.Config.Storage.Server.Save(r.Context(),user1)

			if err != nil {
				http.Error(w, err.Error(), http.StatusNoContent)
				return
			}

			c := http.Cookie{
				Name: "Authorization",
				Value: token.AccessToken,
				Domain: "hiveon.local",
				Path:     "/",
			}

			http.SetCookie(w, &c)
			http.Redirect(w, r, "/refresh1",http.StatusPermanentRedirect)
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
		"csrf_token":        nosurf.Token(*r),
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