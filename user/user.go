package user

import (
	"strings"

	"github.com/stinkyfingers/go-email-agent/agentfilestorage"
	"github.com/stinkyfingers/go-email-agent/tokenstorage"
)

type User struct {
	Email            string
	EmailType        EmailType
	AgentFileStorage agentfilestorage.Storage
	TokenStorage     tokenstorage.Storage
}

type EmailType string

const (
	Gmail EmailType = "gmail"
	// extensible to include additional email types
)

func NewUser(email string, tokenStorage tokenstorage.Storage, agentFileStorage agentfilestorage.Storage) (*User, error) {
	// for additional email types, switch emailType and token here
	emailSuffix := strings.Split(email, "@")[1]
	emailType := Gmail

	switch emailSuffix {
	case "gmail.com":
		emailType = Gmail
	default:
		// defaults to gmail
	}

	return &User{
		Email:            email,
		EmailType:        emailType,
		AgentFileStorage: agentFileStorage,
		TokenStorage:     tokenStorage,
	}, nil
}

func ListUsers(tokenStorage tokenstorage.Storage, agentFileStorage agentfilestorage.Storage) ([]*User, error) {
	emails, err := tokenStorage.ListEmails()
	if err != nil {
		return nil, err
	}
	var users []*User
	for _, email := range emails {
		user, err := NewUser(email, tokenStorage, agentFileStorage)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
