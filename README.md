# kagiana

It is Web Interface Vault wrapper with OAuth.
I'm assuming this will work with kubernetes.
you could check example manifests.
## usage

```
Usage:
  kagiana server [flags]

Flags:
      --client-id string         oauth provider client id
      --client-secret string     oauth provider client secret
  -h, --help                     help for server
      --listener string          listen host (default "localhost:18080")
      --log-level string         log level(debug,info,warn,error) (default "info")
      --oauth-auth-url string    oauth auth url (default "https://github.com/login/oauth/authorize")
      --oauth-provider string    use oauth provier (default "github")
      --oauth-scopes strings     oauth scopes (default [user])
      --oauth-token-url string   oauth token url (default "https://github.com/login/oauth/access_token")
      --redirect-url string      oauth redirect url (default "http://localhost:18080/callback")

Global Flags:
      --config string   config file (default is $HOME/.kagiana)

```

## Install
### Homebrew
```bash
% brew tap pyama86/homebrew-kagiana
% brew install kagiana
```
