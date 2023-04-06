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
	"errors"

	"git.zabbix.com/ap/plugin-support/zbxerr"
	"github.com/jackc/pgx/v4"
)

// walHandler executes select from directory which contains wal files and returns JSON if all is OK or nil otherwise.
func walHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (interface{}, error) {
	var walJSON string

	query := `SELECT row_to_json(T)
			    FROM (
					SELECT
						CASE
							WHEN pg_is_in_recovery() THEN 0
							ELSE pg_wal_lsn_diff(pg_current_wal_lsn(),'0/00000000')
						END AS WRITE,
						CASE 
							WHEN NOT pg_is_in_recovery() THEN 0
							ELSE pg_wal_lsn_diff(pg_last_wal_receive_lsn(),'0/00000000')
						END AS RECEIVE,
						count(*)
						FROM pg_ls_waldir() AS COUNT
					) T;`

	row, err := conn.QueryRow(ctx, query)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&walJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	return walJSON, nil
}
