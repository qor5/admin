package login

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
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
