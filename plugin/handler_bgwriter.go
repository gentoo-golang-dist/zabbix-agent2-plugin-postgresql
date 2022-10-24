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

// bgwriterHandler executes select  with statistics from pg_stat_bgwriter
// and returns JSON if all is OK or nil otherwise.
func bgwriterHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (interface{}, error) {
	var bgwriterJSON string

	query := `
  SELECT row_to_json (T)
    FROM (
          SELECT
              checkpoints_timed
            , checkpoints_req
            , checkpoint_write_time
            , checkpoint_sync_time
            , buffers_checkpoint
            , buffers_clean
            , maxwritten_clean
            , buffers_backend
            , buffers_backend_fsync
            , buffers_alloc
          FROM pg_catalog.pg_stat_bgwriter
		  ) T ;`

	row, err := conn.QueryRow(ctx, query)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&bgwriterJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	return bgwriterJSON, nil
}
