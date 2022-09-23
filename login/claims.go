package login

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	errNoTokenString = errors.New("no token string")
	errTokenExpired  = errors.New("token expired")
	errInvalidToken  = errors.New("invalid token")
)

type UserClaims struct {
	Provider      string
	Email         string
	Name          string
	UserID        string
	AvatarURL     string
	Location      string
	IDToken       string
	PassUpdatedAt string
	TOTPValidated bool
	jwt.RegisteredClaims
}

func mustSignClaims(claims jwt.Claims, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return signed
}

func parseClaims(claims jwt.Claims, val string, secret string) (rc jwt.Claims, err error) {
	if val == "" {
		return nil, errNoTokenString
	}
	token, err := jwt.ParseWithClaims(val, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errTokenExpired
		}
		return nil, errInvalidToken
	}
	if !token.Valid {
		return nil, errInvalidToken
	}
	return token.Claims, nil
}

func parseBaseClaims(val string, secret string) (rc *jwt.RegisteredClaims, err error) {
	c, err := parseClaims(&jwt.RegisteredClaims{}, val, secret)
	if err != nil {
		return nil, err
	}
	rc, ok := c.(*jwt.RegisteredClaims)
	if !ok {
		return nil, errInvalidToken
	}
	return rc, nil
}

func parseClaimsFromCookie(r *http.Request, cookieName string, claims jwt.Claims, secret string) (rc jwt.Claims, err error) {
	tc, err := r.Cookie(cookieName)
	if err != nil || tc.Value == "" {
		return nil, errNoTokenString
	}
	return parseClaims(claims, tc.Value, secret)
}

func parseUserClaimsFromCookie(r *http.Request, cookieName string, secret string) (rc *UserClaims, err error) {
	c, err := parseClaimsFromCookie(r, cookieName, &UserClaims{}, secret)
	if err != nil {
		return nil, err
	}
	rc, ok := c.(*UserClaims)
	if !ok {
		return nil, errInvalidToken
	}
	return rc, nil
}

func parseBaseClaimsFromCookie(r *http.Request, cookieName string, secret string) (rc *jwt.RegisteredClaims, err error) {
	c, err := parseClaimsFromCookie(r, cookieName, &jwt.RegisteredClaims{}, secret)
	if err != nil {
		return nil, err
	}
	rc, ok := c.(*jwt.RegisteredClaims)
	if !ok {
		return nil, errInvalidToken
	}
	return rc, nil
}

// maxAge seconds
func genBaseClaims(sub string, maxAge int) jwt.RegisteredClaims {
	now := time.Now()
	return jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(maxAge) * time.Second)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Subject:   sub,
		ID:        uuid.New().String(),
	}
}
