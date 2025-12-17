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
	"sync"
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

func (plugMod pluginModule) Run(ctx context.Context, plugInf plugin.PluginInfo, param map[string]string, targetIds ...model.ID) error {
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

	plug, err := plugMod.app.pluginProvider.Load(string(cont))
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}
	defer plug.Close()

	conf, err := plugMod.app.pluginRepo.GetConfig(ctx, plugInf.Name, plugInf.Version)
	if err != nil {
		return errors.Join(ErrUnexpected, err)
	}

	plugMod.app.logger.Debug("executing plugin", slog.String("plugin name", plugInf.Name),
		slog.Duration("delay", plugInf.Delay),
	)
	tkr := time.NewTicker(max(plugInf.Delay, 1))
	defer tkr.Stop()
	for _, id := range targetIds {
		<-tkr.C
		err = plug.Run(ctx, conf, param, id)
		if err != nil {
			return errors.Join(ErrUnexpected, err)
		}

	}

	return nil
}

func (plugMod pluginModule) RunAutoRuns(ctx context.Context, targetType plugin.Target, targetIds ...model.ID) error {
	plugs, err := plugMod.GetAll(ctx)
	if err != nil {
		return err
	}
	plugs = slices.DeleteFunc(plugs, func(p plugin.PluginInfo) bool {
		return !p.AutoRun || p.TargetEntity != targetType
	})

	for _, pi := range plugs {
		err = plugMod.Run(ctx, pi, map[string]string{}, targetIds...)
		if err != nil {
			return err
		}
	}

	return nil
}

type PluginQueue struct {
	pluginInfos   []plugin.PluginInfo
	plugins       []plugin.Plugin
	delayTickers  []*time.Ticker
	pluginConfigs map[int]map[string]string
	iDChanel      chan model.ID
	errorChan     chan error
	wg            *sync.WaitGroup
}

func (q *PluginQueue) Add(id model.ID) {
	q.iDChanel <- id
}

func (q *PluginQueue) ErrorChan() chan error {
	return q.errorChan
}

func (q *PluginQueue) Close() error {
	var err error
	close(q.iDChanel)
	// wait for all goworker to finish before closing the pool
	q.wg.Wait()
	for _, v := range q.plugins {
		cErr := v.Close()
		if cErr != nil {
			err = errors.Join(err, cErr)
		}
	}
	close(q.errorChan)
	return err
}

func (q *PluginQueue) start(ctx context.Context, plugDir string, plugProv plugin.Provider, numOfWorker int) error {
	// TODO: clean this up
	pluginsSlice := make(map[int][]plugin.Plugin, numOfWorker)
	q.delayTickers = make([]*time.Ticker, len(q.pluginInfos))
	q.pluginConfigs = make(map[int]map[string]string, len(q.pluginInfos))
	q.wg = &sync.WaitGroup{}
	for plugIdx, plugInf := range q.pluginInfos {
		plugFile, err := os.Open(filepath.Join(plugDir, plugProv.FileName(plugInf)))
		if err != nil {
			if os.IsNotExist(err) {
				return errors.Join(ErrNotFound, fmt.Errorf("plugin %s not found", plugInf.Name))
			}
			return errors.Join(ErrUnexpected, err)
		}
		cont, err := io.ReadAll(plugFile)
		if err != nil {
			plugFile.Close()
			return errors.Join(ErrUnexpected, err)
		}
		plugFile.Close()

		for i := range numOfWorker {
			plug, err := plugProv.Load(string(cont))
			if err != nil {
				q.Close()
				return errors.Join(ErrUnexpected, err)
			}
			q.plugins = append(q.plugins, plug)
			pluginsSlice[i] = append(pluginsSlice[i], plug)
		}

		conf := make(map[string]string, len(plugInf.Config))
		for _, v := range plugInf.Config {
			conf[v.Name] = v.Value
		}
		q.pluginConfigs[plugIdx] = conf
		q.delayTickers[plugIdx] = time.NewTicker(plugInf.Delay)
	}

	for i := range numOfWorker {
		go func() {
			for {
				select {
				case id, ok := <-q.iDChanel:
					if !ok {
						return
					}
					q.wg.Add(1)
					for i, plug := range pluginsSlice[i] {
						<-q.delayTickers[i].C
						err := plug.Run(ctx, q.pluginConfigs[i], map[string]string{}, id)
						if err != nil {
							q.errorChan <- errors.Join(ErrUnexpected, fmt.Errorf("failed to process id %d, error: %w", id, err))
						}
					}
					q.wg.Done()

				case <-ctx.Done():
					return
				}
			}
		}()
	}
	return nil
}

func (plugMod pluginModule) LoadAutoRuns(ctx context.Context, targetType plugin.Target, numOfworker int) (*PluginQueue, error) {
	plugs, err := plugMod.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	plugs = slices.DeleteFunc(plugs, func(p plugin.PluginInfo) bool {
		return !p.AutoRun || p.TargetEntity != targetType
	})

	q := &PluginQueue{
		pluginInfos: plugs,
		iDChanel:    make(chan model.ID, numOfworker),
		errorChan:   make(chan error),
	}
	err = q.start(ctx, plugMod.app.pluginDir, plugMod.app.pluginProvider, numOfworker)
	if err != nil {
		return nil, err
	}

	return q, nil
}
