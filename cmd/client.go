/*
Copyright © 2020 pyama86

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/STNS/libstns-go/libstns"
	"github.com/pyama86/kagiana/kagiana"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "starting kagiana client",
	Long:  `It is starting kagiana client command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runClient(); err != nil {
			logrus.Fatal(err)
		}
	},
}
var endpoint string
var authType string
var userName string
var keyPath string
var keyPass string
var token string
var savePath string

func runClient() error {
	opt := &libstns.Options{
		PrivatekeyPath:     keyPath,
		PrivatekeyPassword: keyPass,
	}

	stns, err := libstns.NewSTNS("", opt)
	if err != nil {
		return err
	}

	signature, err := stns.Sign([]byte(token))
	if err != nil {
		return err
	}

	if err := requestSTNS(endpoint, authType, token, string(signature), userName, savePath); err != nil {
		return err
	}

	return nil
}

func requestSTNS(endpoint, authType, token, signature, userName, savePath string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("auth/%s/challenge", authType))
	u.RawQuery = fmt.Sprintf("user=%s", userName)
	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s can't get challenge code", userName)
	}

	defer resp.Body.Close()
	code, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Set("code", string(code))
	values.Set("token", token)
	values.Add("signature", signature)
	values.Add("user", userName)

	u, err = url.Parse(endpoint)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("auth/%s/verify", authType))

	req, err := http.NewRequest(
		"POST",
		u.String(),
		strings.NewReader(values.Encode()),
	)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		ret := kagiana.STNSResponce{}
		if err := json.Unmarshal(body, &ret); err != nil {
			return err
		}

		usr, _ := user.Current()

		file, err := os.Create(strings.Replace(path.Join(savePath, "token"), "~", usr.HomeDir, 1))
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.Write([]byte(ret.Token))
		if err != nil {
			return err
		}

		for name, keys := range ret.Certs {
			for keyType, keyValue := range keys {
				file, err := os.Create(strings.Replace(path.Join(savePath, fmt.Sprintf("%s.%s", name, keyType)), "~", usr.HomeDir, 1))
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = file.Write([]byte(keyValue))
				if err != nil {
					return err
				}

			}

		}
		return nil
	default:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("status code=%d, body=%s", resp.StatusCode, string(body))
	}
}

func init() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	clientCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "Kagiana Endpoint")
	clientCmd.PersistentFlags().StringVarP(&authType, "auth-type", "a", "stns", "Authentication type")
	clientCmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "Authentication User")
	clientCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "Authentication Token")
	clientCmd.PersistentFlags().StringVarP(&savePath, "savePath", "k", "~/.kagiana", "Certificate save path")

	clientCmd.PersistentFlags().StringVarP(&keyPath, "privatekey", "p", "~/.ssh/id_rsa", "PrivateKey Path")
	clientCmd.PersistentFlags().StringVarP(&keyPass, "privatekey-password", "s", "", "PrivateKey Password")

	clientCmd.MarkPersistentFlagRequired("endpoint")
	clientCmd.MarkPersistentFlagRequired("user")
	clientCmd.MarkPersistentFlagRequired("token")

	rootCmd.AddCommand(clientCmd)
}
