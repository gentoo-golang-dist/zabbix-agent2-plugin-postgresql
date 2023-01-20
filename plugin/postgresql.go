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
	"net/http"
	"net/url"
	"time"

	"git.zabbix.com/ap/plugin-support/tlsconfig"
	"git.zabbix.com/ap/plugin-support/uri"
	"git.zabbix.com/ap/plugin-support/zbxerr"

	"github.com/omeid/go-yarn"

	"git.zabbix.com/ap/plugin-support/plugin"
)

const (
	Name       = "PostgreSQL"
	sqlExt     = ".sql"
	hkInterval = 10
)

// Plugin inherits plugin.Base and store plugin-specific data.
type Plugin struct {
	plugin.Base
	connMgr *ConnManager
	options PluginOptions
}

// Impl is the pointer to the plugin implementation.
var Impl Plugin

// Export implements the Exporter interface.
func (p *Plugin) Export(key string, rawParams []string, _ plugin.ContextProvider) (result interface{}, err error) {
	params, extraParams, err := metrics[key].EvalParams(rawParams, p.options.Sessions)
	if err != nil {
		return nil, err
	}

	details, err := tlsconfig.CreateDetails(params["sessionName"], params["TLSConnect"],
		params["TLSCAFile"], params["TLSCertFile"], params["TLSKeyFile"], params["URI"])
	if err != nil {
		return nil, zbxerr.ErrorInvalidConfiguration.Wrap(err)
	}

	dbname := url.QueryEscape(params["Database"])

	uri, err := uri.NewWithCreds(params["URI"]+"?dbname="+dbname, params["User"], params["Password"], uriDefaults)
	if err != nil {
		return nil, err
	}

	handleMetric := getHandlerFunc(key)
	if handleMetric == nil {
		return nil, zbxerr.ErrorUnsupportedMetric
	}

	conn, err := p.connMgr.GetConnection(*uri, details)
	if err != nil {
		// Special logic of processing connection errors should be used if pgsql.ping is requested
		// because it must return pingFailed if any error occurred.
		if key == keyPing {
			return pingFailed, nil
		}

		p.Errf(err.Error())

		return nil, err
	}

	ctx, cancel := context.WithTimeout(conn.ctx, conn.callTimeout)
	defer cancel()

	result, err = handleMetric(ctx, conn, key, params, extraParams...)

	if err != nil {
		p.Errf(err.Error())
	}

	return result, err
}

// Start implements the Runner interface and performs initialization when plugin is activated.
func (p *Plugin) Start() {
	p.connMgr = NewConnManager(
		time.Duration(p.options.KeepAlive)*time.Second,
		time.Duration(p.options.Timeout)*time.Second,
		time.Duration(p.options.CallTimeout)*time.Second,
		hkInterval*time.Second,
		p.setCustomQuery(),
	)
}

func (p *Plugin) setCustomQuery() yarn.Yarn {
	if p.options.CustomQueriesPath == "" {
		return yarn.NewFromMap(map[string]string{})
	}

	queryStorage, err := yarn.New(http.Dir(p.options.CustomQueriesPath), "*"+sqlExt)
	if err != nil {
		p.Errf(err.Error())
		// create empty storage if error occurred
		return yarn.NewFromMap(map[string]string{})
	}

	return queryStorage
}

// Stop implements the Runner interface and frees resources when plugin is deactivated.
func (p *Plugin) Stop() {
	p.connMgr.Destroy()
	p.connMgr = nil
}
