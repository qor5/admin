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
)

type NoticeCode int

const (
	NoticeCodePasswordSuccessfullyReset NoticeCode = iota + 1
)

const failCodeFlashCookieName = "qor5_login_fc_flash"
const noticeCodeFlashCookieName = "qor5_login_nc_flash"

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

func setNoticeCodeFlash(w http.ResponseWriter, c NoticeCode) {
	http.SetCookie(w, &http.Cookie{
		Name:  noticeCodeFlashCookieName,
		Value: fmt.Sprint(c),
		Path:  "/",
	})
}

func GetNoticeCodeFlash(w http.ResponseWriter, r *http.Request) NoticeCode {
	c, err := r.Cookie(noticeCodeFlashCookieName)
	if err != nil {
		return 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:   noticeCodeFlashCookieName,
		Path:   "/",
		MaxAge: -1,
	})
	v, _ := strconv.Atoi(c.Value)
	return NoticeCode(v)
}

const wrongLoginInputFlashCookieName = "qor5_login_wi_flash"

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

const wrongForgetPasswordInputFlashCookieName = "qor5_login_fpi_flash"

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

const wrongResetPasswordInputFlashCookieName = "qor5_login_rpi_flash"

type WrongResetPasswordInputFlash struct {
	Password        string
	ConfirmPassword string
}

func setWrongResetPasswordInputFlash(w http.ResponseWriter, f WrongResetPasswordInputFlash) {
	bf, _ := json.Marshal(&f)
	http.SetCookie(w, &http.Cookie{
		Name:  wrongLoginInputFlashCookieName,
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
