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
)

// locksHandler executes select from pg_stat_database command and returns JSON if all is OK or nil otherwise.
func locksHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (interface{}, error) {
	var locksJSON string

	query := `
WITH T AS
	(SELECT db.datname dbname,
			lower(replace(Q.mode, 'Lock', '')) AS MODE,
			coalesce(T.qty, 0) val
	FROM pg_database db
	JOIN (
			VALUES ('AccessShareLock') ,('RowShareLock') ,('RowExclusiveLock') ,('ShareUpdateExclusiveLock') ,('ShareLock') ,('ShareRowExclusiveLock') ,('ExclusiveLock') ,('AccessExclusiveLock')) Q(MODE) ON TRUE NATURAL
	LEFT JOIN
		(SELECT datname,
			MODE,
			count(MODE) qty
		FROM pg_locks lc
		RIGHT JOIN pg_database db ON db.oid = lc.database
		GROUP BY 1, 2) T
	WHERE NOT db.datistemplate
	ORDER BY 1, 2)
SELECT json_object_agg(dbname, row_to_json(T2))
FROM
	(SELECT dbname,
			sum(val) AS total,
			sum(CASE
					WHEN MODE = 'accessexclusive' THEN val
				END) AS accessexclusive,
			sum(CASE
					WHEN MODE = 'accessshare' THEN val
				END) AS accessshare,
			sum(CASE
					WHEN MODE = 'exclusive' THEN val
				END) AS EXCLUSIVE,
			sum(CASE
					WHEN MODE = 'rowexclusive' THEN val
				END) AS rowexclusive,
			sum(CASE
					WHEN MODE = 'rowshare' THEN val
				END) AS rowshare,
			sum(CASE
					WHEN MODE = 'share' THEN val
				END) AS SHARE,
			sum(CASE
					WHEN MODE = 'sharerowexclusive' THEN val
				END) AS sharerowexclusive,
			sum(CASE
					WHEN MODE = 'shareupdateexclusive' THEN val
				END) AS shareupdateexclusive
	FROM T
	GROUP BY dbname) T2`

	row, err := conn.QueryRow(ctx, query)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&locksJSON)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	if len(locksJSON) == 0 {
		return nil, errors.New("cannot parse data")
	}

	return locksJSON, nil
}
