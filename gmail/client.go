package gmail

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Client struct {
	service *gmail.Service
	user    string
}

func NewClient(ctx context.Context, httpClient *http.Client, user string) (*Client, error) {
	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}
	return &Client{service: service, user: user}, nil
}

func (c *Client) ListEmails() error {
	// List messages (only returns IDs and thread IDs)
	resp, err := c.service.Users.Messages.List(c.user).MaxResults(10).Do()
	if err != nil {
		return fmt.Errorf("failed to list emails: %w", err)
	}

	if len(resp.Messages) == 0 {
		fmt.Println("📧 No emails found.")
		return nil
	}

	fmt.Printf("📧 Found %d recent emails:\n", len(resp.Messages))

	// Get full message details for each email
	for i, msg := range resp.Messages {
		fullMsg, err := c.service.Users.Messages.Get(c.user, msg.Id).Do()
		if err != nil {
			fmt.Printf("   %d. Error getting message: %v\n", i+1, err)
			continue
		}

		// Extract subject from headers
		subject := "No Subject"
		if fullMsg.Payload != nil && fullMsg.Payload.Headers != nil {
			for _, header := range fullMsg.Payload.Headers {
				if header.Name == "Subject" {
					subject = header.Value
					break
				}
			}
		}

		fmt.Printf("   %d. %s (ID: %s)\n", i+1, subject, msg.Id)
	}

	return nil
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

func (c *Client) StartWatch(topicName string) (*gmail.WatchResponse,error) {
	watchRquest := &gmail.WatchRequest{
		TopicName: topicName,
		LabelIds: []string{"INBOX"},
	}
	
	resp, err := c.service.Users.Watch(c.user, watchRquest).Do()
	if err!= nil{
		return nil, err
	}

	fmt.Printf("Watch started successfully!")
	fmt.Printf("    History ID: %d", resp.HistoryId)
	fmt.Printf("    Expiration: %d",resp.Expiration)

	return resp, nil
}

func (c *Client) StopWatch() error{
	err := c.service.Users.Stop(c.user).Do()
	if err!= nil{
		return fmt.Errorf("failed to stop watching for push notifications %w",err)
	}
	return nil
}