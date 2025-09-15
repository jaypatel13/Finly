package main

import (
	"context"
	"encoding/json"
	"finly/auth"
	"finly/gmail"
	"finly/server"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	oauthManager, err := auth.NewOAuthManager("credentials.json", "token.json")
	if err != nil {
		log.Fatalf("Failed to initialize OAuth: %v", err)
	}

	srv := server.NewServer("8080")
	srv.AddRoute("/oauth2/callback", oauthManager.CallbackHandler)

	log.Println("Starting server...")
	errCh := srv.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Println("Starting OAuth flow...")
	go func() {
		client, err := oauthManager.GetClient()
		if err != nil {
			log.Fatalf("failed to get authenticated client: %v", err)
		}

		gmailClient, err := gmail.NewClient(ctx, client, "me")
		if err != nil {
			log.Fatalf("failed to create Gmail client: %v", err)
		}

		emails, err := gmailClient.GetEmails(time.Now().Add(-24 * time.Hour))
		if err != nil {
			log.Fatalf("Failed to list emails: %v", err)
		}

		file, err := os.Create("emails.json")
		if err != nil {
			log.Fatalf("Failed to create emails.json: %v", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(emails); err != nil {
			log.Fatalf("Failed to encode emails to JSON: %v", err)
		}
		log.Println("Emails saved to emails.json")

	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	case <-stop:
		log.Println("Shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err := srv.Stop(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}
}
