package luaplugin

import (
	"html"

	lua "github.com/yuin/gopher-lua"
)

func WithUtilModule() LuaModule {
	return func(L *lua.LState) {
		L.PreloadModule("util", UtilModuleLoder)
	}
}

func UtilModuleLoder(L *lua.LState) int {
	utilModFuncs := map[string]lua.LGFunction{
		"htmlUnescape": htmlUnescape,
		"htmlEscape":   htmlEscape,
	}
	mod := L.SetFuncs(L.NewTable(), utilModFuncs)
	L.Push(mod)
	return 1
}

func htmlUnescape(L *lua.LState) int {
	s := L.CheckString(1)
	s = html.UnescapeString(s)
	L.Push(lua.LString(s))
	return 1
}

func htmlEscape(L *lua.LState) int {
	s := L.CheckString(1)
	s = html.EscapeString(s)
	L.Push(lua.LString(s))
	return 1
}
