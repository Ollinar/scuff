package luaplugin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/Ollinar/scuff/internal/model"
	"github.com/Ollinar/scuff/internal/plugin"

	lua "github.com/yuin/gopher-lua"
)

type OperationType string

const UpdateChapter OperationType = "UpdateChapter"

type moduleInfo struct {
	plugin.PluginInfo
	runFunc *lua.LFunction
}
type LuaModule func(L *lua.LState)

type LuaPlugin struct {
	httpClient *http.Client
	modules    []LuaModule
}

func NewLuaPlugin(opts ...LuaModule) *LuaPlugin {
	lp := LuaPlugin{
		modules: opts,
	}

	return &lp
}

func (luaPlug *LuaPlugin) AddModule(mods ...LuaModule) {
	luaPlug.modules = append(luaPlug.modules, mods...)
}

func (luaPlug LuaPlugin) Validate(script string) (plugin.PluginInfo, error) {
	L := luaPlug.newVM()
	defer L.Close()
	chunk, err := L.LoadString(script)
	if err != nil {
		return plugin.PluginInfo{}, err
	}

	err = L.CallByParam(lua.P{
		Fn:      chunk,
		NRet:    1,
		Protect: true,
	})
	if err != nil {
		return plugin.PluginInfo{}, err
	}

	modInf, err := pluginParseModule(L.Get(-1))
	L.Pop(1)
	if err != nil {
		return plugin.PluginInfo{}, err
	}

	return modInf.PluginInfo, nil
}

func (luaPlug LuaPlugin) Load(script string) (plugin.Plugin, error) {
	L := luaPlug.newVM()
	err := L.DoString(script)
	if err != nil {
		return nil, err
	}
	modV := L.Get(-1)
	L.Pop(1)
	mod, err := pluginParseModule(modV)
	if err != nil {
		return nil, err
	}

	pI := plugImpl{
		mu:     &sync.Mutex{},
		L:      L,
		module: mod,
	}

	return pI, nil
}

func (luaPlug LuaPlugin) FileName(pi plugin.PluginInfo) string {
	return fmt.Sprintf("%s.%s.lua", pi.Name, pi.Version)
}

func (luaPlug LuaPlugin) newVM() *lua.LState {
	// TODO: make the plugin env more isolated
	L := lua.NewState()
	L.PreloadModule("types", func(L *lua.LState) int {
		L.Push(loadTypes(L))
		return 1
	})
	for _, fn := range luaPlug.modules {
		fn(L)
	}
	return L
}

type plugImpl struct {
	mu     *sync.Mutex
	L      *lua.LState
	module moduleInfo
}

func (plugI plugImpl) Close() error {
	plugI.mu.Lock()
	defer plugI.mu.Unlock()
	plugI.L.Close()
	return nil
}

func (plugI plugImpl) Run(ctx context.Context, config, param map[string]string, id model.ID) error {
	plugI.mu.Lock()
	defer plugI.mu.Unlock()
	L := plugI.L
	L.SetContext(ctx)
	defer L.RemoveContext()

	plugConf, plugParam := toPluginArgs(L, plugI.module.PluginInfo, config, param)
	err := L.CallByParam(lua.P{
		Fn:      plugI.module.runFunc,
		NRet:    1,
		Protect: true,
	}, lua.LNumber(id), plugConf, plugParam)
	if err != nil {
		return err
	}
	lErr := L.Get(-1)
	if lErr.Type() != lua.LTNil {
		return errors.New(lErr.String())
	}

	return nil
}
