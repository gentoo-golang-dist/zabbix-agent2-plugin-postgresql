/*
** Copyright (C) 2001-2026 Zabbix SIA
**
** This program is free software: you can redistribute it and/or modify it under the terms of
** the GNU Affero General Public License as published by the Free Software Foundation, version 3.
**
** This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
** without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
** See the GNU Affero General Public License for more details.
**
** You should have received a copy of the GNU Affero General Public License along with this program.
** If not, see <https://www.gnu.org/licenses/>.
**/

package main

import (
	"errors"
	"fmt"
	"os"

	"golang.zabbix.com/plugin/postgresql/plugin"
	"golang.zabbix.com/sdk/plugin/container"
	"golang.zabbix.com/sdk/plugin/flag"
	"golang.zabbix.com/sdk/zbxerr"
)

const COPYRIGHT_MESSAGE = //
`Copyright (C) 2026 Zabbix SIA
License AGPLv3: GNU Affero General Public License version 3 <https://www.gnu.org/licenses/>.
This is free software: you are free to change and redistribute it according to
the license. There is NO WARRANTY, to the extent permitted by law.`

//nolint:revive // these are directives from the build scripts.
const (
	PLUGIN_VERSION_MAJOR = 7
	PLUGIN_VERSION_MINOR = 4
	PLUGIN_VERSION_PATCH = 12
	PLUGIN_VERSION_RC    = "rc1"
)

func main() {
	err := flag.HandleFlags(
		plugin.Name,
		os.Args[0],
		COPYRIGHT_MESSAGE,
		PLUGIN_VERSION_RC,
		PLUGIN_VERSION_MAJOR,
		PLUGIN_VERSION_MINOR,
		PLUGIN_VERSION_PATCH,
	)
	if err != nil {
		if !errors.Is(err, zbxerr.ErrorOSExitZero) {
			panic(fmt.Sprintf("failed to handle flags %s", err.Error()))
		}

		return
	}

	h, err := container.NewHandler(plugin.Impl.Name())
	if err != nil {
		panic(fmt.Sprintf("failed to create plugin handler %s", err.Error()))
	}

	plugin.Impl.Logger = h

	err = h.Execute()
	if err != nil {
		panic(fmt.Sprintf("failed to execute plugin handler %s", err.Error()))
	}
}
