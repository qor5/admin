package login

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/qor5/x/v3/i18n"
)

const (
	I18nAdminLoginKey i18n.ModuleKey = "I18nAdminLoginKey"
)

func getStringHash(v string, len int) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(v)))[:len]
}

func ip(r *http.Request) string {
	if r == nil {
		return ""
	}

	ips := proxy(r)
	if len(ips) > 0 && ips[0] != "" {
		rip, _, err := net.SplitHostPort(ips[0])
		if err != nil {
			rip = ips[0]
		}
		return rip
	}

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

func proxy(r *http.Request) []string {
	if ips := r.Header.Get("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}

	return nil
}

// func shortName(name string) string {
// 	if name == "" {
// 		return ""
// 	}
// 	runes := []rune(name)
// 	result := strings.ToUpper(string(runes[0:1]))
// 	if len(runes) > 2 {
// 		for i := 2; i < len(runes); i++ {
// 			if runes[i-1] == ' ' && runes[i] != ' ' {
// 				result += strings.ToUpper(string(runes[i : i+1]))
// 				break
// 			}
// 		}
// 	}
// 	return result
// }
