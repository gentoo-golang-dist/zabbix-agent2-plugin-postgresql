//go:build postgresql_tests
// +build postgresql_tests

/*
** Zabbix
** Copyright 2001-2022 Zabbix SIA
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
	"context"
	"fmt"
	"testing"
)

func TestPlugin_archiveHandler(t *testing.T) {
	sharedPool, err := getConnPool()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx         context.Context
		conn        *PGConn
		key         string
		params      map[string]string
		extraParams []string
	}
	tests := []struct {
		name    string
		p       *Plugin
		args    args
		wantErr bool
	}{
		{
			fmt.Sprintf("archiveHandler should return json with data if OK "),
			&Impl,
			args{context.Background(), sharedPool, keyArchiveSize, nil, []string{}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := archiveHandler(tt.args.ctx, tt.args.conn, tt.args.key, tt.args.params, tt.args.extraParams...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Plugin.archiveHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got.(string)) == 0 {
				t.Errorf("Plugin.archiveHandler() = %v", got)
			}
		})
	}
}
