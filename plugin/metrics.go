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

package plugin

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"git.zabbix.com/ap/plugin-support/metric"
	"git.zabbix.com/ap/plugin-support/plugin"
	"git.zabbix.com/ap/plugin-support/uri"
	"git.zabbix.com/ap/postgresql/plugin/handlers"
)

const (
	keyArchiveSize                     = "pgsql.archive"
	keyAutovacuum                      = "pgsql.autovacuum.count"
	keyBgwriter                        = "pgsql.bgwriter"
	keyCache                           = "pgsql.cache.hit"
	keyConnections                     = "pgsql.connections"
	keyCustomQuery                     = "pgsql.custom.query"
	keyDBStat                          = "pgsql.dbstat"
	keyDBStatSum                       = "pgsql.dbstat.sum"
	keyDatabaseAge                     = "pgsql.db.age"
	keyDatabasesBloating               = "pgsql.db.bloating_tables"
	keyDatabasesDiscovery              = "pgsql.db.discovery"
	keyDatabaseSize                    = "pgsql.db.size"
	keyLocks                           = "pgsql.locks"
	keyOldestXid                       = "pgsql.oldest.xid"
	keyPing                            = "pgsql.ping"
	keyQueries                         = "pgsql.queries"
	keyReplicationCount                = "pgsql.replication.count"
	keyReplicationLagB                 = "pgsql.replication.lag.b"
	keyReplicationLagSec               = "pgsql.replication.lag.sec"
	keyReplicationProcessInfo          = "pgsql.replication.process"
	keyReplicationProcessNameDiscovery = "pgsql.replication.process.discovery"
	keyReplicationRecoveryRole         = "pgsql.replication.recovery_role"
	keyReplicationStatus               = "pgsql.replication.status"
	keyUptime                          = "pgsql.uptime"
	keyWal                             = "pgsql.wal.stat"
)

// handlerFunc defines an interface must be implemented by handlers.
type handlerFunc func(ctx context.Context, conn handlers.PostgresClient, key string,
	params map[string]string, extraParams ...string) (res interface{}, err error)

// getHandlerFunc returns a handlerFunc related to a given key.
func getHandlerFunc(key string) handlerFunc {
	switch key {
	case keyDatabasesDiscovery:
		return handlers.DatabasesDiscoveryHandler
	case keyDatabasesBloating:
		return handlers.DatabasesBloatingHandler
	case keyDatabaseSize:
		return handlers.DatabaseSizeHandler
	case keyDatabaseAge:
		return handlers.DatabaseAgeHandler
	case keyArchiveSize:
		return handlers.ArchiveHandler
	case keyPing:
		return handlers.PingHandler
	case keyConnections:
		return handlers.ConnectionsHandler
	case keyWal:
		return handlers.WalHandler
	case keyAutovacuum:
		return handlers.AutovacuumHandler
	case keyDBStat:
		return handlers.DbStatHandler
	case keyDBStatSum:
		return handlers.DbStatSumHandler
	case keyBgwriter:
		return handlers.BgwriterHandler
	case keyCustomQuery:
		return handlers.CustomQueryHandler
	case keyUptime:
		return handlers.UptimeHandler
	case keyCache:
		return handlers.CacheHandler
	case keyReplicationCount:
		return handlers.ReplicationCountHandler
	case keyReplicationStatus:
		return handlers.ReplicationStatusHandler
	case keyReplicationLagSec:
		return handlers.ReplicationLagSecHandler
	case keyReplicationRecoveryRole:
		return handlers.ReplicationRecoveryRoleHandler
	case keyReplicationLagB:
		return handlers.ReplicationLagBHandler
	case keyReplicationProcessInfo:
		return handlers.ReplicationProcessInfoHandler
	case keyReplicationProcessNameDiscovery:
		return handlers.ProcessNameDiscoveryHandler
	case keyLocks:
		return handlers.LocksHandler
	case keyOldestXid:
		return handlers.OldestXIDHandler
	case keyQueries:
		return handlers.QueriesHandler
	default:
		return nil
	}
}

var uriDefaults = &uri.Defaults{Scheme: "tcp", Port: "5432"}

var (
	minDBNameLen = 1
	maxDBNameLen = 63
	maxPassLen   = 512
)

type PostgresURIValidator struct {
	Defaults       *uri.Defaults
	AllowedSchemes []string
}

var reSocketPath = regexp.MustCompile(`^.*\.s\.PGSQL\.\d{1,5}$`)

func (v PostgresURIValidator) Validate(value *string) error {
	if value == nil {
		return nil
	}

	u, err := uri.New(*value, v.Defaults)
	if err != nil {
		return err
	}

	isValidScheme := false
	if v.AllowedSchemes != nil {
		for _, s := range v.AllowedSchemes {
			if u.Scheme() == s {
				isValidScheme = true
				break
			}
		}

		if !isValidScheme {
			return fmt.Errorf("allowed schemes: %s", strings.Join(v.AllowedSchemes, ", "))
		}
	}

	if u.Scheme() == "unix" && !reSocketPath.MatchString(*value) {
		return errors.New(
			`socket file must satisfy the format: "/path/.s.PGSQL.nnnn" where nnnn is the server's port number`)
	}

	return nil
}

// Common params: [URI|Session][,User][,Password][,Database]
var (
	paramURI = metric.NewConnParam("URI", "URI to connect or session name.").
			WithDefault(uriDefaults.Scheme + "://localhost:" + uriDefaults.Port).WithSession().
			WithValidator(PostgresURIValidator{
			Defaults:       uriDefaults,
			AllowedSchemes: []string{"tcp", "postgresql", "unix"},
		})
	paramUsername = metric.NewConnParam("User", "PostgreSQL user.").WithDefault("postgres")
	paramPassword = metric.NewConnParam("Password", "User's password.").WithDefault("").
			WithValidator(metric.LenValidator{Max: &maxPassLen})
	paramDatabase = metric.NewConnParam("Database", "Database name to be used for connection.").
			WithDefault("postgres").WithValidator(metric.LenValidator{Min: &minDBNameLen, Max: &maxDBNameLen})
	paramTLSConnect  = metric.NewSessionOnlyParam("TLSConnect", "DB connection encryption type.").WithDefault("")
	paramTLSCaFile   = metric.NewSessionOnlyParam("TLSCAFile", "TLS ca file path.").WithDefault("")
	paramTLSCertFile = metric.NewSessionOnlyParam("TLSCertFile", "TLS cert file path.").WithDefault("")
	paramTLSKeyFile  = metric.NewSessionOnlyParam("TLSKeyFile", "TLS key file path.").WithDefault("")
)

var metrics = metric.MetricSet{
	keyArchiveSize: metric.New("Returns info about size of archive files.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyAutovacuum: metric.New("Returns count of autovacuum workers.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyBgwriter: metric.New("Returns JSON for sum of each type of bgwriter statistic.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyCache: metric.New("Returns cache hit percent.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyConnections: metric.New("Returns JSON for sum of each type of connection.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyCustomQuery: metric.New("Returns result of a custom query.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase,
			metric.NewParam("QueryName", "Name of a custom query "+
				"(must be equal to a name of an SQL file without an extension).").SetRequired(),
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, true),

	keyDBStat: metric.New("Returns JSON for sum of each type of statistic.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyDBStatSum: metric.New("Returns JSON for sum of each type of statistic for all database.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyDatabaseAge: metric.New("Returns age for specific database.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyDatabasesBloating: metric.New("Returns percent of bloating tables for each database.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyDatabasesDiscovery: metric.New("Returns JSON discovery rule with names of databases.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyDatabaseSize: metric.New("Returns size in bytes for specific database.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyLocks: metric.New("Returns collect all metrics from pg_locks.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyOldestXid: metric.New("Returns age of oldest xid.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyPing: metric.New("Tests if connection is alive or not.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyQueries: metric.New("Returns queries statistic.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase,
			metric.NewParam("TimePeriod", "Execution time limit for count of slow queries.").SetRequired(),
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyReplicationCount: metric.New("Returns number of standby servers.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyReplicationLagB: metric.New("Returns replication lag with Master in byte.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyReplicationLagSec: metric.New("Returns replication lag with Master in seconds.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyReplicationProcessNameDiscovery: metric.New("Returns JSON with application name from pg_stat_replication.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyReplicationProcessInfo: metric.New("Returns flush lag, write lag and replay lag per each sender process.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyReplicationRecoveryRole: metric.New("Returns postgreSQL recovery role.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyReplicationStatus: metric.New("Returns postgreSQL replication status.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyUptime: metric.New("Returns uptime.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),

	keyWal: metric.New("Returns JSON wal by type.",
		[]*metric.Param{paramURI, paramUsername, paramPassword, paramDatabase, paramTLSConnect,
			paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile}, false),
}

func init() {
	plugin.RegisterMetrics(&Impl, Name, metrics.List()...)
}
