package kagiana

import (
	"golang.org/x/oauth2"
)

type Config struct {
	PIDFile       string
	LogFile       string
	LogLevel      string
	Listener      string
	OAuthProvider string
	OAuth         oauth2.Config
	Certs         []Cert `validate:"required"`
}

type Cert struct {
	Common_Name string ` validate:"required"`
	Path        string `validate:"required"`
	Format      string
	TTL         string
	AltNames    string
	IPSans      string
}

func (c Cert) ToVaultOptions() map[string]interface{} {
	r := map[string]interface{}{}
	r["common_name"] = c.Common_Name
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
