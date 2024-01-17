package goluasql

import (
	"reflect"

	"github.com/sirupsen/logrus"
	util "github.com/zrs01/gsql/goluasql/util"

	"github.com/junhsieh/goexamples/fieldbinding/fieldbinding"
	lua "github.com/yuin/gopher-lua"
)

func dbQueryMethod(L *lua.LState) int {
	client := checkClient(L)
	query := L.ToString(2)

	if client.DB == nil {
		return 0
	}

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
	rows, err := client.DB.Query(query, args...)

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
