package goluasql

import (
	"database/sql"

	lua "github.com/yuin/gopher-lua"
)

const (
	RESULT_TYPENAME = "sql{execresult}"
)

type DBExecResult struct {
	Result sql.Result
}

var dbExecResultMethods = map[string]lua.LGFunction{
	"last_insert_id": dbExecResultLastInsertIdMethod,
}

func checkExecResult(L *lua.LState) *DBExecResult {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*DBExecResult); ok {
		return v
	}
	L.ArgError(1, "exec result expected")
	return nil
}

func dbExecResultLastInsertIdMethod(L *lua.LState) int {
	tx := checkExecResult(L)
	id, err := tx.Result.LastInsertId()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNumber(id))
	return 1
}
