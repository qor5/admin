package login

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type FailCode int

const (
	FailCodeSystemError FailCode = iota + 1
	FailCodeCompleteUserAuthFailed
	FailCodeUserNotFound
	FailCodeIncorrectUsernameOrPassword
)

const failCodeFlashCookieName = "qor5_login_fc_flash"

func setFailCodeFlash(w http.ResponseWriter, c FailCode) {
	http.SetCookie(w, &http.Cookie{
		Name:  failCodeFlashCookieName,
		Value: fmt.Sprint(c),
		Path:  "/",
	})
}

func GetFailCodeFlash(w http.ResponseWriter, r *http.Request) FailCode {
	c, err := r.Cookie(failCodeFlashCookieName)
	if err != nil {
		return 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:   failCodeFlashCookieName,
		Path:   "/",
		MaxAge: -1,
	})
	v, _ := strconv.Atoi(c.Value)
	return FailCode(v)
}

const wrongLoginInputFlashCookieName = "qor5_login_wi_flash"

type WrongLoginInputFlash struct {
	Iu string // incorrect username
	Ip string // incorrect password
}

func setWrongLoginInputFlash(w http.ResponseWriter, f WrongLoginInputFlash) {
	bf, _ := json.Marshal(&f)
	http.SetCookie(w, &http.Cookie{
		Name:  wrongLoginInputFlashCookieName,
		Value: base64.StdEncoding.EncodeToString(bf),
		Path:  "/",
	})
}

func GetWrongLoginInputFlash(w http.ResponseWriter, r *http.Request) WrongLoginInputFlash {
	c, err := r.Cookie(wrongLoginInputFlashCookieName)
	if err != nil {
		return WrongLoginInputFlash{}
	}
	http.SetCookie(w, &http.Cookie{
		Name:   wrongLoginInputFlashCookieName,
		Path:   "/",
		MaxAge: -1,
	})
	v, _ := base64.StdEncoding.DecodeString(c.Value)
	wi := WrongLoginInputFlash{}
	json.Unmarshal([]byte(v), &wi)
	return wi
}
