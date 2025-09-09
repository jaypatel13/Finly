package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type OAuthManager struct {
	config    *oauth2.Config
	tokenChan chan *oauth2.Token
	tokenFile string
}

func NewOAuthManager(credentialsFilePath, tokenFilePath string) (*OAuthManager, error) {
	credentialsJson, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(
		credentialsJson,
		gmail.MailGoogleComScope,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &OAuthManager{
		config:    config,
		tokenChan: make(chan *oauth2.Token, 1),
		tokenFile: tokenFilePath,
	}, nil
}

func (om *OAuthManager) GetAuthURL() string {
	return om.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

func (om *OAuthManager) GetClient() (*http.Client, error) {
	token, err := om.loadTokenFromFile()
	if err != nil {
		log.Printf("No valid token found, starting OAuth flow: %v", err)
		token, err = om.getTokenFromWeb()
		if err != nil {
			return nil, err
		}
		om.saveTokenToFile(token)
	} else {
		log.Printf("Using existing token from file")
	}

	tokenSource := &savingTokenSource{
		base:    om.config.TokenSource(context.Background(), token),
		manager: om,
	}

	return oauth2.NewClient(context.Background(), tokenSource), nil
}

func (om *OAuthManager) getTokenFromWeb() (*oauth2.Token, error) {
	authURL := om.config.AuthCodeURL("state-token",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"))
	fmt.Printf("Open this link in your browser to authorize:\n%v\n", authURL)
	fmt.Println("Waiting for authorization...")

	select {
	case token := <-om.tokenChan:
		fmt.Println("Authorization successful!")
		if token.RefreshToken == "" {
			log.Printf(" Warning: No refresh token received. You may need to reauthorize when the token expires.")
		} else {
			log.Printf("Refresh token received and will be saved")
		}
		return token, nil
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("timeout waiting for authorization")
	}
}

func (om *OAuthManager) loadTokenFromFile() (*oauth2.Token, error) {
	file, err := os.Open(om.tokenFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(token)
	return token, err
}

func (om *OAuthManager) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Callback handler called with URL: %s", r.URL.String())

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("Missing authorization code in callback")
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	log.Printf("Received authorization code, exchanging for token...")
	token, err := om.config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Token exchange successful, sending to channel...")
	select {
	case om.tokenChan <- token:
		log.Printf("Token sent to channel successfully")
	default:
		log.Printf("Token channel is full or not ready")
	}

	om.saveTokenToFile(token)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<html>
			<head><title>Authorization Complete</title></head>
			<body>
				<h1>Authorization Successful!</h1>
				<p>You can close this window and return to the application.</p>
			</body>
		</html>
	`))
}

func (om *OAuthManager) saveTokenToFile(token *oauth2.Token) {
	file, err := os.OpenFile(om.tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Warning: Unable to save token: %v", err)
		return
	}
	defer file.Close()
	json.NewEncoder(file).Encode(token)
	log.Printf("Token saved to: %s", om.tokenFile)
}

type savingTokenSource struct {
	base    oauth2.TokenSource
	manager *OAuthManager
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.base.Token()
	if err != nil {
		return nil, err
	}

	s.manager.saveTokenToFile(token)
	return token, nil
}
