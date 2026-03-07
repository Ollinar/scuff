package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/Ollinar/scuff/internal/model"
	"github.com/Ollinar/scuff/internal/plugin"
)

func (ap App) Plugin() pluginModule {
	return pluginModule{app: ap}
}

type pluginModule struct {
	app App
}

func (plugMod pluginModule) Add(script string) error {
	inf, err := plugMod.app.pluginProvider.Validate(script)
	if err != nil {
		return errors.Join(ErrInvalidEntity, err)
	}
	savePath := filepath.Join(plugMod.app.pluginDir, plugMod.app.pluginProvider.FileName(inf))
	err = os.MkdirAll(plugMod.app.pluginDir, os.ModePerm)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	f, err := os.Create(savePath)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	defer f.Close()
	_, err = f.WriteString(script)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	return nil
}

func (plugMod pluginModule) GetAll(ctx context.Context) ([]plugin.PluginInfo, error) {
	dirEntries, err := os.ReadDir(plugMod.app.pluginDir)
	if err != nil {
		return nil, errors.Join(ErrUnexpected, err)
	}

	plugInfs := make([]plugin.PluginInfo, 0, len(dirEntries))

	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		f, err := os.Open(filepath.Join(plugMod.app.pluginDir, de.Name()))
		if err != nil {
			return nil, errors.Join(ErrUnexpected, err)
		}
		cont, err := io.ReadAll(f)
		if err != nil {
			return nil, errors.Join(ErrUnexpected, err)
		}
		f.Close()
		pi, err := plugMod.app.pluginProvider.Validate(string(cont))
		if err != nil {
			// NOTE: ignore the file if the plugin provider could not validate it even if the err is not about the script itself
			continue
		}
		savedConf, err := plugMod.app.pluginRepo.GetConfig(ctx, pi.Name, pi.Version)
		if err != nil {
			return nil, errors.Join(ErrUnexpected, err)
		}

		for i, v := range pi.Config {
			sv, ok := savedConf[v.Name]
			if ok {
				pi.Config[i].Value = sv
			}
		}
		dly, ok := savedConf["delay"]
		if ok {
			dur, err := time.ParseDuration(dly)
			if err != nil {
				return nil, errors.Join(ErrUnexpected, err)
			}
			pi.Delay = dur
		}
		autoRunStr, ok := savedConf["autorun"]
		if ok {
			b, err := strconv.ParseBool(autoRunStr)
			if err != nil {
				return nil, errors.Join(ErrUnexpected, err)
			}
			pi.AutoRun = b
		}

		plugInfs = append(plugInfs, pi)

	}

	return plugInfs, nil
}

func (plugMod pluginModule) SetConfig(ctx context.Context, name, version string, autoRun bool, delay time.Duration, config map[string]string) error {
	config["delay"] = delay.String()
	config["autorun"] = strconv.FormatBool(autoRun)
	err := plugMod.app.pluginRepo.StoreConfig(ctx, name, version, config)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	return nil
}

func (plugMod pluginModule) Exec(ctx context.Context, plugInf plugin.PluginInfo, param map[string]string, targetId model.ID) error {
	plugFile, err := os.Open(filepath.Join(plugMod.app.pluginDir, plugMod.app.pluginProvider.FileName(plugInf)))
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Join(ErrNotFound, fmt.Errorf("plugin %s not found", plugInf.Name))
		}
		return errors.Join(ErrUnexpected, err)
	}
	defer plugFile.Close()
	cont, err := io.ReadAll(plugFile)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	conf, err := plugMod.app.pluginRepo.GetConfig(ctx, plugInf.Name, plugInf.Version)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	plugMod.app.logger.Debug("executing plugin", slog.String("plugin name", plugInf.Name),
		slog.Duration("delay", plugInf.Delay),
	)
	err = plugMod.app.pluginProvider.Execute(ctx, string(cont), conf, param, targetId)
	if err != nil {
		return err
	}

	return nil
}

func (plugMod pluginModule) LoadAutoRuns(ctx context.Context, targetType plugin.Target) ([]plugin.Plugin, error) {
	plugs, err := plugMod.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	plugs = slices.DeleteFunc(plugs, func(p plugin.PluginInfo) bool {
		return !p.AutoRun || p.TargetEntity != targetType
	})

	loadedPlugins := make([]plugin.Plugin, 0, len(plugs))
	errored := false
	defer func() {
		if errored {
			for _, plug := range loadedPlugins {
				plug.Close()
			}
		}
	}()

	for _, plugInf := range plugs {
		plugFile, err := os.Open(filepath.Join(plugMod.app.pluginDir, plugMod.app.pluginProvider.FileName(plugInf)))
		if err != nil {
			errored = true
			if os.IsNotExist(err) {
				return nil, errors.Join(ErrNotFound, fmt.Errorf("plugin %s not found", plugInf.Name))
			}
			return nil, errors.Join(ErrUnexpected, err)
		}
		cont, err := io.ReadAll(plugFile)
		if err != nil {
			errored = true
			plugFile.Close()
			return nil, errors.Join(ErrUnexpected, err)
		}
		plugFile.Close()
		conf := make(map[string]string, len(plugInf.Config))
		for _, c := range plugInf.Config {
			conf[c.Name] = c.Value
		}

		plug, err := plugMod.app.pluginProvider.Load(ctx, string(cont), conf)
		if err != nil {
			errored = true
			return nil, err
		}
		loadedPlugins = append(loadedPlugins, plug)

	}

	return loadedPlugins, nil
}
