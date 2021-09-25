package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
				Goth:    google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:9500/auth/callback?provider=google"),
				Key:     "google",
				Text:    "Login with Google",
				LogoURL: "",
			},
			&login.Provider{
				Goth:    twitter.New(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), "http://localhost:9500/auth/callback?provider=twitter"),
				Key:     "twitter",
				Text:    "Login with Twitter",
				LogoURL: "",
			},
		).
		FetchUserToContextFunc(func(claim *login.UserClaims, r *http.Request) (newR *http.Request, err error) {
			newR = r.WithContext(context.WithValue(r.Context(), "user", claim))
			return
		}).
		HomeURL("/admin")

	b.Mount(mux)
	mux.HandleFunc("/admin", b.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		claim := r.Context().Value("user").(*login.UserClaims)
		_, _ = fmt.Fprintf(w, "%#+v", claim)
	}))

	log.Println("serving at http://localhost:9500")
	log.Fatal(http.ListenAndServe(":9500", mux))
}
