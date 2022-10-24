//go:build !windows
// +build !windows

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
	"database/sql"
	"fmt"
	"os"
	"time"

	"git.zabbix.com/ap/plugin-support/log"
)

var sharedConn *PGConn

func getConnPool() (*PGConn, error) {
	return sharedConn, nil
}

func getEnv() (pgAddr, pgUser, pgPwd, pgDb string) {
	pgAddr = os.Getenv("PG_ADDR")
	pgUser = os.Getenv("PG_USER")
	pgPwd = os.Getenv("PG_PWD")
	pgDb = os.Getenv("PG_DB")

	if pgAddr == "" {
		pgAddr = "localhost:5432"
	}
	if pgUser == "" {
		pgUser = "postgres"
	}
	if pgPwd == "" {
		pgPwd = "postgres"
	}
	if pgDb == "" {
		pgDb = "postgres"
	}

	return
}

func createConnection() error {
	pgAddr, pgUser, pgPwd, pgDb := getEnv()

	connString := fmt.Sprintf("postgresql://%s:%s@%s/%s", pgUser, pgPwd, pgAddr, pgDb)

	newConn, err := sql.Open("pgx", connString)
	if err != nil {
		log.Critf("[createConnection] cannot create connection to PostgreSQL: %s", err.Error())

		return err
	}

	var version int

	err = newConn.QueryRow(`select current_setting('server_version_num');`).Scan(&version)
	if err != nil {
		log.Critf("[createConnection] cannot get PostgreSQL version: %s", err.Error())

		return err
	}

	sharedConn = &PGConn{
		client:         newConn,
		lastTimeAccess: time.Now(),
		version:        version,
		callTimeout:    30,
	}

	return nil
}
