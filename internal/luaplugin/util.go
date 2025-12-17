package luaplugin

import (
	"errors"
	"strings"

	"github.com/Ollinar/scuff/internal/plugin"

	lua "github.com/yuin/gopher-lua"
)

func SandboxLFunc(L *lua.LState, lf *lua.LFunction) {
	envT := L.NewTable()
	envMT := L.NewTable()
	envMT.RawSetString("__index", L.GetGlobal("_G"))
	L.SetMetatable(envT, envMT)
	L.SetFEnv(lf, envT)
}

func pluginParseModule(lv lua.LValue) (moduleInfo, error) {
	modInf := moduleInfo{}
	mod, ok := lv.(*lua.LTable)
	if !ok {
		return modInf, errors.New("not a valid module")
	}

	infoTbl, ok := mod.RawGetString("info").(*lua.LTable)
	if !ok {
		return modInf, errors.New(`expected "info" to be a table`)
	}
	if nameV := lua.LVAsString(infoTbl.RawGetString("name")); nameV != "" {
		modInf.Name = nameV
	} else {
		return modInf, errors.New(`expected "name" to be non empty string`)
	}

	targetType := lua.LVAsString(infoTbl.RawGetString("target"))
	switch strings.ToLower(targetType) {
	case "archive":
		modInf.TargetEntity = plugin.TargetArchive
	case "chapter":
		modInf.TargetEntity = plugin.TargetChapter
	default:
		modInf.TargetEntity = plugin.TargetNone
	}

	modInf.Description = lua.LVAsString(infoTbl.RawGetString("description"))

	if verV := lua.LVAsString(infoTbl.RawGetString("version")); verV != "" {
		modInf.Version = verV
	} else {
		return modInf, errors.New(`expected "version" to be non empty string`)
	}

	var err error
	configTbl, ok := infoTbl.RawGetString("config").(*lua.LTable)
	if !ok && infoTbl.RawGetString("config").Type() != lua.LTNil {
		return modInf, errors.New(`expected "config" to be a table`)
	}

	if ok {
		configTbl.ForEach(func(k, conf lua.LValue) {
			if err != nil {
				return
			}
			nameV := lua.LVAsString(k)
			if nameV == "" || k.Type() == lua.LTNumber {
				err = errors.New(`expected config table to have a key of non-empty string`)
				return
			}
			confT, ok := conf.(*lua.LTable)
			if !ok {
				err = errors.New(`expected config table to have a value of table`)
				return
			}

			confParsed := plugin.PluginConfig{
				Name:        nameV,
				Description: lua.LVAsString(confT.RawGetString("description")),
				Value:       lua.LVAsString(confT.RawGetString("default")),
			}
			modInf.Config = append(modInf.Config, confParsed)
		})
	}
	if err != nil {
		return modInf, err
	}

	paramTbl, ok := infoTbl.RawGetString("param").(*lua.LTable)
	if !ok && infoTbl.RawGetString("param").Type() != lua.LTNil {
		return modInf, errors.New(`expected "param" to be a table`)
	}
	if ok {
		paramTbl.ForEach(func(k, conf lua.LValue) {
			if err != nil {
				return
			}
			nameV := lua.LVAsString(k)
			if nameV == "" || k.Type() == lua.LTNumber {
				err = errors.New(`expected param table to have a key of non-empty string`)
				return
			}
			paramT, ok := conf.(*lua.LTable)
			if !ok {
				err = errors.New("expected param table to have a value of table")
				return
			}

			paramParsed := plugin.PluginParam{
				Name:        nameV,
				Description: lua.LVAsString(paramT.RawGetString("description")),
				Value:       lua.LVAsString(paramT.RawGetString("default")),
			}
			modInf.Param = append(modInf.Param, paramParsed)
		})
	}
	if err != nil {
		return modInf, err
	}

	modInf.runFunc, ok = mod.RawGetString("Run").(*lua.LFunction)
	if !ok {
		return modInf, errors.New(`expected a "Run" function`)
	}

	return modInf, nil
}

func toPluginArgs(L *lua.LState, plugInf plugin.PluginInfo, config, param map[string]string) (luaConfig, luaParam *lua.LTable) {
	luaConfig = L.NewTable()

	for _, v := range plugInf.Config {
		luaConfig.RawSetString(v.Name, lua.LString(v.Value))
	}
	for k, v := range config {
		luaConfig.RawSetString(k, lua.LString(v))
	}

	luaParam = L.NewTable()
	for _, v := range plugInf.Param {
		luaParam.RawSetString(v.Name, lua.LString(v.Value))
	}
	for k, v := range param {
		luaParam.RawSetString(k, lua.LString(v))
	}
	return
}
