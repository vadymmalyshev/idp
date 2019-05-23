package auth

import (
	"net/http"

	"git.tor.ph/hiveon/idp/models/users"
	"github.com/sirupsen/logrus"
)

func (a Auth) AfterEventRegistration(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
	referalID, err := r.Cookie("refId")

	if err == nil && referalID != nil {
		abUser, err := a.authBoss.LoadCurrentUser(&r)
		if abUser != nil && err == nil {
			user := abUser.(*users.User)
			user.PutReferal(referalID.Value)
		}
	}
	challenge, cookie, err := a.getChallengeCodeFromHydra(r)

	if err != nil {
		logrus.Error("can't get challenge code after register", err)
		return true, err
	}
	http.SetCookie(w, cookie)

	return a.handleLogin(challenge, w, r)
}
