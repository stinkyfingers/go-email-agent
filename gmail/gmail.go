package gmail

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stinkyfingers/go-email-agent/tokenstorage"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Gmail struct {
	OAuthConfig  *oauth2.Config
	TokenStorage tokenstorage.Storage
}

type EmailSummary struct {
	ID      string `json:"id" jsonschema:"the Gmail message ID"`
	From    string `json:"from" jsonschema:"the sender of the email"`
	Subject string `json:"subject" jsonschema:"the subject of the email"`
	Body    string `json:"body" jsonschema:"the body of the email"`
	To      string `json:"to" jsonschema:"the recipient of the email"`
}

func NewGmail(tokenstorage tokenstorage.Storage) (*Gmail, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID is empty")
	}
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET is empty")
	}
	googleOAuthRedirect := os.Getenv("GOOGLE_OAUTH_REDIRECT_URL")
	if googleOAuthRedirect == "" {
		return nil, fmt.Errorf("GOOGLE_OAUTH_REDIRECT_URL is empty")
	}
	return &Gmail{
		TokenStorage: tokenstorage,
		OAuthConfig: &oauth2.Config{
			RedirectURL:  googleOAuthRedirect,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{gmail.GmailReadonlyScope, gmail.GmailComposeScope, "email", "openid"},
			Endpoint:     google.Endpoint,
		},
	}, nil
}

func (g *Gmail) RunOAuthServer(ctx context.Context) error {
	http.HandleFunc("/login", g.handleLogin)
	http.HandleFunc("/callback", g.handleCallback)

	log.Println("Server started on :8080")
	return http.ListenAndServe(":8080", nil)
}

/* handlers */

// 1. Redirect user to Google Auth URL
func (g *Gmail) handleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateStateOauthCookie(w)

	// AccessTypeOffline forces Google to return a Refresh Token
	// ApprovalForce ensures the prompt shows up if testing repeatedly
	url := g.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// 2. Receive Auth Code and Exchange for Tokens
func (g *Gmail) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Validate state cookie against CSRF
	oauthState, err := r.Cookie("oauthstate")
	if err != nil || r.FormValue("state") != oauthState.Value {
		log.Println("Invalid OAuth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := g.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Fprintf(w, "Code exchange failed: %s", err.Error())
		return
	}
	// Get email from token
	jwks, err := keyfunc.Get("https://www.googleapis.com/oauth2/v3/certs", keyfunc.Options{})
	if err != nil {
		fmt.Fprintf(w, "Token parse failed: %s", err.Error())
		return
	}
	idToken, _ := token.Extra("id_token").(string)
	jwtToken, err := jwt.Parse(idToken, jwks.Keyfunc)
	if err != nil {
		fmt.Fprintf(w, "Token parse failed: %s", err.Error())
		return
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Fprintf(w, "unexpected claims type %T", jwtToken.Claims)
		return
	}
	email, ok := claims["email"].(string)
	if !ok {
		fmt.Fprintf(w, "email not found in jw4t %T", jwtToken.Claims)
		return
	}
	err = g.TokenStorage.StoreToken(email, token)
	if err != nil {
		fmt.Fprintf(w, "Token storage failed: %s", err.Error())
		return
	}

	fmt.Fprintf(w, "gmail token stored successfully")
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	cookie := &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,  // Protects against XSS
		Secure:   false, // Set to true in production with HTTPS
	}
	http.SetCookie(w, cookie)
	return state
}

/* Service for use by other packages */

func (g *Gmail) GmailService(ctx context.Context, email string) (*gmail.Service, error) {
	token := &oauth2.Token{}
	err := g.TokenStorage.RetrieveToken(email, token)
	if err != nil {
		return nil, err
	}
	httpClient := g.OAuthConfig.Client(context.Background(), token)

	return gmail.NewService(ctx, option.WithHTTPClient(httpClient))
}

// extractPlainTextBody finds the first inline text/plain part in a message,
// recursing into nested multipart structures (e.g. multipart/mixed wrapping
// a multipart/alternative) and falling back to the top-level payload itself
// for single-part messages. Parts with a Filename are attachments, not the
// message body, even if their MimeType happens to be text/plain.
func extractPlainTextBody(part *gmail.MessagePart) string {
	if part == nil {
		return ""
	}
	if part.MimeType == "text/plain" && part.Filename == "" && part.Body != nil && part.Body.Data != "" {
		// Gmail's docs call this "base64url encoded" but don't guarantee
		// padding either way — it depends on the decoded byte length.
		// RawURLEncoding rejects any "=" padding outright, so strip it
		// first to handle both padded and unpadded data.
		data := strings.TrimRight(part.Body.Data, "=")
		decoded, err := base64.RawURLEncoding.DecodeString(data)
		if err != nil {
			log.Printf("unable to decode text/plain body part: %v", err)
			return ""
		}
		return string(decoded)
	}
	for _, sub := range part.Parts {
		if body := extractPlainTextBody(sub); body != "" {
			return body
		}
	}
	return ""
}

func CheckGmail(ctx context.Context, tokenStorage tokenstorage.Storage, email string) ([]EmailSummary, error) {
	g, err := NewGmail(tokenStorage)
	if err != nil {
		return nil, err
	}
	srv, err := g.GmailService(ctx, email)
	if err != nil {
		return nil, err
	}

	const user = "me"

	// Query for unread emails specifically in the INBOX.
	list, err := srv.Users.Messages.List(user).Q("label:INBOX is:unread").MaxResults(25).Do()
	if err != nil {
		return nil, fmt.Errorf("listing unread messages: %w", err)
	}

	// 6. Fetch just the headers we need for each message (cheaper than the
	// full message body, which we don't need at this stage).
	emails := make([]EmailSummary, 0, len(list.Messages))
	for _, item := range list.Messages {
		msg, err := srv.Users.Messages.Get(user, item.Id).
			Format("full").
			MetadataHeaders("Subject", "From").
			Do()
		if err != nil {
			log.Printf("unable to retrieve message %s: %v", item.Id, err)
			continue
		}

		summary := EmailSummary{ID: item.Id}
		for _, header := range msg.Payload.Headers {
			switch header.Name {
			case "Subject":
				summary.Subject = header.Value
			case "From":
				summary.From = header.Value
			}
		}
		summary.Body = extractPlainTextBody(msg.Payload)

		emails = append(emails, summary)
	}

	return emails, nil
}

func DraftEmail(ctx context.Context, messageID, body string, tokenStorage tokenstorage.Storage, email string) (string, error) {
	g, err := NewGmail(tokenStorage)
	if err != nil {
		return "", err
	}

	srv, err := g.GmailService(ctx, email)
	if err != nil {
		return "", err
	}

	const user = "me"

	// Look up the original message so the reply threads correctly: same
	// thread ID, "Re:" subject, and In-Reply-To/References pointing at the
	// original Message-Id.
	original, err := srv.Users.Messages.Get(user, messageID).
		Format("metadata").
		MetadataHeaders("Subject", "From", "Message-Id").
		Do()
	if err != nil {
		return "", fmt.Errorf("looking up message %s: %w", messageID, err)
	}

	var to, subject, originalMessageID string
	for _, header := range original.Payload.Headers {
		switch header.Name {
		case "From":
			to = header.Value
		case "Subject":
			subject = header.Value
		case "Message-Id":
			originalMessageID = header.Value
		}
	}
	if to == "" {
		return "", fmt.Errorf("message %s has no From header to reply to", messageID)
	}
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	var raw strings.Builder
	fmt.Fprintf(&raw, "To: %s\r\n", to)
	fmt.Fprintf(&raw, "Subject: %s\r\n", subject)
	if originalMessageID != "" {
		fmt.Fprintf(&raw, "In-Reply-To: %s\r\n", originalMessageID)
		fmt.Fprintf(&raw, "References: %s\r\n", originalMessageID)
	}
	raw.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	raw.WriteString("\r\n")
	raw.WriteString(body)

	draft, err := srv.Users.Drafts.Create(user, &gmail.Draft{
		Message: &gmail.Message{
			Raw:      base64.RawURLEncoding.EncodeToString([]byte(raw.String())),
			ThreadId: original.ThreadId,
		},
	}).Do()
	if err != nil {
		return "", fmt.Errorf("creating draft: %w", err)
	}

	return draft.Id, nil
}
