package kagiana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/STNS/libstns-go/libstns"
	"github.com/sirupsen/logrus"
)

type STNS struct {
	config    *Config
	tokenType string
}

func NewSTNS(config *Config, tokenType string) *STNS {
	return &STNS{
		config:    config,
		tokenType: tokenType,
	}
}

func (s *STNS) Call(w http.ResponseWriter, r *http.Request) {
	stns, err := libstns.NewSTNS(s.config.STNSEndpoint, &s.config.STNSOptions)

	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := stns.VerifyWithUser(r.FormValue("user"), []byte(r.FormValue("token")), []byte(r.FormValue("signature"))); err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vlt, err := NewVault(s.config, map[string]string{s.tokenType: r.FormValue("token")})
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	cbs, err := vlt.CreateCert()
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ret := map[string]map[string]string{}

	for name, cb := range cbs {
		ret[name] = map[string]string{
			"ca":   strings.Join(cb.CAChain, "\\n"),
			"cert": cb.Certificate,
			"key":  cb.PrivateKey,
		}
	}

	w.WriteHeader(http.StatusOK)

	b, err := json.Marshal(&ret)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(b))

	return

}
