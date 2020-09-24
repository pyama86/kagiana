package kagiana

import "net/http"

const CookieKey = "kagiana_oauth_state"

type OAuthProvider interface {
	Login(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request)
}

func getCert(w http.ResponseWriter, r *http.Request, vlt *Vault) {
	certBundles, err := vlt.CreateCert()
	if err != nil {
		RenderError(w, http.StatusUnauthorized, err)
		return
	}

	RenderSuccess(w, certBundles, vlt.Token())
}
