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
	emails := c.service.Users.Messages.List(c.user).MaxResults(10)
	resp, err := emails.Do()
	if err != nil {
		return fmt.Errorf("failed to list emails: %w", err)
	}
	for _, msg := range resp.Messages {
		fmt.Println("Message Header:", msg.Payload.Headers)
	}
	return nil
}
