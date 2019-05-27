package auth

import (
	"git.tor.ph/hiveon/idp/models/users"
	"github.com/sirupsen/logrus"
	"net/http"
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

	err = a.createLoginRecord(w, r, "R")
	if err != nil {
		return true, err
	}

	return a.handleLogin(challenge, w, r)
}

func (a Auth) AfterEventLogin(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
	err := a.createLoginRecord(w, r, "L")
	return true, err
}

func (a Auth) createLoginRecord(w http.ResponseWriter, r *http.Request, loginType string) error {
	abUser, err := a.authBoss.LoadCurrentUser(&r)
	if abUser != nil && err == nil {
		newLog := a.userLogger.New()
		IP := r.RemoteAddr;
		ua := r.Header.Get("User-Agent")
		fromUrl, err := r.Cookie("fromUrl")
		fu := "Unknown"
		if err == nil {
			fu = fromUrl.Domain
		}
		newLog.PutIP(IP)
		newLog.PutAgent(ua)
		newLog.PutDomen(fu)
		newLog.PutUserID(abUser.GetPID())
		newLog.Type = loginType

		err = a.userLogger.CreateRecord(newLog)
	}
	return err
}