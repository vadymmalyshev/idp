package auth

import (
	"context"
	"net/http"
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"git.tor.ph/hiveon/idp/internal/hydra"
	"git.tor.ph/hiveon/idp/models/users"
	"github.com/justinas/nosurf"
	"github.com/volatiletech/authboss"
)

func challengeCode(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chal, _ := authboss.GetSession(r, "Challenge")
		log.Info("chal:" + chal)

		if r.URL.Path == "/login" && r.Method == "GET" {

			challenge := r.URL.Query().Get("login_challenge")

			if len(challenge) == 0 {
				ro := authboss.RedirectOptions{
					Code:         http.StatusTemporaryRedirect,
					RedirectPath: "/",
					Success:      "You have no login challenge",
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

		writ, _ := authboss.GetSession(r, "Challenge")
		log.Info("writ:" + writ)

		log.Info("/login mw skipped")
	})
}

//nosurfing is a more verbose wrapper around csrf handling
func nosurfing(h http.Handler) http.Handler {
	surfing := nosurf.New(h)
	surfing.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Failed to validate CSRF token:", nosurf.Reason(r))
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
