package goluasql

import (
	"database/sql"
	"reflect"

	"github.com/junhsieh/goexamples/fieldbinding/fieldbinding"
	"github.com/sirupsen/logrus"
	util "github.com/zrs01/gsql/goluasql/util"

	lua "github.com/yuin/gopher-lua"
)

const (
	TX_TYPENAME = "sql{tx}"
)

type DBTx struct {
	Tx     *sql.Tx
	Result *sql.Result
}

var dbTxMethods = map[string]lua.LGFunction{
	"exec":     dbTxExecMethod,
	"commit":   dbTxCommitMethod,
	"rollback": dbTxRollbackMethod,
	"query":    dbTxQueryMethod,
}

func checkTx(L *lua.LState) *DBTx {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*DBTx); ok {
		return v
	}
	L.ArgError(1, "tx expected")
	return nil
}

func dbTxExecMethod(L *lua.LState) int {
	tx := checkTx(L)

	query := L.ToString(2)
	if query == "" {
		L.ArgError(2, "query string required")
		return 0
	}

	top := L.GetTop()
	args := make([]interface{}, 0, top-2)
	for i := 3; i <= top; i++ {
		args = append(args, util.GetValue(L, i))
	}

	logrus.Debugf("binding args: %+v", args)
	logrus.Debugf("query: %s", query)
	result, err := tx.Tx.Exec(query, args...)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	tx.Result = &result

	ud := L.NewUserData()
	ud.Value = tx.Result
	L.SetMetatable(ud, L.GetTypeMetatable(RESULT_TYPENAME))
	L.Push(ud)

	return 1
}

func dbTxCommitMethod(L *lua.LState) int {
	tx := checkTx(L)
	if err := tx.Tx.Commit(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func dbTxRollbackMethod(L *lua.LState) int {
	tx := checkTx(L)
	if err := tx.Tx.Rollback(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func dbTxQueryMethod(L *lua.LState) int {
	tx := checkTx(L)
	query := L.ToString(2)

	if query == "" {
		L.ArgError(2, "query string required")
		return 0
	}

	top := L.GetTop()
	args := make([]interface{}, 0, top-2)
	for i := 3; i <= top; i++ {
		args = append(args, util.GetValue(L, i))
	}

	logrus.Debugf("binding args: %+v", args)
	logrus.Debugf("query: %s", query)
	rows, err := tx.Tx.Query(query, args...)

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer rows.Close()

	fb := fieldbinding.NewFieldBinding()
	cols, err := rows.Columns()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	fb.PutFields(cols)

	tb := L.NewTable()
	for rows.Next() {
		if err := rows.Scan(fb.GetFieldPtrArr()...); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbRow := util.ToTableFromMap(L, reflect.ValueOf(fb.GetFieldArr()))
		tb.Append(tbRow)
	}

	L.Push(tb)
	return 1
}
