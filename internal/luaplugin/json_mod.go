// Package luaplugin defines the plugin system with lua
package luaplugin

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"
)

func WithJSONModule() LuaModule {
	return func(L *lua.LState) {
		L.PreloadModule("json", JSONModuleLoder)
	}
}

func JSONModuleLoder(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), jsonModFunc)
	L.Push(mod)
	return 1
}

var jsonModFunc = map[string]lua.LGFunction{
	"decode": jsonDecodeLGFunc,
	"encode": jsonEncodeLGFunc,
}

func jsonDecodeLGFunc(L *lua.LState) int {
	jsonStr := L.CheckString(1)
	var jsnVal any
	err := json.Unmarshal([]byte(jsonStr), &jsnVal)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	lv := jsonToLuaValue(L, jsnVal)
	L.Push(lv)
	L.Push(lua.LNil)

	return 2
}

func jsonEncodeLGFunc(L *lua.LState) int {
	lv := L.CheckAny(1)
	jsnVal := jsonFromLuaValue(L, lv)

	jsnB, err := json.Marshal(jsnVal)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(jsnB))
	L.Push(lua.LNil)
	return 2
}

func jsonToLuaValue(L *lua.LState, v any) lua.LValue {
	switch v := v.(type) {
	case bool:
		return lua.LBool(v)
	case float64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []byte:
		return lua.LString(v)
	case []any:
		arr := L.NewTable()
		for _, av := range v {
			arr.Append(jsonToLuaValue(L, av))
		}
		return arr

	case map[string]any:
		mp := L.NewTable()
		for k, mv := range v {
			mp.RawSetString(k, jsonToLuaValue(L, mv))
		}
		return mp
	default:
		return lua.LNil
	}
}

func jsonFromLuaValue(L *lua.LState, lv lua.LValue) any {
	switch lv := lv.(type) {
	case lua.LBool:
		return lv
	case lua.LNumber:
		return lv
	case lua.LString:
		return lv
	case *lua.LTable:
		arr := make([]any, 0, lv.Len())
		mp := make(map[string]any, 0)
		isarr := true
		lv.ForEach(func(k, v lua.LValue) {
			gV := jsonFromLuaValue(L, v)
			if k.Type() == lua.LTNumber {
				arr = append(arr, gV)
				return
			}
			isarr = false
			mp[k.String()] = gV
		})
		if isarr {
			return arr
		}
		return mp

	default:
		return nil
	}
}
