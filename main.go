/*
** Zabbix
** Copyright 2001-2024 Zabbix SIA
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

package main

import (
	"errors"
	"fmt"
	"os"

	"git.zabbix.com/ap/plugin-support/plugin/container"
	"git.zabbix.com/ap/plugin-support/plugin/flag"
	"git.zabbix.com/ap/plugin-support/zbxerr"
	"git.zabbix.com/ap/postgresql/plugin"
)

// DO NOT GROUP THESE CONSTANTS! The makefile needs, them as is.
const PLUGIN_VERSION_MAJOR = 6
const PLUGIN_VERSION_MINOR = 0
const PLUGIN_VERSION_PATCH = 27
const PLUGIN_VERSION_RC = ""

const COPYRIGHT_MESSAGE = //
`Copyright 2001-2024 Zabbix SIA
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.`

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
	plugin.Impl.Logger = &h

	err = h.Execute()
	if err != nil {
		panic(fmt.Sprintf("failed to execute plugin handler %s", err.Error()))
	}
}
