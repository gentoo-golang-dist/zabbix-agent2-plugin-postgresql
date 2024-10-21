/*
** Copyright (C) 2001-2024 Zabbix SIA
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

package plugin

import (
	"golang.zabbix.com/sdk/conf"
)

// Session struct holds individual options for PostgreSQL connection for each session.
type Session struct {
	// URI is a connection string consisting of a network scheme, a host address and a port or a path to a Unix-socket.
	URI string `conf:"name=Uri,optional"`

	// User of PostgreSQL server.
	User string `conf:"optional"`

	// Password to send to protected PostgreSQL server.
	Password string `conf:"optional"`

	// Database of PostgreSQL server.
	Database string `conf:"optional"`

	// Connection type of PostgreSQL server.
	TLSConnect string `conf:"name=TLSConnect,optional"`

	// Certificate Authority filepath for PostgreSQL server.
	TLSCAFile string `conf:"name=TLSCAFile,optional"`

	// Certificate filepath for PostgreSQL server.
	TLSCertFile string `conf:"name=TLSCertFile,optional"`

	// Key filepath for PostgreSQL server.
	TLSKeyFile string `conf:"name=TLSKeyFile,optional"`

	// CacheMode for PostgreSQL server.
	CacheMode string `conf:"name=CacheMode,optional"`
}

// Validate implements the Configurator interface.
// Returns an error if validation of a plugin's configuration is failed.
func (p *Plugin) Validate(options interface{}) error {
	var opts PluginOptions

	return conf.Unmarshal(options, &opts)
}
