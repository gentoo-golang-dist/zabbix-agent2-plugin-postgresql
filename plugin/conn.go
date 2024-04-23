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

package plugin

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/omeid/go-yarn"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/tlsconfig"
	"golang.zabbix.com/sdk/uri"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	// pgx dns field names
	password  = "password"
	sslMode   = "sslmode"
	rootCA    = "sslrootcert"
	cert      = "sslcert"
	key       = "sslkey"
	cacheMode = "statement_cache_mode"

	// connType
	disable    = "disable"
	require    = "require"
	verifyCa   = "verify-ca"
	verifyFull = "verify-full"

	MinSupportedPGVersion = 100000
)

type PostgresClient interface {
	Query(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error)
	QueryByName(ctx context.Context, queryName string, args ...interface{}) (rows *sql.Rows, err error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (row *sql.Row, err error)
	QueryRowByName(ctx context.Context, queryName string, args ...interface{}) (row *sql.Row, err error)
	PostgresVersion() int
}

// PGConn holds pointer to the Pool of PostgreSQL Instance.
type PGConn struct {
	client         *sql.DB
	callTimeout    time.Duration
	ctx            context.Context
	lastTimeAccess time.Time
	version        int
	queryStorage   *yarn.Yarn
	address        string
}

type connID struct {
	uri       uri.URI
	cacheMode string
}

var errorQueryNotFound = "query %q not found"

// Query wraps pgxpool.Query.
func (conn *PGConn) Query(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	rows, err = conn.client.QueryContext(ctx, query, args...)

	if ctxErr := ctx.Err(); ctxErr != nil {
		err = ctxErr
	}

	return
}

// QueryByName executes a query from queryStorage by its name and returns a single row.
func (conn *PGConn) QueryByName(ctx context.Context, queryName string, args ...interface{}) (rows *sql.Rows, err error) {
	if sql, ok := (*conn.queryStorage).Get(queryName + sqlExt); ok {
		normalizedSQL := strings.TrimRight(strings.TrimSpace(sql), ";")

		return conn.Query(ctx, normalizedSQL, args...)
	}

	return nil, fmt.Errorf(errorQueryNotFound, queryName)
}

// QueryRow wraps pgxpool.QueryRow.
func (conn *PGConn) QueryRow(ctx context.Context, query string, args ...interface{}) (row *sql.Row, err error) {
	row = conn.client.QueryRowContext(ctx, query, args...)

	if ctxErr := ctx.Err(); ctxErr != nil {
		err = ctxErr
	}

	return
}

// QueryRowByName executes a query from queryStorage by its name and returns a single row.
func (conn *PGConn) QueryRowByName(ctx context.Context, queryName string, args ...interface{}) (row *sql.Row, err error) {
	if sql, ok := (*conn.queryStorage).Get(queryName + sqlExt); ok {
		normalizedSQL := strings.TrimRight(strings.TrimSpace(sql), ";")

		return conn.QueryRow(ctx, normalizedSQL, args...)
	}

	return nil, fmt.Errorf(errorQueryNotFound, queryName)
}

// GetPostgresVersion exec SQL query to retrieve the version of PostgreSQL server we are currently connected to.
func getPostgresVersion(ctx context.Context, conn *sql.DB) (version int, err error) {
	err = conn.QueryRowContext(ctx, `select current_setting('server_version_num');`).Scan(&version)

	return
}

// PostgresVersion returns the version of PostgreSQL server we are currently connected to.
func (conn *PGConn) PostgresVersion() int {
	return conn.version
}

// updateAccessTime updates the last time a connection was accessed.
func (conn *PGConn) updateAccessTime() {
	conn.lastTimeAccess = time.Now()
}

// ConnManager is a thread-safe structure for manage connections.
type ConnManager struct {
	sync.Mutex
	connMutex      sync.Mutex
	connections    map[connID]*PGConn
	keepAlive      time.Duration
	connectTimeout time.Duration
	callTimeout    time.Duration
	Destroy        context.CancelFunc
	queryStorage   yarn.Yarn
}

// NewConnManager initializes connManager structure and runs Go Routine that watches for unused connections.
func NewConnManager(keepAlive, connectTimeout, callTimeout,
	hkInterval time.Duration, queryStorage yarn.Yarn) *ConnManager {
	ctx, cancel := context.WithCancel(context.Background())

	connMgr := &ConnManager{
		connections:    make(map[connID]*PGConn),
		keepAlive:      keepAlive,
		connectTimeout: connectTimeout,
		callTimeout:    callTimeout,
		Destroy:        cancel, // Destroy stops originated goroutines and closes connections.
		queryStorage:   queryStorage,
	}

	go connMgr.housekeeper(ctx, hkInterval)

	return connMgr
}

// closeUnused closes each connection that has not been accessed at least within the keepalive interval.
func (c *ConnManager) closeUnused() {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	for ci, conn := range c.connections {
		if time.Since(conn.lastTimeAccess) > c.keepAlive {
			conn.client.Close()
			delete(c.connections, ci)
			Impl.Debugf("[%s] Closed unused connection: %s", Name, ci.uri.Addr())
		}
	}
}

// closeAll closes all existed connections.
func (c *ConnManager) closeAll() {
	c.connMutex.Lock()
	for ci, conn := range c.connections {
		conn.client.Close()
		delete(c.connections, ci)
	}
	c.connMutex.Unlock()
}

// housekeeper repeatedly checks for unused connections and closes them.
func (c *ConnManager) housekeeper(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			c.closeAll()

			return
		case <-ticker.C:
			c.closeUnused()
		}
	}
}

// create creates a new connection with given credentials.
func (c *ConnManager) create(ci connID, details tlsconfig.Details) (*PGConn, error) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if _, ok := c.connections[ci]; ok {
		// Should never happen.
		panic("connection already exists")
	}

	ctx := context.Background()

	host := ci.uri.Host()
	port := ci.uri.Port()

	if ci.uri.Scheme() == "unix" {
		socket := ci.uri.Addr()
		host = filepath.Dir(socket)

		ext := filepath.Ext(filepath.Base(socket))
		if len(ext) <= 1 {
			return nil, fmt.Errorf("incorrect socket: %q", socket)
		}

		port = ext[1:]
	}

	dbname, err := url.QueryUnescape(ci.uri.GetParam("dbname"))
	if err != nil {
		return nil, err
	}

	client, err := createClient(
		createDNS(host, port, dbname, ci.uri.User(), ci.uri.Password(), ci.cacheMode, details), c.connectTimeout,
	)
	if err != nil {
		return nil, err
	}

	serverVersion, err := getPostgresVersion(ctx, client)
	if err != nil {
		client.Close()
		return nil, err
	}

	if serverVersion < MinSupportedPGVersion {
		client.Close()
		return nil, fmt.Errorf("PostgreSQL version %d is not supported", serverVersion)
	}

	c.connections[ci] = &PGConn{
		client:         client,
		callTimeout:    c.callTimeout,
		version:        serverVersion,
		lastTimeAccess: time.Now(),
		ctx:            ctx,
		queryStorage:   &c.queryStorage,
		address:        ci.uri.Addr(),
	}

	Impl.Debugf("[%s] Created new connection: %s", Name, ci.uri.Addr())

	return c.connections[ci], nil
}

func createDNS(host, port, dbname, user, pass, mode string, details tlsconfig.Details) string {
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s", host, port, dbname, user)

	tmp := map[string]string{
		password:  pass,
		sslMode:   details.TlsConnect,
		rootCA:    details.TlsCaFile,
		cert:      details.TlsCertFile,
		key:       details.TlsKeyFile,
		cacheMode: mode,
	}

	for k, v := range tmp {
		if v != "" {
			dsn = fmt.Sprintf("%s %s=%s", dsn, k, v)
		}
	}

	return dsn
}

func renameTLS(in string) string {
	switch in {
	case "required":
		return "require"
	case "verify_ca":
		return "verify-ca"
	case "verify_full":
		return "verify-full"
	default:
		return in
	}
}

func createClient(dsn string, timeout time.Duration) (*sql.DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		d := net.Dialer{}
		ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		conn, err := d.DialContext(ctxTimeout, network, addr)

		return conn, err
	}

	return stdlib.OpenDB(*config.ConnConfig), nil
}

// get returns a connection with given uri if it exists and also updates lastTimeAccess, otherwise returns nil.
func (c *ConnManager) get(cd connID) *PGConn {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if conn, ok := c.connections[cd]; ok {
		conn.updateAccessTime()
		return conn
	}

	return nil
}

// GetConnection returns an existing connection or creates a new one.
func (c *ConnManager) GetConnection(ci connID, params map[string]string) (conn *PGConn, err error) {
	c.Lock()
	defer c.Unlock()

	conn = c.get(ci)
	if conn != nil {
		return
	}

	details, err := getTlsDetails(params)
	if err != nil {
		return nil, err
	}

	conn, err = c.create(ci, details)
	if err != nil {
		err = zbxerr.ErrorConnectionFailed.Wrap(err)
	}

	return
}

func getTlsDetails(params map[string]string) (tlsconfig.Details, error) {
	tlsType := renameTLS(params[tlsConnectParam])
	validateCA := true

	if tlsType == "" {
		tlsType = disable
	}

	details := tlsconfig.NewDetails(
		params[metric.SessionParam],
		tlsType,
		params[tlsCAParam],
		params[tlsCertParam],
		params[tlsKeyParam],
		params[uriParam],
		disable,
		require,
		verifyCa,
		verifyFull,
	)

	if tlsType == disable || tlsType == require {
		validateCA = false
	}

	err := details.Validate(validateCA, false, false)
	return details, err
}

func createConnID(params map[string]string) (connID, error) {
	u, err := uri.NewWithCreds(
		fmt.Sprintf("%s?dbname=%s", params[uriParam], url.QueryEscape(params[databaseParam])),
		params[userParam],
		params[passwordParam],
		uriDefaults,
	)
	if err != nil {
		return connID{}, err
	}

	return connID{uri: *u, cacheMode: params[cacheModeParam]}, nil
}
