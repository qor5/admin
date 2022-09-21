package login

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
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
