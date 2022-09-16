/*
** Zabbix
** Copyright (C) 2001-2022 Zabbix SIA
**
** This program is free software; you can redistribute it and/or modify
** it under the terms of the GNU General Public License as published by
** the Free Software Foundation; either version 2 of the License, or
** (at your option) any later version.
**
** This program is distributed in the hope that it will be useful,
** but WITHOUT ANY WARRANTY; without even the implied warranty of
** MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
** GNU General Public License for more details.
**
** You should have received a copy of the GNU General Public License
** along with this program; if not, write to the Free Software
** Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
**/

package handlers

import (
	"context"
	"fmt"
)

const (
	PingFailed = 0
	PingOk     = 1
)

// PingHandler queries 'SELECT 1' and returns PingOk if a connection is alive or PingFailed otherwise.
func PingHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (interface{}, error) {
	var res int

	row, err := conn.QueryRow(ctx, fmt.Sprintf("SELECT %d", PingOk))
	if err != nil {
		return PingFailed, nil
	}

	err = row.Scan(&res)

	if err != nil || res != PingOk {
		return PingFailed, nil
	}

	return PingOk, nil
}
