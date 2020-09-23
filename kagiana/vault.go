package kagiana

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"

	"github.com/hashicorp/vault/sdk/helper/certutil"
)

const VaultTimeout = 30

type Vault struct {
	client *api.Client
	config *Config
	token  string
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
	var secret *api.Secret
	switch config.OAuthProvider {
	case "github":
		s, err := client.Logical().Write("auth/github/login", map[string]interface{}{
			"token": strings.TrimSpace(m["github_token"]),
		})
		if err != nil {
			return nil, err
		}
		if s == nil {
			return nil, fmt.Errorf("empty response from credential provider")
		}
		secret = s

	default:
		return nil, fmt.Errorf("unknown provider %s", config.OAuthProvider)
	}

	client.SetToken(secret.Auth.ClientToken)
	return &Vault{
		client: client,
		config: config,
	}, nil
}

func (v *Vault) Token() string {
	return v.client.Token()
}
func (v *Vault) CreateCert() (map[string]*certutil.CertBundle, error) {
	cbs := map[string]*certutil.CertBundle{}
	for _, c := range v.config.Certs {
		ret, err := v.client.Logical().Write(c.Path, c.ToVaultOptions())
		if err != nil {
			return nil, err
		}
		cert, err := certutil.ParsePKIMap(ret.Data)
		if err != nil {
			return nil, err
		}
		b, err := cert.ToCertBundle()
		if err != nil {
			return nil, err
		}

		cbs[c.Common_Name] = b
	}
	return cbs, nil
}
