package hydra

import (
	"encoding/json"
	"fmt"
	"strconv"

	"git.tor.ph/hiveon/idp/pkg/errors"

	"git.tor.ph/hiveon/idp/config"
	hydraConsent "github.com/ory/hydra/consent"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

var log *logrus.Logger

var RememberFor = 30 * 24 * 60 * 60

func init() {
	log = config.Logger()
	resty.SetRedirectPolicy(resty.FlexibleRedirectPolicy(20))

}

func AcceptConsentChallengeCode(challenge string) (string, error) {
	hConfig, _ := config.GetHydraConfig()
	url := fmt.Sprintf("%s/oauth2/auth/requests/consent/%s", hConfig.Admin, challenge)
	consent := hydraConsent.ConsentRequest{}

	res, err := resty.R().Get(url)

	if err != nil || res.StatusCode() < 200 || res.StatusCode() > 302 {
		log.Errorf("an error occured while making hydra accept consent url: %s", err.Error())
		return "", err
	}

	json.Unmarshal(res.Body(), &consent)

	req := hydraConsent.HandledConsentRequest{GrantedScope: getScopes(), GrantedAudience: consent.RequestedAudience,
		Remember: false, RememberFor: RememberFor}

	accept := hydraConsent.RequestHandlerResponse{}

	res, err = resty.R().
		SetBody(req).
		SetHeader("Content-Type", "application/json").
		Put(url + "/accept")

	if err != nil {
		log.Errorf("an error occured while making hydra accept consent url: %s", err.Error())
		return "", nil
	}

	json.Unmarshal(res.Body(), &accept)

	return accept.RedirectTo, nil
}

func CheckChallengeCode(challenge string) (hydraConsent.AuthenticationRequest, error) {
	hConfig, _ := config.GetHydraConfig()

	url := fmt.Sprintf("%s/oauth2/auth/requests/login/%s", hConfig.Admin, challenge)
	authResult := hydraConsent.AuthenticationRequest{}

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
	hConfig, _ := config.GetHydraConfig()
	url := fmt.Sprintf("%s/oauth2/auth/requests/login/%s/accept", hConfig.Admin, challenge)

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

func getScopes() []string {
	return []string{"openid", "offline"}
}
