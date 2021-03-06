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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/pyama86/kagiana/kagiana"
)

func Test_requestSTNS(t *testing.T) {
	type args struct {
		authType  string
		token     string
		signature string
		userName  string
		code      string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				authType:  "stns",
				token:     "test toke",
				signature: "test sig",
				userName:  "test-user",
				code:      "test-code",
			},
			want: []string{
				"test.example.com.ca",
				"test.example.com.cert",
				"test.example.com.key",
				"token",
			},
			wantErr: false,
		},
		{
			name: "notfound",
			args: args{
				authType:  "notfound",
				token:     "test toke",
				signature: "test sig",
				userName:  "test-user",
				code:      "test-code",
			}, wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/auth/stns/challenge" {
					w.WriteHeader(http.StatusOK)
					fmt.Fprintf(w, tt.name)
				} else if r.URL.Path == "/auth/stns/verify" {
					if err := r.ParseForm(); err != nil {
						t.Error(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					if r.FormValue("code") != tt.args.code {
						t.Error(errors.New("unmatch code"))
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					w.WriteHeader(http.StatusOK)

					ret := kagiana.STNSResponce{
						Token: "test token",
						Certs: map[string]map[string]string{
							"test.example.com": map[string]string{
								"ca":   "ca value",
								"cert": "cert value",
								"key":  "key value",
							},
						},
					}

					b, err := json.Marshal(ret)
					if err != nil {
						t.Error(err)
					}
					fmt.Fprintf(w, string(b))

				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer ts.Close()

			dir, err := ioutil.TempDir("", "example")
			if err != nil {
				t.Error(err)
			}

			defer os.RemoveAll(dir)
			if err := verify(ts.URL, tt.args.authType, tt.args.token, tt.args.signature, tt.args.userName, dir, tt.args.code); (err != nil) != tt.wantErr {
				t.Errorf("verify() error = %v, wantErr %v", err, tt.wantErr)
			}
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				t.Error(err)
			}
			var paths []string
			for _, file := range files {
				paths = append(paths, file.Name())
			}

			if !reflect.DeepEqual(paths, tt.want) {
				t.Errorf("verify() = %v, want %v", paths, tt.want)
			}
		})
	}
}
