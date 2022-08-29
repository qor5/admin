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
	FailCodeIncorrectAccountNameOrPassword
	FailCodeUserLocked
	FailCodeAccountIsRequired
	FailCodePasswordCannotBeEmpty
	FailCodePasswordNotMatch
	FailCodeInvalidToken
	FailCodeTokenExpired
	FailCodeIncorrectTOTP
)

type WarnCode int

const (
	WarnCodePasswordHasBeenChanged = iota + 1
)

type InfoCode int

const (
	InfoCodePasswordSuccessfullyReset InfoCode = iota + 1
)

const failCodeFlashCookieName = "qor5_fc_flash"
const warnCodeFlashCookieName = "qor5_wc_flash"
const infoCodeFlashCookieName = "qor5_ic_flash"

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

func setWarnCodeFlash(w http.ResponseWriter, c WarnCode) {
	http.SetCookie(w, &http.Cookie{
		Name:  warnCodeFlashCookieName,
		Value: fmt.Sprint(c),
		Path:  "/",
	})
}

func GetWarnCodeFlash(w http.ResponseWriter, r *http.Request) WarnCode {
	c, err := r.Cookie(warnCodeFlashCookieName)
	if err != nil {
		return 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:   warnCodeFlashCookieName,
		Path:   "/",
		MaxAge: -1,
	})
	v, _ := strconv.Atoi(c.Value)
	return WarnCode(v)
}

func setInfoCodeFlash(w http.ResponseWriter, c InfoCode) {
	http.SetCookie(w, &http.Cookie{
		Name:  infoCodeFlashCookieName,
		Value: fmt.Sprint(c),
		Path:  "/",
	})
}

func GetInfoCodeFlash(w http.ResponseWriter, r *http.Request) InfoCode {
	c, err := r.Cookie(infoCodeFlashCookieName)
	if err != nil {
		return 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:   infoCodeFlashCookieName,
		Path:   "/",
		MaxAge: -1,
	})
	v, _ := strconv.Atoi(c.Value)
	return InfoCode(v)
}

const customErrorMessageFlashCookieName = "qor5_cem_flash"

func setCustomErrorMessageFlash(w http.ResponseWriter, f string) {
	http.SetCookie(w, &http.Cookie{
		Name:  customErrorMessageFlashCookieName,
		Value: f,
		Path:  "/",
	})
}

func GetCustomErrorMessageFlash(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie(customErrorMessageFlashCookieName)
	if err != nil {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:   customErrorMessageFlashCookieName,
		Path:   "/",
		MaxAge: -1,
	})
	return c.Value
}

const wrongLoginInputFlashCookieName = "qor5_wli_flash"

type WrongLoginInputFlash struct {
	Ia string // incorrect account name
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

const wrongForgetPasswordInputFlashCookieName = "qor5_wfpi_flash"

type WrongForgetPasswordInputFlash struct {
	Account string
}

func setWrongForgetPasswordInputFlash(w http.ResponseWriter, f WrongForgetPasswordInputFlash) {
	bf, _ := json.Marshal(&f)
	http.SetCookie(w, &http.Cookie{
		Name:  wrongForgetPasswordInputFlashCookieName,
		Value: base64.StdEncoding.EncodeToString(bf),
		Path:  "/",
	})
}

func GetWrongForgetPasswordInputFlash(w http.ResponseWriter, r *http.Request) WrongForgetPasswordInputFlash {
	c, err := r.Cookie(wrongForgetPasswordInputFlashCookieName)
	if err != nil {
		return WrongForgetPasswordInputFlash{}
	}
	http.SetCookie(w, &http.Cookie{
		Name:   wrongForgetPasswordInputFlashCookieName,
		Path:   "/",
		MaxAge: -1,
	})
	v, _ := base64.StdEncoding.DecodeString(c.Value)
	f := WrongForgetPasswordInputFlash{}
	json.Unmarshal([]byte(v), &f)
	return f
}

const wrongResetPasswordInputFlashCookieName = "qor5_wrpi_flash"

type WrongResetPasswordInputFlash struct {
	Password        string
	ConfirmPassword string
}

func setWrongResetPasswordInputFlash(w http.ResponseWriter, f WrongResetPasswordInputFlash) {
	bf, _ := json.Marshal(&f)
	http.SetCookie(w, &http.Cookie{
		Name:  wrongResetPasswordInputFlashCookieName,
		Value: base64.StdEncoding.EncodeToString(bf),
		Path:  "/",
	})
}

func GetWrongResetPasswordInputFlash(w http.ResponseWriter, r *http.Request) WrongResetPasswordInputFlash {
	c, err := r.Cookie(wrongResetPasswordInputFlashCookieName)
	if err != nil {
		return WrongResetPasswordInputFlash{}
	}
	http.SetCookie(w, &http.Cookie{
		Name:   wrongResetPasswordInputFlashCookieName,
		Path:   "/",
		MaxAge: -1,
	})
	v, _ := base64.StdEncoding.DecodeString(c.Value)
	f := WrongResetPasswordInputFlash{}
	json.Unmarshal([]byte(v), &f)
	return f
}

const secondsToRedoFlashCookieName = "qor5_fc_flash"

func setSecondsToRedoFlash(w http.ResponseWriter, c int) {
	http.SetCookie(w, &http.Cookie{
		Name:  secondsToRedoFlashCookieName,
		Value: fmt.Sprint(c),
		Path:  "/",
	})
}

func GetSecondsToRedoFlash(w http.ResponseWriter, r *http.Request) int {
	c, err := r.Cookie(secondsToRedoFlashCookieName)
	if err != nil {
		return 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:   secondsToRedoFlashCookieName,
		Path:   "/",
		MaxAge: -1,
	})
	v, _ := strconv.Atoi(c.Value)
	return v
}
