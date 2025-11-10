package starter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qor5/x/v3/login"
	"github.com/theplant/inject"

	plogin "github.com/qor5/admin/v3/login"
)

// signClaims signs JWT claims with the provided secret
func signClaims(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.New("failed to sign claims")
	}
	return signed, nil
}

func createDummyAuthCookies(ctx context.Context, user *User, secret string, sessionBuilder *plogin.SessionBuilder) ([]*http.Cookie, error) {
	if user.GetSecure() == "" {
		return nil, errors.New("secure salt is empty")
	}

	cookies := make([]*http.Cookie, 0, 2)

	userID := fmt.Sprint(user.GetID())
	now := time.Now()
	maxAge := 3600
	claims := login.UserClaims{
		UserID:        userID,
		PassUpdatedAt: user.GetPasswordUpdatedAt(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(maxAge) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	tokenValue, err := signClaims(claims, secret)
	if err != nil {
		return nil, err
	}
	cookies = append(cookies, &http.Cookie{
		Name:     "auth",
		Value:    tokenValue,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	})

	tokenValue, err = signClaims(claims.RegisteredClaims, secret+user.GetSecure())
	if err != nil {
		return nil, err
	}
	cookies = append(cookies, &http.Cookie{
		Name:     "qor5_auth_secure",
		Value:    tokenValue,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	})

	if sessionBuilder != nil {
		r := httptest.NewRequest("GET", "/", http.NoBody).WithContext(ctx)
		for _, cookie := range cookies {
			r.AddCookie(cookie)
		}
		if err := sessionBuilder.CreateSession(r, userID); err != nil {
			return nil, errors.Wrap(err, "failed to create session")
		}
	}

	return cookies, nil
}

func (h *Handler) BuildForTest(ctx context.Context, ctors ...any) error {
	if err := AutoMigrate(ctx, h.DB); err != nil {
		return err
	}
	if err := h.Build(ctx, ctors...); err != nil {
		return err
	}

	var upsertUserOpts *UpsertUserOptions
	if err := h.ResolveContext(ctx, &upsertUserOpts); err != nil && !errors.Is(err, inject.ErrTypeNotProvided) {
		return err
	}
	if upsertUserOpts == nil {
		return nil
	}

	var sessionBuilder *plogin.SessionBuilder
	if err := h.ResolveContext(ctx, &sessionBuilder); err != nil {
		return err
	}

	user, err := UpsertUser(ctx, h.DB, upsertUserOpts)
	if err != nil {
		return err
	}

	uid := fmt.Sprint(user.GetID())
	if user.GetSecure() == "" {
		err := user.UpdateSecure(h.DB, user, uid)
		if err != nil {
			return err
		}
	}

	cookies, err := createDummyAuthCookies(ctx, user, h.Auth.Secret, sessionBuilder)
	if err != nil {
		return err
	}

	h.WithHandlerHook(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, cookie := range cookies {
				r.AddCookie(cookie)
			}
			next.ServeHTTP(w, r)
		})
	})

	return nil
}
