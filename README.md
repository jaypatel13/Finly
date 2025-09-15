# Finly

A Gmail assistant application that authenticates with Google's OAuth2 and retrieves recent emails through the Gmail API.

## Features

- OAuth2 authentication with Google Gmail API
- Retrieve emails from the last 24 hours
- Extract email metadata (subject, sender, recipient, body content)
- Convert HTML email content to plain text
- JSON export of retrieved emails
- Web server for OAuth2 callback handling
- Token persistence for subsequent runs

## Installation

### Prerequisites

- Go 1.25+ installed
- A Google Cloud Platform project with Gmail API enabled
- OAuth2 credentials from Google Cloud Console

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/jaypatel13/Finly.git
   cd Finly
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up Google OAuth2 credentials:
   - Go to the [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select an existing one
   - Enable the Gmail API
   - Create OAuth2 credentials (Desktop application)
   - Download the credentials JSON file and save it as `credentials.json` in the project root

## Usage

### Basic Usage

Run the application:
```bash
go run main.go
```

The application will:
1. Start a web server on port 8080
2. Open your browser to authorize Gmail access
3. Retrieve emails from the last 24 hours
4. Save the results to `emails.json`

### OAuth Flow

1. The application will print an authorization URL to the console
2. Open the URL in your browser
3. Sign in with your Google account and grant permissions
4. You'll be redirected to a success page
5. The application will automatically proceed to fetch emails

## Configuration

### Files

- `credentials.json` - Google OAuth2 client credentials (required)
- `token.json` - OAuth2 access/refresh tokens (auto-generated)
- `emails.json` - Retrieved emails output (auto-generated)

### Environment Variables

Currently, the application uses default configurations:
- Server port: `8080`
- Email query: Last 24 hours
- User ID: `me` (authenticated user)

## Project Structure

```
.
├── auth/
│   └── oauth.go          # OAuth2 authentication manager
├── gmail/
│   └── client.go         # Gmail API client wrapper
├── server/
│   └── server.go         # HTTP server for OAuth callbacks
├── main.go               # Main application entry point
├── go.mod                # Go module dependencies
└── README.md             # This file
```

## API Documentation

### Gmail Client

The `gmail.Client` provides methods to:
- `GetEmails(date time.Time) ([]Email, error)` - Retrieve emails after a specific date
- `ListLabels() error` - List Gmail labels
- `StartWatch(topicName string)` - Start push notifications
- `StopWatch() error` - Stop push notifications

### Email Structure

```go
type Email struct {
    To            string `json:"to"`
    From          string `json:"from"`
    Subject       string `json:"subject"`
    BodyHTML      string `json:"bodyHtml"`
    BodyPlainText string `json:"bodyPlainText"`
    Snippet       string `json:"snippet"`
}
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go conventions and formatting (`go fmt`)
- Add appropriate error handling
- Include comments for exported functions
- Test OAuth flow with your own Gmail account

## Security Notes

- Keep your `credentials.json` file secure and never commit it to version control
- The `token.json` file contains sensitive access tokens
- Consider adding both files to `.gitignore`

## Dependencies

- `golang.org/x/oauth2` - OAuth2 client implementation
- `google.golang.org/api/gmail/v1` - Gmail API client
- `github.com/gorilla/mux` - HTTP router
- `github.com/k3a/html2text` - HTML to text conversion

## License

This project is open source. Please check the repository for license information or contact the maintainer.
