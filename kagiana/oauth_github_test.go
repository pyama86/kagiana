package kagiana

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"golang.org/x/oauth2"
)

func TestAuthGitHub_Callback(t *testing.T) {
	type fields struct {
		config *Config
	}
	tests := []struct {
		name       string
		fields     fields
		wantStatus int
		cookie     string
		state      string
		code       string
	}{
		{
			name: "callback ok",
			fields: fields{
				config: &Config{
					OAuthProvider: "github",
					OAuth: oauth2.Config{
						RedirectURL:  "REDIRECT_URL",
						ClientID:     "id",
						ClientSecret: "secret",
					},
				},
			},
			wantStatus: http.StatusOK,
			cookie:     "test state",
			state:      "test state",
			code:       "test code",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &AuthGitHub{
				config: tt.fields.config,
				getCert: func(w http.ResponseWriter, r *http.Request, vlt *Vault) {
					if vlt.Token() != "test-token" {
						t.Errorf("Unexpected authorization token %q, want %q", vlt.Token(), "test-token")
					}
					w.WriteHeader(http.StatusOK)
				},
			}

			values := url.Values{}
			values.Set("state", tt.state)
			values.Set("code", tt.code)

			req := httptest.NewRequest("POST", "/callback", strings.NewReader(values.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Cookie", fmt.Sprintf("%s=%s", CookieKey, tt.cookie))
			resp := httptest.NewRecorder()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() != "/token" {
					t.Errorf("Unexpected exchange request URL %q", r.URL)
				}
				headerAuth := r.Header.Get("Authorization")
				if want := "Basic aWQ6c2VjcmV0"; headerAuth != want {
					t.Errorf("Unexpected authorization header %q, want %q", headerAuth, want)
				}
				headerContentType := r.Header.Get("Content-Type")
				if headerContentType != "application/x-www-form-urlencoded" {
					t.Errorf("Unexpected Content-Type header %q", headerContentType)
				}
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("Failed reading request body: %s.", err)
				}
				if string(body) != "code=test+code&grant_type=authorization_code&redirect_uri=REDIRECT_URL" {
					t.Errorf("Unexpected exchange payload; got %q", body)
				}
				w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
				w.Write([]byte("access_token=90d64460d14870c08c81352a05dedd3465940a7c&scope=user&token_type=bearer"))
			}))
			defer ts.Close()
			tt.fields.config.OAuth.Endpoint = oauth2.Endpoint{
				AuthURL:  ts.URL + "/auth",
				TokenURL: ts.URL + "/token",
			}

			tv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() != "/v1/auth/github/login" {
					t.Errorf("Unexpected exchange request URL %q", r.URL)
				}

				responseToken := &api.Secret{
					RequestID: "97a37d40-5227-71e5-07a4-25d699cb8118",
					Auth: &api.SecretAuth{
						ClientToken: "test-token",
					},
				}
				s, _ := json.Marshal(responseToken)
				w.Write(s)
			}))
			defer tv.Close()
			os.Setenv("VAULT_ADDR", tv.URL)

			g.Callback(resp, req)

			if resp.Code != tt.wantStatus {
				t.Errorf("callback status code does not match, expected %d, got %d", tt.wantStatus, resp.Code)
			}

		})
	}
}
