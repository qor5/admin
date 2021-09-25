package login_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/qor/qor5/login"
)

func TestAuthenticate(t *testing.T) {
	b := login.New().Secret("123").
		FetchUserFunc(func(claim *login.UserClaims, r *http.Request) (newR *http.Request, err error) {
			newR = r.WithContext(context.WithValue(r.Context(), "user", claim))
			return
		})

	token, err := b.SignClaims(&login.UserClaims{
		Email:  "abc@gmail.com",
		Name:   "ABC",
		UserID: "123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "abc@gmail.com",
			ID:        "123",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	h := b.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		claim := r.Context().Value("user").(*login.UserClaims)
		_, _ = fmt.Fprintf(w, "%#+v", claim)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/home?auth="+token, nil)
	h.ServeHTTP(w, r)
	t.Logf("%s\n%s\n", token, w.Body.String())
	if !strings.Contains(w.Body.String(), `Email:"abc@gmail.com", Name:"ABC", UserID:"123"`) {
		t.Errorf("didn't get correct claim: %#+v", w.Body.String())
	}
}
