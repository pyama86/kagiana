package kagiana

import (
	"github.com/STNS/libstns-go/libstns"
	"golang.org/x/oauth2"
)

type Config struct {
	PIDFile       string          `mapstructure:"pid_file"`
	LogFile       string          `mapstructure:"log_file"`
	LogLevel      string          `mapstructure:"log_level"`
	Listener      string          `mapstructure:"listener"`
	OAuthProvider string          `mapstructure:"oauth_provider"`
	OAuth         oauth2.Config   `mapstructure:"oauth"`
	Certs         []Cert          `mapstructure:"certs" validate:"required"`
	STNSEndpoint  string          `mapstructure:"stns_endpoint"`
	STNSOptions   libstns.Options `mapstructure:"stns_options"`
	VaultAuthPath string          `mapstructure:"vault_auth_path"`
}

type Cert struct {
	CommonName string `mapstructure:"common_name" validate:"required"`
	Path       string `validate:"required"`
	Format     string
	TTL        string
	AltNames   string
	IPSans     string
}

func (c Cert) ToVaultOptions() map[string]interface{} {
	r := map[string]interface{}{}
	r["common_name"] = c.CommonName
	if c.Format != "" {
		r["format"] = c.Format
	}

	if c.TTL != "" {
		r["ttl"] = c.TTL
	}

	if c.AltNames != "" {
		r["alt_names"] = c.AltNames
	}

	if c.IPSans != "" {
		r["ip_sans"] = c.IPSans
	}
	return r
}
