package hydra

import (
	"encoding/json"
	"fmt"
	"strconv"

	"git.tor.ph/hiveon/idp/pkg/errors"

	"git.tor.ph/hiveon/idp/config"
	"github.com/ory/hydra/consent"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

var log *logrus.Logger

var RememberFor = 30 * 24 * 60 * 60

func init() {
	log = config.Logger()
	resty.SetRedirectPolicy(resty.FlexibleRedirectPolicy(20))
}

func CheckChallengeCode(challenge string) (consent.AuthenticationRequest, error) {
	url := fmt.Sprintf("%s/oauth2/auth/requests/login/%s", config.HydraAdmin, challenge)
	authResult := consent.AuthenticationRequest{}

	res, err := resty.R().Get(url)
	if err != nil {
		log.Error(err)
		return authResult, err
	}

	if res.StatusCode() < 200 || res.StatusCode() > 302 {
		log.WithFields(logrus.Fields{
			"challenge": challenge,
		}).Debug("An error occurred while making an hydra challenge request")

		return authResult, err
	}

	json.Unmarshal(res.Body(), &authResult)
	return authResult, nil
}

func ConfirmLogin(userID uint, remember bool, challenge string) (LoginResponse, error) {
	url := fmt.Sprintf("%s/oauth2/auth/requests/login/%s/accept", config.HydraAdmin, challenge)

	response := LoginResponse{}

	request := LoginRequest{}
	request.Subject = strconv.FormatUint(uint64(userID), 10)
	request.Remember = remember
	request.RememberFor = RememberFor
	// request.ACR = "normal"

	res, err := resty.R().SetBody(request).
		SetHeader("Content-Type", "application/json").Put(url)
	if err != nil {
		log.WithFields(logrus.Fields{
			"Challenge": challenge,
			"UserID":    request.Subject,
		}).Debug("hydra/login/accept request failed")

		return response, errors.ErrHydraAcceptLogin
	}

	json.Unmarshal(res.Body(), &response)
	log.WithFields(logrus.Fields{"redirect_url": response.RedirectTo}).Info("redirect")
	return response, nil
}
