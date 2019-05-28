package auth

import (
	"git.tor.ph/hiveon/idp/models/users"
	"github.com/sirupsen/logrus"
	"net/http"
)

type LogType string

func (l LogType) String() string {
	return string(l)
}

var (
	LogTypeLogin LogType = "login"
	LogTypeRegistration LogType = "registration"
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

	err = a.createLoginRecord(w, r, LogTypeRegistration)
	if err != nil {
		return true, err
	}

	return a.handleLogin(challenge, w, r)
}

func (a Auth) AfterEventLogin(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
	err := a.createLoginRecord(w, r, LogTypeLogin)
	return true, err
}

func (a Auth) createLoginRecord(w http.ResponseWriter, r *http.Request, loginType LogType) error {
	abUser, err := a.authBoss.LoadCurrentUser(&r)

	if err != nil {
		return err;
	}
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
	newLog.Type = loginType.String()

	return a.userLogger.CreateRecord(newLog)
}
