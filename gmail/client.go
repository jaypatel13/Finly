package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Client struct {
	service *gmail.Service
	user    string
}

type Email struct {
	To      string
	From    string
	Subject string
	Body    string
}

func NewClient(ctx context.Context, httpClient *http.Client, user string) (*Client, error) {
	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}
	return &Client{service: service, user: user}, nil
}

func GetEmail(srv *gmail.Service, messageId string) {

	gmailMessageResposne, err := srv.Users.Messages.Get("me", messageId).Format("RAW").Do()
	if err != nil {
		log.Println("error when getting mail content: ", err)
	}

	if gmailMessageResposne != nil {

		decodedData, err := base64.RawURLEncoding.DecodeString(gmailMessageResposne.Raw)

		if err != nil {
			log.Println("error b64 decoding: ", err)
		}
		fmt.Printf("- %s\n", decodedData)
	}
}

func (c *Client) ListEmails() ([]Email, error) {

	query := fmt.Sprintf("after:%s", time.Now().Format(time.DateOnly))
	resp, err := c.service.Users.Messages.List(c.user).Q(query).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list emails: %w", err)
	}

	if len(resp.Messages) == 0 {
		fmt.Println("No emails found.")
		return nil, nil
	}

	fmt.Printf("Found %d recent emails:\n", len(resp.Messages))

	emails := []Email{}

	// Get full message details for each email
	for i, msg := range resp.Messages {
		// GetEmail(c.service, msg.Id)
		fullMsg, err := c.service.Users.Messages.Get(c.user, msg.Id).Do()
		if err != nil {
			fmt.Printf("   %d. Error getting message: %v\n", i+1, err)
			continue
		}

		// Extract subject from headers
		subject := "No Subject"
		to := ""
		from := ""
		body := ""

		if fullMsg.Payload != nil && fullMsg.Payload.Headers != nil {
			for _, header := range fullMsg.Payload.Headers {
				switch header.Name {
				case "Subject":
					subject = header.Value
				case "To":
					to = header.Value
				case "From":
					from = header.Value
				}
			}
		}

		if fullMsg.Payload != nil {
			for _, part := range fullMsg.Payload.Parts {
				if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
					decodedBody, err := base64.URLEncoding.DecodeString(part.Body.Data)
					if err != nil {
						log.Printf("error decoding email body: %v", err)
					} else {
						body = string(decodedBody)
					}
				}
			}
		}

		curEmail := Email{
			To:      to,
			From:    from,
			Subject: subject,
			Body:    body,
		}

		emails = append(emails, curEmail)
		fmt.Printf("   %d. To: %s\n", i+1, to)
		fmt.Printf("      From: %s\n", from)
		fmt.Printf("      Subject: %s\n", subject)
		fmt.Printf("      Body: %s\n\n", body)
	}

	return emails, nil
}

func (c *Client) ListLabels() error {
	labels, err := c.service.Users.Labels.List(c.user).Do()
	if err != nil {
		return fmt.Errorf("failed to retrieve labels: %v", err)
	}

	if len(labels.Labels) == 0 {
		fmt.Println("📧 No labels found.")
		return nil
	}

	fmt.Printf("📧 Found %d Gmail labels:\n", len(labels.Labels))
	for _, label := range labels.Labels {
		fmt.Printf("   • %s\n", label.Name)
	}

	return nil
}

func (c *Client) StartWatch(topicName string) (*gmail.WatchResponse, error) {
	watchRquest := &gmail.WatchRequest{
		TopicName: topicName,
		LabelIds:  []string{"INBOX"},
	}

	resp, err := c.service.Users.Watch(c.user, watchRquest).Do()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Watch started successfully!")
	fmt.Printf("    History ID: %d", resp.HistoryId)
	fmt.Printf("    Expiration: %d", resp.Expiration)

	return resp, nil
}

func (c *Client) StopWatch() error {
	err := c.service.Users.Stop(c.user).Do()
	if err != nil {
		return fmt.Errorf("failed to stop watching for push notifications %w", err)
	}
	return nil
}
