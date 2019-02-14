package hydra

import (
	"encoding/json"
	"fmt"

	"git.tor.ph/hiveon/idp/config"
	"github.com/ory/hydra/consent"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

var log *logrus.Logger

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
		}).Error("An error occurred while making an hydra challenge request")

		return authResult, err
	}

	json.Unmarshal(res.Body(), &authResult)
	return authResult, nil
}
