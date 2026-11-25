package goluasql

import (
	"strings"

	lua "github.com/yuin/gopher-lua"
)

var exports = map[string]lua.LGFunction{
	"new":    newFn,
	"escape": escapeFn,
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)

	L.SetField(mod, "_DEBUG", lua.LBool(false))
	L.SetField(mod, "_VERSION", lua.LString("0.0.0"))

	registerType(L, CLIENT_TYPENAME, dbMethods)
	registerType(L, TX_TYPENAME, dbTxMethods)
	registerType(L, RESULT_TYPENAME, dbExecResultMethods)

	return 1
}

// pushErr reports a Go error to the Lua caller as (nil, err) and returns the
// number of values pushed, ready to return from a Lua C function.
func pushErr(L *lua.LState, err error) int {
	L.Push(lua.LNil)
	L.Push(lua.LString(err.Error()))
	return 2
}

// registerType binds a userdata type's method table, so Lua methods such as
// c:connect and tx:commit resolve through the metatable __index.
func registerType(L *lua.LState, name string, methods map[string]lua.LGFunction) {
	mt := L.NewTypeMetatable(name)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
}

func newFn(L *lua.LState) int {
	client := &Client{}
	ud := L.NewUserData()
	ud.Value = client
	L.SetMetatable(ud, L.GetTypeMetatable(CLIENT_TYPENAME))
	L.Push(ud)
	return 1
}

// escapeReplacer backslash-escapes the bytes sql.escape handles, so a Lua
// string can be embedded in a DSN or statement without breaking quoting.
var escapeReplacer = strings.NewReplacer(
	"\r", `\r`,
	"\n", `\n`,
	`\`, `\\`,
	"'", `\'`,
	"\"", `\"`,
	"\x1a", `\Z`,
)

func escape(source string) string {
	return escapeReplacer.Replace(source)
}

func escapeFn(L *lua.LState) int {
	s := L.ToString(1)
	L.Push(lua.LString(escape(s)))
	return 1
}
