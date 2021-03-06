package kagiana

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

func NewGitHub(config *Config) *AuthGitHub {
	return &AuthGitHub{
		config:  config,
		getCert: getCert,
	}
}

type AuthGitHub struct {
	config  *Config
	getCert func(http.ResponseWriter, *http.Request, *Vault)
}

func (g *AuthGitHub) Login(w http.ResponseWriter, r *http.Request) {
	oAuthState := g.generateStateOAuthCookie(w)
	u := g.config.OAuth.AuthCodeURL(oAuthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func (g *AuthGitHub) Callback(w http.ResponseWriter, r *http.Request) {
	oAuthState, err := r.Cookie(CookieKey)
	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
		return
	}

	if err := r.ParseForm(); err != nil {
		RenderError(w, http.StatusUnauthorized, err)
		return
	}

	if r.FormValue("state") != oAuthState.Value {
		RenderError(w, http.StatusUnauthorized, err)
		return
	}

	token, err := g.getAccessToken(r.FormValue("code"))
	if err != nil {
		RenderError(w, http.StatusUnauthorized, err)
		return
	}

	vlt, err := NewVault(g.config, map[string]string{"github_token": token})
	if err != nil {
		RenderError(w, http.StatusUnauthorized, err)
		return
	}

	g.getCert(w, r, vlt)
}

func (g *AuthGitHub) generateStateOAuthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(3 * time.Minute)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: CookieKey, Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func (g *AuthGitHub) getAccessToken(code string) (string, error) {
	token, err := g.config.OAuth.Exchange(context.Background(), code)
	if err != nil {
		return "", fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	return token.AccessToken, nil
}
