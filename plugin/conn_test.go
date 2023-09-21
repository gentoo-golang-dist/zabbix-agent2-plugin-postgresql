/*
** Zabbix
** Copyright 2001-2023 Zabbix SIA
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
**     http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
**/

package plugin

import (
	"strings"
	"testing"

	"git.zabbix.com/ap/plugin-support/tlsconfig"
)

func Test_createDNS(t *testing.T) {
	type args struct {
		host     string
		port     string
		dbname   string
		user     string
		password string
		details  tlsconfig.Details
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"default",
			args{host: "127.0.0.1", port: "123", dbname: "postgres", user: "foo"},
			[]string{"host=127.0.0.1", "port=123", "dbname=postgres", "user=foo"},
		},
		{
			"with_password",
			args{host: "127.0.0.1", port: "123", dbname: "postgres", user: "foo", password: "bar"},
			[]string{"host=127.0.0.1", "port=123", "dbname=postgres", "user=foo", "password=bar"},
		},
		{
			"tls_connect_require",
			args{
				host:    "127.0.0.1",
				port:    "123",
				dbname:  "postgres",
				user:    "foo",
				details: tlsconfig.Details{TlsConnect: "require"}},
			[]string{"host=127.0.0.1", "port=123", "dbname=postgres", "user=foo", "sslmode=require"},
		},
		{
			"tls_connect_verify_ca",
			args{
				host:    "127.0.0.1",
				port:    "123",
				dbname:  "postgres",
				user:    "foo",
				details: tlsconfig.Details{TlsConnect: "verify-ca", TlsCaFile: "path/to/ca"}},
			[]string{
				"host=127.0.0.1",
				"port=123",
				"dbname=postgres",
				"user=foo",
				"sslmode=verify-ca",
				"sslrootcert=path/to/ca",
			},
		},
		{
			"tls_full",
			args{
				host:   "127.0.0.1",
				port:   "123",
				dbname: "postgres",
				user:   "foo",
				details: tlsconfig.Details{
					TlsConnect:  "verify-full",
					TlsCaFile:   "path/to/ca",
					TlsCertFile: "path/to/cert",
					TlsKeyFile:  "path/to/key",
				}},
			[]string{
				"host=127.0.0.1", "port=123",
				"dbname=postgres",
				"user=foo",
				"sslmode=verify-full",
				"sslrootcert=path/to/ca",
				"sslcert=path/to/cert",
				"sslkey=path/to/key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := createDNS(
				tt.args.host,
				tt.args.port,
				tt.args.dbname,
				tt.args.user,
				tt.args.password,
				"",
				tt.args.details,
			)

			if !sameValues(strings.Split(tmp, " "), tt.want) {
				t.Errorf(
					"createDNS() = %v, want %v, test checks for values and not value order",
					tmp,
					strings.Join(tt.want, " "),
				)
			}
		})
	}
}

func Test_renameTLS(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"rqeuired", args{"required"}, "require"},
		{"verify_ca", args{"verify_ca"}, "verify-ca"},
		{"verify_full", args{"verify_full"}, "verify-full"},
		{"any_other_string", args{"foobar"}, "foobar"},
		{"empty", args{""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renameTLS(tt.args.in); got != tt.want {
				t.Errorf("renameTLS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func sameValues(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}

	dif := make(map[string]int)
	for _, v := range x {
		dif[v]++
	}

	for _, v := range y {
		if _, ok := dif[v]; !ok {
			return false
		}

		dif[v]--

		if dif[v] == 0 {
			delete(dif, v)
		}
	}

	return len(dif) == 0
}
