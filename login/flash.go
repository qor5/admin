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
	FailCodeIncorrectPassword
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
	InfoCodePasswordSuccessfullyChanged
)

const failCodeFlashCookieName = "qor5_fc_flash"
const warnCodeFlashCookieName = "qor5_wc_flash"
const infoCodeFlashCookieName = "qor5_ic_flash"

func setFailCodeFlash(config CookieConfig, w http.ResponseWriter, c FailCode) {
	http.SetCookie(w, &http.Cookie{
		Name:     failCodeFlashCookieName,
		Value:    fmt.Sprint(c),
		Path:     config.Path,
		Domain:   config.Domain,
		HttpOnly: true,
	})
}

func GetFailCodeFlash(config CookieConfig, w http.ResponseWriter, r *http.Request) FailCode {
	c, err := r.Cookie(failCodeFlashCookieName)
	if err != nil {
		return 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:     failCodeFlashCookieName,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
	})
	v, _ := strconv.Atoi(c.Value)
	return FailCode(v)
}

func setWarnCodeFlash(config CookieConfig, w http.ResponseWriter, c WarnCode) {
	http.SetCookie(w, &http.Cookie{
		Name:     warnCodeFlashCookieName,
		Value:    fmt.Sprint(c),
		Path:     config.Path,
		Domain:   config.Domain,
		HttpOnly: true,
	})
}

func GetWarnCodeFlash(config CookieConfig, w http.ResponseWriter, r *http.Request) WarnCode {
	c, err := r.Cookie(warnCodeFlashCookieName)
	if err != nil {
		return 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:     warnCodeFlashCookieName,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
	})
	v, _ := strconv.Atoi(c.Value)
	return WarnCode(v)
}

func setInfoCodeFlash(config CookieConfig, w http.ResponseWriter, c InfoCode) {
	http.SetCookie(w, &http.Cookie{
		Name:     infoCodeFlashCookieName,
		Value:    fmt.Sprint(c),
		Path:     config.Path,
		Domain:   config.Domain,
		HttpOnly: true,
	})
}

func GetInfoCodeFlash(config CookieConfig, w http.ResponseWriter, r *http.Request) InfoCode {
	c, err := r.Cookie(infoCodeFlashCookieName)
	if err != nil {
		return 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:     infoCodeFlashCookieName,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
	})
	v, _ := strconv.Atoi(c.Value)
	return InfoCode(v)
}

const customErrorMessageFlashCookieName = "qor5_cem_flash"

func setCustomErrorMessageFlash(config CookieConfig, w http.ResponseWriter, f string) {
	http.SetCookie(w, &http.Cookie{
		Name:     customErrorMessageFlashCookieName,
		Value:    f,
		Path:     config.Path,
		Domain:   config.Domain,
		HttpOnly: true,
	})
}

func GetCustomErrorMessageFlash(config CookieConfig, w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie(customErrorMessageFlashCookieName)
	if err != nil {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     customErrorMessageFlashCookieName,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
	})
	return c.Value
}

const wrongLoginInputFlashCookieName = "qor5_wli_flash"

type WrongLoginInputFlash struct {
	Ia string // incorrect account name
	Ip string // incorrect password
}

func setWrongLoginInputFlash(config CookieConfig, w http.ResponseWriter, f WrongLoginInputFlash) {
	bf, _ := json.Marshal(&f)
	http.SetCookie(w, &http.Cookie{
		Name:     wrongLoginInputFlashCookieName,
		Value:    base64.StdEncoding.EncodeToString(bf),
		Path:     config.Path,
		Domain:   config.Domain,
		HttpOnly: true,
		Secure:   true,
	})
}

func GetWrongLoginInputFlash(config CookieConfig, w http.ResponseWriter, r *http.Request) WrongLoginInputFlash {
	c, err := r.Cookie(wrongLoginInputFlashCookieName)
	if err != nil {
		return WrongLoginInputFlash{}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     wrongLoginInputFlashCookieName,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
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

func setWrongForgetPasswordInputFlash(config CookieConfig, w http.ResponseWriter, f WrongForgetPasswordInputFlash) {
	bf, _ := json.Marshal(&f)
	http.SetCookie(w, &http.Cookie{
		Name:     wrongForgetPasswordInputFlashCookieName,
		Value:    base64.StdEncoding.EncodeToString(bf),
		Path:     config.Path,
		Domain:   config.Domain,
		HttpOnly: true,
		Secure:   true,
	})
}

func GetWrongForgetPasswordInputFlash(config CookieConfig, w http.ResponseWriter, r *http.Request) WrongForgetPasswordInputFlash {
	c, err := r.Cookie(wrongForgetPasswordInputFlashCookieName)
	if err != nil {
		return WrongForgetPasswordInputFlash{}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     wrongForgetPasswordInputFlashCookieName,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
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

func setWrongResetPasswordInputFlash(config CookieConfig, w http.ResponseWriter, f WrongResetPasswordInputFlash) {
	bf, _ := json.Marshal(&f)
	http.SetCookie(w, &http.Cookie{
		Name:     wrongResetPasswordInputFlashCookieName,
		Value:    base64.StdEncoding.EncodeToString(bf),
		Path:     config.Path,
		Domain:   config.Domain,
		HttpOnly: true,
		Secure:   true,
	})
}

func GetWrongResetPasswordInputFlash(config CookieConfig, w http.ResponseWriter, r *http.Request) WrongResetPasswordInputFlash {
	c, err := r.Cookie(wrongResetPasswordInputFlashCookieName)
	if err != nil {
		return WrongResetPasswordInputFlash{}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     wrongResetPasswordInputFlashCookieName,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})
	v, _ := base64.StdEncoding.DecodeString(c.Value)
	f := WrongResetPasswordInputFlash{}
	json.Unmarshal([]byte(v), &f)
	return f
}

const wrongChangePasswordInputFlashCookieName = "qor5_wcpi_flash"

type WrongChangePasswordInputFlash struct {
	OldPassword     string
	NewPassword     string
	ConfirmPassword string
}

func setWrongChangePasswordInputFlash(config CookieConfig, w http.ResponseWriter, f WrongChangePasswordInputFlash) {
	bf, _ := json.Marshal(&f)
	http.SetCookie(w, &http.Cookie{
		Name:     wrongChangePasswordInputFlashCookieName,
		Value:    base64.StdEncoding.EncodeToString(bf),
		Path:     config.Path,
		Domain:   config.Domain,
		HttpOnly: true,
		Secure:   true,
	})
}

func GetWrongChangePasswordInputFlash(config CookieConfig, w http.ResponseWriter, r *http.Request) WrongChangePasswordInputFlash {
	c, err := r.Cookie(wrongChangePasswordInputFlashCookieName)
	if err != nil {
		return WrongChangePasswordInputFlash{}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     wrongChangePasswordInputFlashCookieName,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})
	v, _ := base64.StdEncoding.DecodeString(c.Value)
	f := WrongChangePasswordInputFlash{}
	json.Unmarshal([]byte(v), &f)
	return f
}

const secondsToRedoFlashCookieName = "qor5_stre_flash"

func setSecondsToRedoFlash(config CookieConfig, w http.ResponseWriter, c int) {
	http.SetCookie(w, &http.Cookie{
		Name:     secondsToRedoFlashCookieName,
		Value:    fmt.Sprint(c),
		Path:     config.Path,
		Domain:   config.Domain,
		HttpOnly: true,
	})
}

func GetSecondsToRedoFlash(config CookieConfig, w http.ResponseWriter, r *http.Request) int {
	c, err := r.Cookie(secondsToRedoFlashCookieName)
	if err != nil {
		return 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:     secondsToRedoFlashCookieName,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
	})
	v, _ := strconv.Atoi(c.Value)
	return v
}
