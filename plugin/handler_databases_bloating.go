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

// databasesBloatingHandler gets info about count and size of archive files and returns JSON if all is OK or nil otherwise.
func databasesBloatingHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (interface{}, error) {
	var countBloating int64

	query := `SELECT count(*)
				FROM pg_catalog.pg_stat_all_tables
	   		   WHERE (n_dead_tup/(n_live_tup+n_dead_tup)::float8) > 0.2
		 		 AND (n_live_tup+n_dead_tup) > 50;`

	row, err := conn.QueryRow(ctx, query)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&countBloating)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	return countBloating, nil
}
