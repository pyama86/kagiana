package kagiana

import "net/http"

type OAuthProvider interface {
	Login(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request)
}
