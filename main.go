package main

import (
	"context"
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

		if emails, err := gmailClient.ListEmails(); err != nil {
			log.Fatalf("Failed to list emails: %v", err)
			for _, email := range emails {
				log.Printf("To: %s, From: %s, Subject: %s, Body: %s", email.To, email.From, email.Subject, email.Body)
			}

		}

		log.Println("Gmail operations completed successfully!")
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
