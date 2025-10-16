package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/qor5/admin/v3/cmd/qor5/website-template/admin"
	"github.com/theplant/osenv"
)

var (
	port       = osenv.Get("PORT", "The port to serve on", "9001")
	publishURL = osenv.Get("PUBLISH_URL", "Publish Target URL", "http://localhost:9002")

	writeTimeout = osenv.GetInt64("WriteTimeout", "HTTP write timeout in seconds - max time to write response (prevents slow clients)", 5)
	readTimeout  = osenv.GetInt64("ReadTimeout", "HTTP read timeout in seconds - max time to read request (prevents slowloris attacks)", 5)
)

func serverWithTimeouts(addr string, handler http.Handler) *http.Server {
	// Warn if timeouts are disabled in production
	if writeTimeout == 0 || readTimeout == 0 {
		log.Printf("WARNING: HTTP timeouts are disabled (WriteTimeout=%ds, ReadTimeout=%ds). This is unsafe for production use.", writeTimeout, readTimeout)
	}

	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
	}
}

func main() {
	// CMS server
	cmsMux := admin.Router(admin.ConnectDB())
	cmsServer := serverWithTimeouts(":"+port, cmsMux)
	go func() {
		if err := cmsServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	fmt.Println("CMS Served at http://localhost:" + port + "/admin")

	// Publish server
	u, _ := url.Parse(publishURL)
	publishPort := u.Port()
	if publishPort == "" {
		publishPort = "9002"
	}
	publishMux := http.FileServer(http.Dir(admin.PublishDir))
	publishServer := serverWithTimeouts(":"+publishPort, publishMux)
	fmt.Println("Publish Served at http://localhost:" + publishPort)
	log.Fatal(publishServer.ListenAndServe())
}
