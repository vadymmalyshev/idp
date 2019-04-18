package auth

import (
	"fmt"
	"net/http"
	"strings"

	"git.tor.ph/hiveon/idp/config"
)

func SetAccessTokenCookie(w http.ResponseWriter, token string) {
	var value string

	splitToken := strings.Split(token, " ")

	if len(splitToken) < 2 && splitToken[0] != "" {
		value = splitToken[0]
	}
	if len(splitToken) > 1 {
		value = splitToken[1]
	}

	cookieDomain, _ := config.GetCookieDomain()

	cookie := http.Cookie{
		Name:   "Authorization",
		Value:  fmt.Sprintf("Bearer %s", value),
		Domain: cookieDomain,
		Path:   "/",
	}

	http.SetCookie(w, &cookie)
}
