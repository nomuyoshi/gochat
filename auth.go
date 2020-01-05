package main

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleOauth2 "google.golang.org/api/oauth2/v2"
)

type authHandler struct {
	next http.Handler
}

type authConfig struct {
	CLIENT_ID    string
	SECRET_KEY   string
	STATE        string
	REDIRECT_URL string
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("auth"); err == http.ErrNoCookie || cookie.Value == "" {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		panic(err.Error())
	} else {
		h.next.ServeHTTP(w, r)
	}
}

// MustAuth は任意のhttp.HandlerをラップしたauthHandlerを生成する
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

func googleConfig() *oauth2.Config {
	var config authConfig
	envconfig.Process("GOOGLE", &config)
	return &oauth2.Config{
		ClientID:     config.CLIENT_ID,
		ClientSecret: config.SECRET_KEY,
		RedirectURL:  config.REDIRECT_URL,
		Scopes:       []string{googleOauth2.UserinfoProfileScope},
		Endpoint:     google.Endpoint,
	}
}

func googleLoginHandler(w http.ResponseWriter, r *http.Request) {
	var config authConfig
	envconfig.Process("GOOGLE", &config)
	url := googleConfig().AuthCodeURL(config.STATE)
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func googleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	var config authConfig
	envconfig.Process("GOOGLE", &config)
	state := r.FormValue("state")
	if state != config.STATE {
		log.Fatal("error: invalid state")
	}

	conf := googleConfig()
	code := r.FormValue("code")
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}

	svc, err := googleOauth2.New(conf.Client(oauth2.NoContext, tok))
	if err != nil {
		log.Fatal(err)
	}

	info, err := svc.Userinfo.Get().Do()
	if err != nil {
		log.Fatal(err)
	}

	// Cookieにユーザ名を保存
	http.SetCookie(w, &http.Cookie{
		Name:  "auth",
		Value: base64.StdEncoding.EncodeToString([]byte(info.Name)),
		Path:  "/",
	})

	w.Header().Set("Location", "/chat")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.Header().Set("Location", "/login")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
