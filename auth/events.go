package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"git.tor.ph/hiveon/idp/models/users"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

type LogType string

func (l LogType) String() string {
	return string(l)
}

var (
	LogTypeLogin        LogType = "login"
	LogTypeRegistration LogType = "registration"
)

func (a Auth) AfterEventRegistration(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
	abUser, err := a.authBoss.LoadCurrentUser(&r)
	if abUser != nil && err == nil {
		user := abUser.(*users.User)
		if referalID, err := r.Cookie("refId"); err != nil {
			user.PutReferal(referalID.Value)
		}

		if hiveOsID, err := a.createUserAtHiveOS(user); err != nil {
			user.PutHiveOSUserID(hiveOsID)
		}

		a.db.Model(&user).Save(user)
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
		return err
	}
	newLog := a.userLogger.New()
	IP := r.RemoteAddr
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

func (a Auth) createUserAtHiveOS(u users.User) (int64, error) {
	resp, err := resty.R().SetFormData(map[string]string{
		"login":     u.Login,
		"name":      u.Name,
		"email":     u.Email,
		"ref_id":    u.ReferalID,
		"promocode": u.Promocode,
	}).Post(fmt.Sprintf("%s/api/int/users", a.conf.IDP.HiveOSApiURL))
	if err != nil {
		return 0, err
	}

	var responseStruct = make(map[string]interface{})
	if err := json.Unmarshal(resp.Body(), &responseStruct); err != nil {
		return 0, err
	}

	return responseStruct["id"].(int64), nil
}
