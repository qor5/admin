package admin

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type DataTableHeader struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Width    string `json:"width"`
	Sortable bool   `json:"sortable"`
}

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
