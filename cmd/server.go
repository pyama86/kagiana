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
	"syscall"
	"time"

	"github.com/facebookgo/pidfile"
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
		viper.SetEnvPrefix("Kagiana")
		viper.AutomaticEnv()
		if err := viper.Unmarshal(&config); err != nil {
			logrus.Fatal(err)
		}

		if config.LogFile != "" {
			f, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				logrus.Fatal("error opening file :" + err.Error())
			}
			logrus.SetOutput(f)
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
	pidfile.SetPidfilePath(config.PIDFile)
	if err := pidfile.Write(); err != nil {
		return err
	}

	defer func() {
		if err := os.Remove(pidfile.GetPidfilePath()); err != nil {
			logrus.Errorf("Error removing %s: %s", pidfile.GetPidfilePath(), err)
		}
	}()

	var provider kagiana.OAuthProvider
	switch config.OAuthProvider {
	case "github":
		provider = kagiana.NewGitHub(config)
	default:
		return fmt.Errorf("unknown provider %s", config.OAuthProvider)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", provider.Login)
	mux.HandleFunc("/callback", provider.Callback)

	server := http.Server{
		Handler: mux,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
		<-quit
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		logrus.Info("starting shutdown stnsd")
		if err := server.Shutdown(ctx); err != nil {
			logrus.Errorf("shutting down the server: %s", err)
		}
	}()
	logrus.Info("starting kagiana")
	if err := http.ListenAndServe(config.Listener, mux); err != nil {
		if err.Error() != "http: Server closed" {
			logrus.Error(err)
		} else {
			logrus.Info("shutdown kagiana")
		}
	}

	return nil

}

func init() {
	serverCmd.PersistentFlags().StringP("pid-file", "p", "/var/run/kagiana.pid", "pid file")
	viper.BindPFlag("PIDFile", serverCmd.PersistentFlags().Lookup("pid-file"))

	serverCmd.PersistentFlags().StringP("log-file", "l", "/var/log/kagiana.log", "log file")
	viper.BindPFlag("LogFile", serverCmd.PersistentFlags().Lookup("log-file"))

	serverCmd.PersistentFlags().String("log-level", "info", "log level(debug,info,warn,error)")
	viper.BindPFlag("LogLevel", serverCmd.PersistentFlags().Lookup("log-level"))

	serverCmd.PersistentFlags().String("oauth-provider", "github", "use oauth provier")
	viper.BindPFlag("OAuthProvider", serverCmd.PersistentFlags().Lookup("oauth-provider"))

	serverCmd.PersistentFlags().String("client-id", "", "oauth provider client id")
	viper.BindPFlag("OAuth.ClientID", serverCmd.PersistentFlags().Lookup("client-id"))

	serverCmd.PersistentFlags().String("client-secret", "", "oauth provider client secret")
	viper.BindPFlag("OAuth.ClientSecret", serverCmd.PersistentFlags().Lookup("client-secret"))

	serverCmd.PersistentFlags().String("github-auth-url", "https://github.com/login/oauth/authorize", "github auth url")
	viper.BindPFlag("OAuth.Endpoint.AuthURL", serverCmd.PersistentFlags().Lookup("github-auth-url"))

	serverCmd.PersistentFlags().String("github-token-url", "https://github.com/login/oauth/access_token", "github token url")
	viper.BindPFlag("OAuth.Endpoint.AuthURL", serverCmd.PersistentFlags().Lookup("github-token-url"))
	rootCmd.AddCommand(serverCmd)
}
