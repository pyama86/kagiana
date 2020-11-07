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

func (s *STNS) getCertsAndToken(userName, userToken string) (map[string]map[string]string, string, error) {
	vlt, err := NewVault(s.config, map[string]string{s.tokenType: userToken})
	if err != nil {
		return nil, "", fmt.Errorf("%s vault login failed: %s", userName, err.Error())
	}

	cbs, err := vlt.CreateCert()
	if err != nil {
		return nil, "", fmt.Errorf("%s create cert failed: %s", userName, err.Error())
	}

	certs := map[string]map[string]string{}
	for name, cb := range cbs {
		certs[name] = map[string]string{
			"ca":   strings.Join(cb.CAChain, "\\n"),
			"cert": cb.Certificate,
			"key":  cb.PrivateKey,
		}
	}

	return certs, vlt.Token(), nil

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
	s.ResponceCerts(w, r, userName, userToken)
}

func (s *STNS) Challenge(w http.ResponseWriter, r *http.Request) {
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

	if userName == "" {
		logrus.Error("should set userName")
		w.WriteHeader(http.StatusBadRequest)

	}
	code, err := stns.CreateUserChallengeCode(userName)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusBadRequest)
	}

	logrus.Infof("%s get challenge code", userName)
	fmt.Fprintf(w, string(code))
}

func (s *STNS) Verify(w http.ResponseWriter, r *http.Request) {
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
	challengeCode := r.FormValue("code")
	if err := stns.VerifyWithUser(userName, []byte(challengeCode), []byte(r.FormValue("signature"))); err != nil {
		logrus.Errorf("%s verify failed: %s", userName, err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	code, err := stns.PopUserChallengeCode(userName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Warnf("%s can't pop challenge code", userName)
		return
	}

	if string(code) == r.FormValue("code") {
		s.ResponceCerts(w, r, userName, userToken)
		logrus.Infof("%s verify success", userName)
		return
	}
	logrus.Warnf("%s missmatch challenge code", userName)

	w.WriteHeader(http.StatusInternalServerError)
}

func (s *STNS) ResponceCerts(w http.ResponseWriter, r *http.Request, userName, userToken string) {
	certs, token, err := s.getCertsAndToken(userName, userToken)
	if err != nil {
		logrus.Errorf("%s vault auth failed: %s", userName, err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ret := STNSResponce{
		Token: token,
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
}
