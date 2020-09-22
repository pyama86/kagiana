package kagiana

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
	credGitHub "github.com/hashicorp/vault/builtin/credential/github"
)

const VaultTimeout = 30

type Vault struct {
	client *api.Client
}

func NewVault(config *Config, m map[string]string) (*Vault, error) {
	var httpClient = &http.Client{
		Timeout: VaultTimeout * time.Second,
	}

	client, err := api.NewClient(&api.Config{
		Address:    os.Getenv("VAULT_ADDR"),
		HttpClient: httpClient,
	})

	if err != nil {
		return nil, err
	}
	var handler vaultCommandLoginHandler
	switch config.OAuthProvider {
	case "github":
		handler = credGitHub.CLIHandler{}
	default:
		return nil, fmt.Errorf("unknown provider %s", config.OAuthProvider)
	}

	secret, err := handler.Auth(client, m)
	if err != nil {
		return nil, fmt.Error("Error authenticating: %s", err)
	}

	secret, _, err := c.extractToken(client, secret, false)
	if err != nil {
		return nil, fmt.Errorf("Error extracting token: %s", err)
	}
	if secret == nil {
		return nil, errors.New("Vault returned an empty secret")
	}

	client.SetToken(secret.Auth.ClientToken)
	return &Vault{
		client: client,
	}, nil
}
