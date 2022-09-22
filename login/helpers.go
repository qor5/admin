package login

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

func underlyingReflectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return underlyingReflectType(t.Elem())
	}
	return t
}

func genHashSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// renewSession will reset cookie with the latest user info in DB.
func renewSession(
	b *Builder,
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	claims, err := parseUserClaimsFromCookie(r, b.authCookieName, b.secret)
	if err != nil {
		return
	}

	var user interface{}
	user, err = b.findUserByID(claims.UserID)
	if err != nil {
		return
	}

	if err = b.setSecureCookiesByClaims(w, user, *claims); err != nil {
		return
	}
	return nil
}

func mustSetQuery(u string, keyVals ...string) string {
	pu, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	if len(keyVals)%2 != 0 {
		panic("invalid keyVals")
	}

	q := pu.Query()
	for i := 0; i < len(keyVals); i += 2 {
		q.Set(keyVals[i], keyVals[i+1])
	}

	pu.RawQuery = ""
	return fmt.Sprintf("%s?%s", pu.String(), q.Encode())
}

func recaptchaTokenCheck(b *Builder, token string) bool {
	f := make(url.Values)
	f.Add("secret", b.recaptchaConfig.SecretKey)
	f.Add("response", token)

	res, err := http.Post("https://www.google.com/recaptcha/api/siteverify",
		"application/x-www-form-urlencoded", strings.NewReader(f.Encode()))
	if err != nil {
		log.Println(err)
		return false
	}
	defer res.Body.Close()

	var r struct {
		Success bool `json:"success"`
	}

	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Println(err)
		return false
	}

	return r.Success
}
