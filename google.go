package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleClient struct {
	ctx          context.Context
	driveService *drive.Service
}

func NewGoogleClient(account GoogleAccount) (*GoogleClient, error) {
	ctx := context.Background()

	b, err := os.ReadFile(account.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	// Try to detect if it's a service account
	var credType struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(b, &credType); err == nil && credType.Type == "service_account" {
		// Use service account
		config, err := google.JWTConfigFromJSON(b, drive.DriveReadonlyScope)
		if err != nil {
			return nil, fmt.Errorf("unable to parse service account: %w", err)
		}

		client := config.Client(ctx)
		driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			return nil, fmt.Errorf("unable to create Drive service: %w", err)
		}

		return &GoogleClient{
			ctx:          ctx,
			driveService: driveService,
		}, nil
	}

	// Otherwise use OAuth
	config, err := google.ConfigFromJSON(b, drive.DriveReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	client := getClient(config, account.TokenFile)

	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Drive service: %w", err)
	}

	return &GoogleClient{
		ctx:          ctx,
		driveService: driveService,
	}, nil
}

func getClient(config *oauth2.Config, tokenFile string) *http.Client {
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\nGo to the following link in your browser:\n%v\n\n", authURL)
	fmt.Print("Enter authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		panic(fmt.Sprintf("Unable to read authorization code: %v", err))
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		panic(fmt.Sprintf("Unable to retrieve token from web: %v", err))
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(fmt.Sprintf("Unable to cache oauth token: %v", err))
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func extractID(url string) (string, error) {
	re := regexp.MustCompile(`/d/([a-zA-Z0-9-_]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid Google Docs/Sheets URL")
	}
	return matches[1], nil
}

func (g *GoogleClient) ExportSheet(url, format string) ([]byte, error) {
	id, err := extractID(url)
	if err != nil {
		return nil, err
	}

	mimeType := getSheetMimeType(format)

	resp, err := g.driveService.Files.Export(id, mimeType).Download()
	if err != nil {
		return nil, fmt.Errorf("unable to export sheet: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (g *GoogleClient) ExportDoc(url, format string) ([]byte, error) {
	id, err := extractID(url)
	if err != nil {
		return nil, err
	}

	// For markdown, download as HTML and convert
	if format == "md" {
		htmlData, err := g.exportDocAsHTML(id)
		if err != nil {
			return nil, err
		}
		return htmlToMarkdown(htmlData), nil
	}

	mimeType := getDocMimeType(format)

	resp, err := g.driveService.Files.Export(id, mimeType).Download()
	if err != nil {
		return nil, fmt.Errorf("unable to export doc: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// DownloadDriveFile downloads any file from Google Drive by URL
func (g *GoogleClient) DownloadDriveFile(url string) ([]byte, string, error) {
	id, err := extractID(url)
	if err != nil {
		return nil, "", err
	}

	// Get file metadata to determine the filename
	file, err := g.driveService.Files.Get(id).Fields("name, mimeType").Do()
	if err != nil {
		return nil, "", fmt.Errorf("unable to get file metadata: %w", err)
	}

	// Download the file
	resp, err := g.driveService.Files.Get(id).Download()
	if err != nil {
		return nil, "", fmt.Errorf("unable to download file: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return data, file.Name, nil
}

func (g *GoogleClient) exportDocAsHTML(id string) ([]byte, error) {
	resp, err := g.driveService.Files.Export(id, "text/html").Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func getSheetMimeType(format string) string {
	switch format {
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "ods":
		return "application/vnd.oasis.opendocument.spreadsheet"
	case "pdf":
		return "application/pdf"
	case "csv":
		return "text/csv"
	case "html":
		return "text/html"
	default:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
}

func getDocMimeType(format string) string {
	switch format {
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "pdf":
		return "application/pdf"
	case "txt":
		return "text/plain"
	case "html":
		return "text/html"
	default:
		return "text/plain"
	}
}

func htmlToMarkdown(htmlData []byte) []byte {
	converter := md.NewConverter("", true, nil)

	// Don't keep HTML - convert everything to markdown
	markdown, err := converter.ConvertString(string(htmlData))
	if err != nil {
		// Fallback to basic conversion if error
		return basicHtmlToMarkdown(htmlData)
	}

	return []byte(markdown)
}

// Fallback basic converter
func basicHtmlToMarkdown(htmlData []byte) []byte {
	html := string(htmlData)

	// Remove HTML header and style tags
	html = regexp.MustCompile(`(?s)<head>.*?</head>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`).ReplaceAllString(html, "")

	// Convert headings
	html = regexp.MustCompile(`(?s)<h1[^>]*>(.*?)</h1>`).ReplaceAllString(html, "# $1\n\n")
	html = regexp.MustCompile(`(?s)<h2[^>]*>(.*?)</h2>`).ReplaceAllString(html, "## $1\n\n")
	html = regexp.MustCompile(`(?s)<h3[^>]*>(.*?)</h3>`).ReplaceAllString(html, "### $1\n\n")
	html = regexp.MustCompile(`(?s)<h4[^>]*>(.*?)</h4>`).ReplaceAllString(html, "#### $1\n\n")
	html = regexp.MustCompile(`(?s)<h5[^>]*>(.*?)</h5>`).ReplaceAllString(html, "##### $1\n\n")
	html = regexp.MustCompile(`(?s)<h6[^>]*>(.*?)</h6>`).ReplaceAllString(html, "###### $1\n\n")

	// Convert bold and italic
	html = regexp.MustCompile(`(?s)<strong[^>]*>(.*?)</strong>`).ReplaceAllString(html, "**$1**")
	html = regexp.MustCompile(`(?s)<b[^>]*>(.*?)</b>`).ReplaceAllString(html, "**$1**")
	html = regexp.MustCompile(`(?s)<em[^>]*>(.*?)</em>`).ReplaceAllString(html, "*$1*")
	html = regexp.MustCompile(`(?s)<i[^>]*>(.*?)</i>`).ReplaceAllString(html, "*$1*")

	// Convert links
	html = regexp.MustCompile(`(?s)<a[^>]*href="([^"]*)"[^>]*>(.*?)</a>`).ReplaceAllString(html, "[$2]($1)")

	// Convert lists
	html = regexp.MustCompile(`(?s)<li[^>]*>(.*?)</li>`).ReplaceAllString(html, "- $1\n")
	html = regexp.MustCompile(`<ul[^>]*>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`</ul>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`<ol[^>]*>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`</ol>`).ReplaceAllString(html, "\n")

	// Convert paragraphs
	html = regexp.MustCompile(`(?s)<p[^>]*>(.*?)</p>`).ReplaceAllString(html, "$1\n\n")

	// Convert line breaks
	html = regexp.MustCompile(`<br[^>]*>`).ReplaceAllString(html, "\n")

	// Convert code
	html = regexp.MustCompile(`(?s)<pre[^>]*>(.*?)</pre>`).ReplaceAllString(html, "```\n$1\n```\n")
	html = regexp.MustCompile(`(?s)<code[^>]*>(.*?)</code>`).ReplaceAllString(html, "`$1`")

	// Remove remaining HTML tags
	html = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(html, "")

	// Decode HTML entities
	replacements := map[string]string{
		"&nbsp;":  " ",
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
		"&quot;":  "\"",
		"&#39;":   "'",
		"&mdash;": "—",
		"&ndash;": "–",
	}
	for old, new := range replacements {
		html = regexp.MustCompile(old).ReplaceAllString(html, new)
	}

	// Clean up multiple newlines
	html = regexp.MustCompile(`\n{3,}`).ReplaceAllString(html, "\n\n")
	html = regexp.MustCompile(` {2,}`).ReplaceAllString(html, " ")

	return []byte(html)
}
