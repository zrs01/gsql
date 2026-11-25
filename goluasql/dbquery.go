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
	args := make([]any, 0, top-2)
	for i := 3; i <= top; i++ {
		args = append(args, util.GetValue(L, i))
	}

	logrus.Debugf("binding args: %+v", args)
	logrus.Debugf("query: %s", query)
	rows, err := client.DB.Query(query, args...)

	if err != nil {
		return pushErr(L, err)
	}
	defer rows.Close()

	fb := fieldbinding.NewFieldBinding()
	cols, err := rows.Columns()
	if err != nil {
		return pushErr(L, err)
	}

	fb.PutFields(cols)

	tb := L.NewTable()
	for rows.Next() {
		if err := rows.Scan(fb.GetFieldPtrArr()...); err != nil {
			return pushErr(L, err)
		}

		tbRow := util.ToTableFromMap(L, reflect.ValueOf(fb.GetFieldArr()))
		tb.Append(tbRow)
	}

	if err := rows.Err(); err != nil {
		return pushErr(L, err)
	}

	L.Push(tb)
	return 1
}
