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

// archiveHandler gets info about count and size of archive files and returns JSON if all is OK or nil otherwise.
func archiveHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (interface{}, error) {
	var archiveCountJSON, archiveSizeJSON string

	queryArchiveCount := `SELECT row_to_json(T)
							FROM (
									SELECT archived_count, failed_count
								   	  FROM pg_stat_archiver
								) T;`

	queryArchiveSize := `SELECT row_to_json(T)
							FROM (
								WITH values AS (
									SELECT
										4096/(ceil(pg_settings.setting::numeric/1024/1024))::int AS segment_parts_count,
										setting::bigint AS segment_size,
										('x' || substring(pg_stat_archiver.last_archived_wal from 9 for 8))::bit(32)::int AS last_wal_div,
										('x' || substring(pg_stat_archiver.last_archived_wal from 17 for 8))::bit(32)::int AS last_wal_mod,
										CASE WHEN pg_is_in_recovery() THEN NULL 
											ELSE ('x' || substring(pg_walfile_name(pg_current_wal_lsn()) from 9 for 8))::bit(32)::int END AS current_wal_div,
										CASE WHEN pg_is_in_recovery() THEN NULL 
											ELSE ('x' || substring(pg_walfile_name(pg_current_wal_lsn()) from 17 for 8))::bit(32)::int END AS current_wal_mod
									FROM pg_settings, pg_stat_archiver
									WHERE pg_settings.name = 'wal_segment_size')
								SELECT 
									greatest(coalesce((segment_parts_count - last_wal_mod) + ((current_wal_div - last_wal_div - 1) * segment_parts_count) + current_wal_mod - 1, 0), 0) AS count_files,
									greatest(coalesce(((segment_parts_count - last_wal_mod) + ((current_wal_div - last_wal_div - 1) * segment_parts_count) + current_wal_mod - 1) * segment_size, 0), 0) AS size_files
								FROM values
							) T;`

	row, err := conn.QueryRow(ctx, queryArchiveCount)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&archiveCountJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	row, err = conn.QueryRow(ctx, queryArchiveSize)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&archiveSizeJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	result := archiveCountJSON[:len(archiveCountJSON)-1] + "," + archiveSizeJSON[1:]

	return result, nil
}
