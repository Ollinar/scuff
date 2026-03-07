package luaplugin

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

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
	logger     *slog.Logger
	modules    []LuaModule
}

func NewLuaPlugin(logger *slog.Logger, opts ...LuaModule) *LuaPlugin {
	lp := LuaPlugin{
		modules: opts,
		logger:  logger,
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

// Execute implements [plugin.Provider].
func (luaPlug *LuaPlugin) Execute(ctx context.Context, script string, config map[string]string, param map[string]string, id model.ID) error {

	L := luaPlug.newVM()
	err := L.DoString(script)
	if err != nil {
		return err
	}
	modV := L.Get(-1)
	L.Pop(1)
	mod, err := pluginParseModule(modV)
	if err != nil {
		return err
	}
	L.SetContext(ctx)
	plugConf, plugParam := toPluginArgs(L, mod.PluginInfo, config, nil)
	err = L.CallByParam(lua.P{
		Fn:      mod.runFunc,
		NRet:    1,
		Protect: true,
	}, lua.LNumber(id), plugConf, plugParam)
	if err != nil {
		return err
	}
	lErr := L.Get(-1)
	if lErr.Type() != lua.LTNil {
		return err
	}
	return nil
}

func (luaPlug LuaPlugin) Load(ctx context.Context, script string, config map[string]string) (plugin.Plugin, error) {
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
	pI := newPluginImpl(ctx, L, mod, config, nil)

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
	L *lua.LState

	addIdChan chan model.ID
}

func newPluginImpl(ctx context.Context, L *lua.LState, mod moduleInfo, config map[string]string, logger *slog.Logger) plugImpl {
	pi := plugImpl{
		L:         L,
		addIdChan: make(chan model.ID),
	}

	workerChan := make(chan model.ID)

	go func() {
		var ids []model.ID
		// nextID is needed because the (case conditionalChan <- nextID:) will out of bounds
		var nextID model.ID
		var inputChan <-chan model.ID = pi.addIdChan
		for {
			// chan will stay nil if ids is empty so it will lock the case clause
			var conditionalChan chan<- model.ID

			if len(ids) > 0 {
				conditionalChan = workerChan
				nextID = ids[0]
			} else if inputChan == nil && len(ids) == 0 {
				close(workerChan)
				return
			}

			select {
			case <-ctx.Done():
				close(workerChan)
				return
			case id, ok := <-inputChan:
				if !ok {
					inputChan = nil
					continue
				}
				ids = append(ids, id)
			case conditionalChan <- nextID:
				ids = ids[1:]
				if len(ids) == 0 {
					ids = nil
				}

			}
		}

	}()

	go func() {
		for id := range workerChan {
			L := pi.L
			L.SetContext(ctx)
			plugConf, plugParam := toPluginArgs(L, mod.PluginInfo, config, nil)
			err := L.CallByParam(lua.P{
				Fn:      mod.runFunc,
				NRet:    1,
				Protect: true,
			}, lua.LNumber(id), plugConf, plugParam)
			if err != nil {
				logger.Error("failed to execute plugin", slog.Any("error", err))
				continue
			}
			lErr := L.Get(-1)
			if lErr.Type() != lua.LTNil {
				logger.Error("plugin returned an error", slog.Any("error", err))
				continue
			}
		}
	}()

	return pi
}

// QueueUp implements [plugin.Plugin].
func (plugI plugImpl) QueueUp(id model.ID) {
	plugI.addIdChan <- id
}

func (plugI plugImpl) Close() error {
	plugI.L.Close()
	return nil
}
