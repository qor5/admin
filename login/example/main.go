package main

import (
	"log"
	"net/http"

	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/twitter"
	"github.com/qor/qor5/login"
)

func main() {

	mux := http.NewServeMux()

	b := login.New().
		Secret("123").
		Providers(
			&login.Provider{
				Goth:    google.New("", "", ""),
				Key:     "google",
				Text:    "Login with Google",
				LogoURL: "",
			},
			&login.Provider{
				Goth:    twitter.New("", "", ""),
				Key:     "twitter",
				Text:    "Login with Twitter",
				LogoURL: "",
			},
		)

	b.Mount(mux)
	log.Println("serving at http://localhost:9500")
	log.Fatal(http.ListenAndServe(":9500", mux))
}
