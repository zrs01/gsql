package goluasql

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/samber/lo"
	lua "github.com/yuin/gopher-lua"
	util "github.com/zrs01/gsql/goluasql/util"
)

const (
	CLIENT_TYPENAME = "sql{client}"
)

// Client mysql
type Client struct {
	DB      *sql.DB
	Timeout time.Duration
}

var dbMethods = map[string]lua.LGFunction{
	"connect":       dbConnectMethod,
	"set_timeout":   dbSetTimeoutMethod,
	"set_keepalive": dbSetKeepaliveMethod,
	"close":         dbCloseMethod,
	"query":         dbQueryMethod,
	"begin_tx":      dbBeginTxMethod,
}

func checkClient(L *lua.LState) *Client {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Client); ok {
		return v
	}
	L.ArgError(1, "client expected")
	return nil
}

func dbCloseMethod(L *lua.LState) int {
	client := checkClient(L)

	if client.DB == nil {
		L.Push(lua.LBool(true))
		return 1
	}

	err := client.DB.Close()
	// always clean
	client.DB = nil
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LBool(true))
	return 1
}

func dbSetKeepaliveMethod(L *lua.LState) int {
	client := checkClient(L)
	idleTimeout := L.ToInt64(2) // timeout (in ms)
	poolSize := L.ToInt(3)

	if client.DB == nil {
		L.Push(lua.LBool(true))
		L.Push(lua.LString("connect required"))
		return 2
	}

	client.DB.SetConnMaxLifetime(time.Millisecond * time.Duration(idleTimeout))
	client.DB.SetMaxIdleConns(poolSize)

	L.Push(lua.LBool(true))
	return 1
}

// dbBeginTxMethod is a Go function that starts a new transaction for a client.
//
// It takes in a Lua state as a parameter.
// It returns an integer indicating the number of values pushed to the Lua stack.
func dbBeginTxMethod(L *lua.LState) int {
	client := checkClient(L)
	tx := &DBTx{}
	clientTx, err := client.DB.BeginTx(context.Background(), nil)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	tx.Tx = clientTx
	ud := L.NewUserData()
	ud.Value = tx
	L.SetMetatable(ud, L.GetTypeMetatable(TX_TYPENAME))
	L.Push(ud)
	return 1
}

func dbSetTimeoutMethod(L *lua.LState) int {
	client := checkClient(L)
	timeout := L.ToInt64(2) // timeout (in ms)

	client.Timeout = time.Millisecond * time.Duration(timeout)
	return 0
}

func dbConnectMethod(L *lua.LState) int {
	client := checkClient(L)
	driverName := util.GetValue(L, 2)
	dsn := util.GetValue(L, 3)

	if lo.IsNil(driverName) {
		L.ArgError(1, "driver name required")
		return 0
	}
	if lo.IsNil(dsn) {
		L.ArgError(2, "data source name required")
		return 0
	}

	db, err := sql.Open(driverName.(string), dsn.(string))
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(err.Error()))
		return 2
	}
	client.DB = db

	L.Push(lua.LBool(true))
	return 1
}
