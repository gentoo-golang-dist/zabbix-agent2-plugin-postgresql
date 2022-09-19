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
	"database/sql"
	"errors"
	"strconv"

	"git.zabbix.com/ap/plugin-support/zbxerr"
	"github.com/jackc/pgx/v4"
)

// replicationHandler gets info about recovery state if all is OK or nil otherwise.
func replicationHandler(ctx context.Context, conn PostgresClient,
	key string, _ map[string]string, _ ...string) (interface{}, error) {
	var (
		replicationResult int64
		status            int
		query             string
		stringResult      sql.NullString
		inRecovery        bool
	)

	switch key {
	case keyReplicationStatus:
		row, err := conn.QueryRow(ctx, `SELECT pg_is_in_recovery()`)
		if err != nil {
			return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
		}

		err = row.Scan(&inRecovery)
		if err != nil {
			return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
		}

		if inRecovery {
			row, err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM pg_stat_wal_receiver`)
			if err != nil {
				return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
			}

			err = row.Scan(&status)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, zbxerr.ErrorEmptyResult.Wrap(err)
				}

				return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
			}
		} else {
			status = 2
		}

		return strconv.Itoa(status), nil

	case keyReplicationLagSec:
		query = `SELECT
					CASE
		  				WHEN NOT pg_is_in_recovery() OR pg_last_wal_receive_lsn() = pg_last_wal_replay_lsn() THEN 0
		  				ELSE COALESCE(EXTRACT(EPOCH FROM now() - pg_last_xact_replay_timestamp())::integer, 0)
					END AS lag;`
	case keyReplicationLagB:
		row, err := conn.QueryRow(ctx, `SELECT pg_is_in_recovery()`)
		if err != nil {
			return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
		}

		err = row.Scan(&inRecovery)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, zbxerr.ErrorEmptyResult.Wrap(err)
			}

			return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
		}

		if inRecovery {
			query = `SELECT pg_catalog.pg_wal_lsn_diff (pg_last_wal_receive_lsn(), pg_last_wal_replay_lsn());`
			row, err = conn.QueryRow(ctx, query)

			if err != nil {
				return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
			}

			err = row.Scan(&replicationResult)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, zbxerr.ErrorEmptyResult.Wrap(err)
				}

				return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
			}
		} else {
			replicationResult = 0
		}

		return replicationResult, nil

	case keyReplicationRecoveryRole:
		query = `SELECT pg_is_in_recovery()::int`

	case keyReplicationCount:
		query = `SELECT COUNT(DISTINCT client_addr) + COALESCE(SUM(CASE WHEN client_addr IS NULL THEN 1 ELSE 0 END), 0) FROM pg_stat_replication;`

	case keyReplicationProcessInfo:
		query = `SELECT json_object_agg(application_name, row_to_json(T))
				   FROM (
						SELECT
						    application_name,
							EXTRACT(epoch FROM COALESCE(flush_lag,'0'::interval)) AS flush_lag, 
							EXTRACT(epoch FROM COALESCE(replay_lag,'0'::interval)) AS replay_lag,
							EXTRACT(epoch FROM COALESCE(write_lag, '0'::interval)) AS write_lag
						FROM pg_stat_replication
					) T; `
		row, err := conn.QueryRow(ctx, query)

		if err != nil {
			return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
		}

		err = row.Scan(&stringResult)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, zbxerr.ErrorEmptyResult.Wrap(err)
			}

			return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
		}

		return stringResult.String, nil
	}

	row, err := conn.QueryRow(ctx, query)

	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&replicationResult)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	return replicationResult, nil
}
