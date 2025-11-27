package admin

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/qor5/admin/v3/example/models"
	"gorm.io/gorm"
)

type DataTableHeader struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Width    string `json:"width"`
	Sortable bool   `json:"sortable"`
}

func getStringHash(v string, length int) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(v)))[:length]
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

func exportOrders(db *gorm.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var orders []*models.Order

		if err := db.Model(&models.Order{}).Find(&orders).Error; err != nil {
			panic(err)
		}

		name := time.Now().Format("20060102150405")
		w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="orders-%v.csv"`, name))
		w.Header().Set("Content-Type", req.Header.Get("Content-Type"))

		gocsv.Marshal(orders, w)
	})
}
