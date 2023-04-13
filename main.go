/*
** Zabbix
** Copyright 2001-2023 Zabbix SIA
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
	"flag"
	"fmt"

	"git.zabbix.com/ap/plugin-support/plugin/comms"
	"git.zabbix.com/ap/plugin-support/plugin/container"
	"git.zabbix.com/ap/postgresql/plugin"
)

const PLUGIN_VERSION_MAJOR = 6
const PLUGIN_VERSION_MINOR = 0
const PLUGIN_VERSION_PATCH = 16
const PLUGIN_VERSION_RC = ""
const PLUGIN_LICENSE_YEAR = 2023

func main() {
	handleFlags()

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

func handleFlags() {
	comms.InitCommonFlags()

	flag.Parse()

	comms.VisitCommonFlags(plugin.Name, copyrightMessage(), comms.GetCommonHelpMessage(), PLUGIN_VERSION_MAJOR,
		PLUGIN_VERSION_MINOR, PLUGIN_VERSION_PATCH, PLUGIN_VERSION_RC)
}

func copyrightMessage() string {
	return fmt.Sprintf("Copyright 2001-%d Zabbix SIA\n"+
		"Licensed under the Apache License, Version 2.0 (the \"License\");\n"+
		"you may not use this file except in compliance with the License.\n"+
		"You may obtain a copy of the License at\n\n"+
		"\thttp://www.apache.org/licenses/LICENSE-2.0\n\n"+
		"Unless required by applicable law or agreed to in writing, software\n"+
		"distributed under the License is distributed on an \"AS IS\" BASIS,\n"+
		"WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n"+
		"See the License for the specific language governing permissions and\n"+
		"limitations under the License.\n", PLUGIN_LICENSE_YEAR)
}
