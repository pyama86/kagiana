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

type STNSResponce struct {
	Token string
	Certs map[string]map[string]string
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

	userName := r.FormValue("user")
	userToken := r.FormValue("token")
	if err := stns.VerifyWithUser(userName, []byte(userToken), []byte(r.FormValue("signature"))); err != nil {
		logrus.Errorf("%s verify failed: %s", userName, err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logrus.Infof("login successfully %s", userName)
	vlt, err := NewVault(s.config, map[string]string{s.tokenType: userToken})
	if err != nil {
		logrus.Errorf("%s vault login failed: %s", userName, err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	cbs, err := vlt.CreateCert()
	if err != nil {
		logrus.Errorf("%s create cert failed: %s", userName, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ret := STNSResponce{
		Token: vlt.Token(),
	}

	certs := map[string]map[string]string{}
	for name, cb := range cbs {
		certs[name] = map[string]string{
			"ca":   strings.Join(cb.CAChain, "\\n"),
			"cert": cb.Certificate,
			"key":  cb.PrivateKey,
		}
	}

	ret.Certs = certs
	w.WriteHeader(http.StatusOK)

	b, err := json.Marshal(&ret)
	if err != nil {
		logrus.Errorf("%s json marshal failed: %s", userName, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(b))

	return

}
