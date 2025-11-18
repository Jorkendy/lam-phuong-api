package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func main() {
	b, err := os.ReadFile("credentials_1.json")
	if err != nil {
		log.Fatal(err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatal(err)
	}

	// Tạo URL OAuth
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Visit the URL below to authorize:")
	fmt.Println(authURL)

	// Webserver lắng callback
	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		tok, err := config.Exchange(context.Background(), code)
		if err != nil {
			log.Fatal("Token exchange error:", err)
		}

		fmt.Println("Your REFRESH TOKEN:")
		fmt.Println(tok.RefreshToken)

		w.Write([]byte("Done! Copy the refresh token from your console."))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
