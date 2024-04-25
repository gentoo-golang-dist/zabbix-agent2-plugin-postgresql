/*
** Zabbix
** Copyright 2001-2024 Zabbix SIA
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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"golang.zabbix.com/sdk/metric"
)

func Test_getParameters(t *testing.T) {
	type args struct {
		additional *additionalParam
	}
	tests := []struct {
		name string
		args args
		want []*metric.Param
	}{
		{
			"common parameters",
			args{nil},
			[]*metric.Param{
				paramURI,
				paramUsername,
				paramPassword,
				paramDatabase,
				paramTLSConnect,
				paramTLSCaFile,
				paramTLSCertFile,
				paramTLSKeyFile,
				paramCacheMode,
			},
		},
		{
			"empty additions map",
			args{&additionalParam{}},
			[]*metric.Param{
				paramURI,
				paramUsername,
				paramPassword,
				paramDatabase,
				paramTLSConnect,
				paramTLSCaFile,
				paramTLSCertFile,
				paramTLSKeyFile,
				paramCacheMode,
			},
		},
		{
			"with additional parameter",
			args{
				&additionalParam{
					param:    metric.NewParam("test", "Foo bar."),
					position: 4,
				},
			},
			[]*metric.Param{
				paramURI,
				paramUsername,
				paramPassword,
				paramDatabase,
				metric.NewParam("test", "Foo bar."),
				paramTLSConnect,
				paramTLSCaFile,
				paramTLSCertFile,
				paramTLSKeyFile,
				paramCacheMode,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getParameters(tt.args.additional); !reflect.DeepEqual(got, tt.want) {
				var gotString, wantString string

				for _, v := range got {
					gotString = fmt.Sprintf("%s %+v", gotString, v)
				}

				for _, v := range tt.want {
					wantString = fmt.Sprintf("%s %+v", wantString, v)
				}

				t.Errorf("getParameters() = %v,\n want %v", strings.TrimSpace(gotString), strings.TrimSpace(wantString))
			}
		})
	}
}
