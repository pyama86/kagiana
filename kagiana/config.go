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
}
