/*
** Zabbix
** Copyright (C) 2001-2025 Zabbix SIA
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
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func Test_versionHandler(t *testing.T) {
	type mock struct {
		row *sqlmock.Rows
		err error
	}

	tests := []struct {
		name    string
		mock    mock
		want    any
		wantErr bool
	}{
		{
			"+valid",
			mock{
				row: sqlmock.NewRows([]string{"version"}).AddRow("postgres 69"),
			},
			"postgres 69",
			false,
		},
		{
			"-queryErr",
			mock{
				row: sqlmock.NewRows([]string{"version"}).AddRow("postgres 69"),
				err: errors.New("query err"),
			},
			nil,
			true,
		},
		{
			"-noRows",
			mock{
				row: sqlmock.NewRows([]string{"version"}),
			},
			nil,
			true,
		},
		{
			"-scanErr",
			mock{row: sqlmock.NewRows([]string{"version"}).AddRow(nil)},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sql mock: %s", err.Error())
			}

			defer db.Close()

			mock.ExpectQuery(`^SELECT version\(\);$`).
				WillReturnRows(tt.mock.row).
				WillReturnError(tt.mock.err)

			got, err := versionHandler(
				context.Background(), &PGConn{client: db}, "", nil,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"versionHandler() error = %v, wantErr %v", err, tt.wantErr,
				)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("versionHandler() = %v, want %v", got, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf(
					"versionHandler() sql mock expectations where not met: %s",
					err.Error(),
				)
			}
		})
	}
}
