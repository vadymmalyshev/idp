package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/models"
	ginutils "git.tor.ph/hiveon/idp/pkg/gin"
	"git.tor.ph/hiveon/idp/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/baloo.v3"
	"net/http"
	"net/http/httptest"
	"testing"
)

var authTest *Auth

var test *baloo.Client

var conf *config.CommonConfig

func init() {
	flag.Parse()
	conf = config.InitViperConfig()
	r := gin.New()
	db := config.DB(conf.IDP.DB)
	models.Migrate(db)

	logger := log.NewLogger(log.Config{
		Level:  "debug",
		Format: "text",
	})

	r.Use(ginutils.Middleware(logger))

	authTest = NewAuth(r, db, conf)
	authTest.Init()
	go r.Run(conf.ServerConfig.Addr)

	test = baloo.New(fmt.Sprintf("http://%s", conf.ServerConfig.Addr))
	baloo.AddAssertFunc("serverErrResp", serverErrResponse)

	// crutch
	defer func() {
		for {
			_, err := http.Get(fmt.Sprintf("http://%s", conf.ServerConfig.Addr))
			fmt.Println("error: ", err)
			if err == nil {

				return
			}
		}
	}()

	//test = baloo.New("http://localhost:3000")
}

func TestAuth_Init(t *testing.T) {

}

func TestPostWrongLogin(t *testing.T) {
	requestBody, err := json.Marshal(map[string]string{
		"email":    "email@notfound.com",
		"password": "password",
	})

	if err != nil {
		logrus.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/idp/login", bytes.NewBuffer(requestBody))
	assert.Nil(t, err)

	recorder := httptest.NewRecorder()
	authTest.r.ServeHTTP(recorder, req)

	assert.Equal(t, 422, recorder.Code)
	assert.NotEqual(t, "{}", recorder.Body.String())

	var result map[string]interface{}

	json.NewDecoder(recorder.Body).Decode(&result)

	fmt.Println("body strings:")
	for key, value := range result {
		fmt.Println(key, value)
	}

	assert.Equal(t, false, result["success"])
	assert.Equal(t, "error", result["status"])

}

func TestPostRegister(t *testing.T) {

	requestBody, err := json.Marshal(map[string]string{
		"name":             "test",
		"login":            "test",
		"email":            "test@gmail.com",
		"password":         "testtest",
		"confirm_password": "testtest",
	})

	if err != nil {
		logrus.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/idp/register", bytes.NewBuffer(requestBody))
	assert.Nil(t, err)

	recorder := httptest.NewRecorder()
	authTest.r.ServeHTTP(recorder, req)

	assert.NotEqual(t, "{}", recorder.Body.String())
	assert.Equal(t, http.StatusOK, recorder.Code)

	var result map[string]interface{}
	json.NewDecoder(recorder.Body).Decode(&result)

	assert.NotEqual(t, nil, result["access_token"])
	assert.NotEqual(t, "", result["access_token"])

	assert.Equal(t, "bearer", result["token_type"])

	cookies := recorder.Result().Cookies()

	fmt.Println("Cookies:")
	for key, value := range cookies {
		fmt.Println(key, value)

	}

	checkCookies(t, cookies)

	fmt.Println("body strings:")

	for key, value := range result {
		fmt.Println(key, value)
	}
}

func TestPostLogin(t *testing.T) {
	requestBody, err := json.Marshal(map[string]string{
		"email":    "test@gmail.com",
		"password": "testtest",
		//"rm":       "true",
	})

	if err != nil {
		logrus.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/idp/login", bytes.NewBuffer(requestBody))
	assert.Nil(t, err)

	recorder := httptest.NewRecorder()
	authTest.r.ServeHTTP(recorder, req)

	assert.NotEqual(t, "{}", recorder.Body.String())
	assert.Equal(t, recorder.Code, http.StatusOK)

	var result map[string]interface{}

	json.NewDecoder(recorder.Body).Decode(&result)

	assert.NotEqual(t, nil, result["access_token"])
	assert.NotEqual(t, "", result["access_token"])
	assert.Equal(t, "bearer", result["token_type"])

	cookies := recorder.Result().Cookies()

	fmt.Println("Cookies:")
	for key, value := range cookies {
		fmt.Println(key, value)

	}

	checkCookies(t, cookies)

	fmt.Println("body strings:")
	for key, value := range result {
		fmt.Println(key, value)
	}

	//logrus.Printf("Body: %v", recorder.Body.String())
}

func TestPostLoginWithRememberMe(t *testing.T) {
	requestBody, err := json.Marshal(map[string]string{
		"email":    "test@gmail.com",
		"password": "testtest",
		"rm":       "true",
	})

	if err != nil {
		logrus.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/idp/login", bytes.NewBuffer(requestBody))
	assert.Nil(t, err)

	recorder := httptest.NewRecorder()
	authTest.r.ServeHTTP(recorder, req)

	assert.NotEqual(t, "{}", recorder.Body.String())
	assert.Equal(t, recorder.Code, http.StatusOK)

	var result map[string]interface{}

	json.NewDecoder(recorder.Body).Decode(&result)

	assert.NotEqual(t, nil, result["access_token"])
	assert.NotEqual(t, "", result["access_token"])
	assert.Equal(t, "bearer", result["token_type"])

	cookies := recorder.Result().Cookies()

	fmt.Println("Cookies:")
	for key, value := range cookies {
		fmt.Println(key, value)

	}

	checkCookies(t, cookies)

	fmt.Println("body strings:")
	for key, value := range result {
		fmt.Println(key, value)
	}

	//Test rm token
	fmt.Println("-----TEST RM TOKEN-----")
	recorder = httptest.NewRecorder()
	//req, err = http.NewRequest("GET", "/api/idp/userinfo", nil)
	req = httptest.NewRequest("GET", "/api/idp/userinfo", nil)
	if err != nil {
		t.Error(err)
	}

	// Set expired auth token:
	domain, err := config.GetCookieDomain()
	if err != nil {
		t.Error(err)
	}

	c := http.Cookie{
		Name:     "Authorization",
		Value:    "Bearer vBJZcI3vImSkbfrpnNFip9N8R_lIkXIZc474BVPfrrU.7w7OH0OAbmyoO0Dq1kgnFovE73hSLJ4I-u19-B5xHl4",
		Path:     "/",
		Domain:   domain,
		HttpOnly: false,
	}

	req.AddCookie(&c)
	req.AddCookie(cookies[2])
	req.AddCookie(cookies[4])

	authTest.r.ServeHTTP(recorder, req)
	fmt.Println("body strings:")

	fmt.Println(recorder.Body.String())

	assert.Equal(t, http.StatusOK, recorder.Code)
	fmt.Println("STATUSCODE:", recorder.Code)
	printCookies(recorder.Result().Cookies())
	//logrus.Printf("Body: %v", recorder.Body.String())
}

func serverErrResponse(res *http.Response, req *http.Request) error {
	if res.StatusCode >= 400 {
		return errors.New("invalid server response (> 400)")
	}
	return nil
}

func cookieCheck(res *http.Response, req *http.Request) error {
	printCookies(res.Cookies())
	return nil
}

func checkCookies(t *testing.T, cookies []*http.Cookie) {
	var expCookiesCount = 3
	if len(cookies) <= expCookiesCount {
		t.Errorf("not enough cookies, expected: >%d actual: %d", expCookiesCount, len(cookies))
	}

	var cookieTests = []struct {
		key   string
		value string
	}{
		{"login_csrftoken", ""},
		{"Authorization", ""},
		{"idp_session", ""},
	}

	for i, el := range cookieTests {
		assert.Equal(t, el.key, cookies[i].Name)
		assert.NotEqual(t, el.value, cookies[i].Value)
	}
}

func printCookies(cookies []*http.Cookie) {
	fmt.Println("Cookies:")
	if len(cookies) == 0 {
		fmt.Println("cookies are empty!")
	}
	for key, value := range cookies {
		fmt.Println(key, value)
	}
}

func TestBalooPostLogin(t *testing.T) {
	test.Post("/api/idp/login").
		JSON(map[string]string{"email": "test@gmail.com", "password": "testtest"}).
		Expect(t).
		Status(200).
		Type("json").
		//BodyEquals("").
		Assert("serverErrResp").
		AssertFunc(cookieCheck).
		Done()
}
