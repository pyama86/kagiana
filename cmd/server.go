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
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/pyama86/kagiana/kagiana"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "starting kagiana server",
	Long:  `It is starting kagiana servercommand.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := &kagiana.Config{}
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()
		if err := viper.Unmarshal(&config); err != nil {
			logrus.Fatal(err)
		}

		validate := validator.New()
		if err := validate.Struct(config); err != nil {
			logrus.Fatal(err)
		}
		switch config.LogLevel {
		case "debug":
			logrus.SetLevel(logrus.DebugLevel)
		case "info":
			logrus.SetLevel(logrus.InfoLevel)
		case "warn":
			logrus.SetLevel(logrus.WarnLevel)
		case "error":
			logrus.SetLevel(logrus.ErrorLevel)
		}

		if err := runServer(config); err != nil {
			logrus.Fatal(err)
		}
	},
}

func runServer(config *kagiana.Config) error {
	var provider kagiana.OAuthProvider
	tokenType := "github_token"
	switch config.OAuthProvider {
	case "github":
		provider = kagiana.NewGitHub(config)
	default:
		return fmt.Errorf("unknown provider %s", config.OAuthProvider)
	}

	stns := kagiana.NewSTNS(config, tokenType)
	mux := http.NewServeMux()
	mux.HandleFunc("/", provider.Login)
	mux.HandleFunc("/auth/stns/challenge", stns.Challenge)
	mux.HandleFunc("/auth/stns/verify", stns.Verify)
	mux.HandleFunc("/auth/stns", stns.Call)
	mux.HandleFunc("/callback", provider.Callback)

	server := http.Server{
		Handler: mux,
		Addr:    config.Listener,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
		<-quit
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		logrus.Info("starting shutdown kagiana")
		if err := server.Shutdown(ctx); err != nil {
			logrus.Errorf("shutting down the server: %s", err)
		}
	}()
	logrus.Info("starting kagiana")
	if err := server.ListenAndServe(); err != nil {
		if err.Error() != "http: Server closed" {
			logrus.Error(err)
		} else {
			logrus.Info("shutdown kagiana")
		}
	}

	return nil

}

func init() {
	serverCmd.PersistentFlags().String("log-level", "info", "log level(debug,info,warn,error)")
	viper.BindPFlag("LogLevel", serverCmd.PersistentFlags().Lookup("log-level"))

	serverCmd.PersistentFlags().String("oauth-provider", "github", "use oauth provier")
	viper.BindPFlag("OAuthProvider", serverCmd.PersistentFlags().Lookup("oauth-provider"))

	serverCmd.PersistentFlags().String("redirect-url", "http://localhost:18080/callback", "oauth redirect url")
	viper.BindPFlag("OAuth.RedirectURL", serverCmd.PersistentFlags().Lookup("redirect-url"))

	serverCmd.PersistentFlags().String("client-id", "", "oauth provider client id")
	viper.BindPFlag("OAuth.ClientID", serverCmd.PersistentFlags().Lookup("client-id"))

	serverCmd.PersistentFlags().String("client-secret", "", "oauth provider client secret")
	viper.BindPFlag("OAuth.ClientSecret", serverCmd.PersistentFlags().Lookup("client-secret"))

	serverCmd.PersistentFlags().String("oauth-auth-url", "https://github.com/login/oauth/authorize", "oauth auth url")
	viper.BindPFlag("OAuth.Endpoint.AuthURL", serverCmd.PersistentFlags().Lookup("oauth-auth-url"))

	serverCmd.PersistentFlags().String("oauth-token-url", "https://github.com/login/oauth/access_token", "oauth token url")
	viper.BindPFlag("OAuth.Endpoint.TokenURL", serverCmd.PersistentFlags().Lookup("oauth-token-url"))

	serverCmd.PersistentFlags().StringSlice("oauth-scopes", []string{"user"}, "oauth scopes")
	viper.BindPFlag("OAuth.Scopes", serverCmd.PersistentFlags().Lookup("oauth-scopes"))

	serverCmd.PersistentFlags().String("listener", "localhost:18080", "listen host")
	viper.BindPFlag("Listener", serverCmd.PersistentFlags().Lookup("listener"))

	rootCmd.AddCommand(serverCmd)
}
