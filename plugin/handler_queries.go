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
	"strconv"

	"git.zabbix.com/ap/plugin-support/zbxerr"
)

// queriesHandler executes select from pg_database command and returns JSON if all is OK or nil otherwise.
func queriesHandler(ctx context.Context, conn PostgresClient,
	_ string, params map[string]string, _ ...string) (interface{}, error) {
	var queriesJSON string

	period, err := strconv.Atoi(params["TimePeriod"])
	if err != nil {
		return nil, zbxerr.ErrorInvalidParams.Wrap(
			fmt.Errorf("TimePeriod must be an integer, %s", err.Error()),
		)
	}

	if period < 1 {
		return nil, zbxerr.ErrorInvalidParams.Wrap(
			fmt.Errorf("TimePeriod must be greater than 0"),
		)
	}

	exp := `^(\\s*(--[^\\n]*\\n|/\\*.*\\*/|\\n))*(autovacuum|VACUUM|ANALYZE|REINDEX|CLUSTER|CREATE|ALTER|TRUNCATE|DROP)`
	query := fmt.Sprintf(`WITH T AS (
		SELECT
			db.datname,
			coalesce(T.query_time_max, 0) query_time_max,
			coalesce(T.tx_time_max, 0) tx_time_max,
			coalesce(T.mro_time_max, 0) mro_time_max,
			coalesce(T.query_time_sum, 0) query_time_sum,
			coalesce(T.tx_time_sum, 0) tx_time_sum,
			coalesce(T.mro_time_sum, 0) mro_time_sum,
			coalesce(T.query_slow_count, 0) query_slow_count,
			coalesce(T.tx_slow_count, 0) tx_slow_count,
			coalesce(T.mro_slow_count, 0) mro_slow_count
		FROM
			pg_database db NATURAL
			LEFT JOIN (
				SELECT
					datname,
					extract(
						epoch
						FROM
							now()
					) :: integer ts,
					coalesce(
						max(
							extract(
								'epoch'
								FROM
									(clock_timestamp() - query_start)
							) :: integer * (
								state NOT IN (
									'idle',
									'idle in transaction',
									'idle in transaction (aborted)'
								)
								AND query !~* E'%s'
							) :: integer
						),
						0
					) query_time_max,
					coalesce(
						max(
							extract(
								'epoch'
								FROM
									(clock_timestamp() - query_start)
							) :: integer * (
								state NOT IN ('idle')
								AND query !~* E'%s'
							) :: integer
						),
						0
					) tx_time_max,
					coalesce(
						max(
							extract(
								'epoch'
								FROM
									(clock_timestamp() - query_start)
							) :: integer * (
								state NOT IN ('idle')
								AND query ~* E'%s'
							) :: integer
						),
						0
					) mro_time_max,
					coalesce(
						sum(
							extract(
								'epoch'
								FROM
									(clock_timestamp() - query_start)
							) :: integer * (
								state NOT IN (
									'idle',
									'idle in transaction',
									'idle in transaction (aborted)'
								)
								AND query !~* E'%s'
							) :: integer
						),
						0
					) query_time_sum,
					coalesce(
						sum(
							extract(
								'epoch'
								FROM
									(clock_timestamp() - query_start)
							) :: integer * (
								state NOT IN ('idle')
								AND query !~* E'%s'
							) :: integer
						),
						0
					) tx_time_sum,
					coalesce(
						sum(
							extract(
								'epoch'
								FROM
									(clock_timestamp() - query_start)
							) :: integer * (
								state NOT IN ('idle')
								AND query ~* E'%s'
							) :: integer
						),
						0
					) mro_time_sum,
					coalesce(
						sum(
							(
								extract(
									'epoch'
									FROM
										(clock_timestamp() - query_start)
								) > % d
							) :: integer * (
								state NOT IN (
									'idle',
									'idle in transaction',
									'idle in transaction (aborted)'
								)
								AND query !~* E'%s'
							) :: integer
						),
						0
					) query_slow_count,
					coalesce(
						sum(
							(
								extract(
									'epoch'
									FROM
										(clock_timestamp() - query_start)
								) > % d
							) :: integer * (
								state NOT IN ('idle')
								AND query !~* E'%s'
							) :: integer
						),
						0
					) tx_slow_count,
					coalesce(
						sum(
							(
								extract(
									'epoch'
									FROM
										(clock_timestamp() - query_start)
								) > % d
							) :: integer * (
								state NOT IN ('idle')
								AND query ~* E'%s'
							) :: integer
						),
						0
					) mro_slow_count
				FROM
					pg_stat_activity
				WHERE
					pid <> pg_backend_pid()
				GROUP BY
					1
			) T
		WHERE
			NOT db.datistemplate
	)
	SELECT
		json_object_agg(datname, row_to_json(T))
	FROM
		T`,
		exp, exp, exp, exp, exp, exp, period, exp, period, exp, period, exp)

	row, err := conn.QueryRow(ctx, query)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&queriesJSON)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	if len(queriesJSON) == 0 {
		return nil, zbxerr.ErrorCannotParseResult
	}

	return queriesJSON, nil
}
