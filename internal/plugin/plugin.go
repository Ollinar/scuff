// Package plugin defines the plugin system that app uses
package plugin

import (
	"context"
	"time"

	"github.com/Ollinar/scuff/internal/model"
)

type Target string

const (
	TargetNone    Target = "none"
	TargetArchive Target = "archive"
	TargetChapter Target = "chapter"
)

type Provider interface {
	Validate(string) (PluginInfo, error)
	FileName(PluginInfo) string
	Execute(ctx context.Context, script string, config, param map[string]string, id model.ID) error
	Load(ctx context.Context, script string, config map[string]string) (Plugin, error)
}

type Plugin interface {
	Close() error
	QueueUp(id model.ID)
}

type PluginInfo struct {
	Name        string
	Description string
	Version     string
	// Delay in milliseconds when executing multiple times
	Delay   time.Duration
	AutoRun bool
	// TargetEntity specifies what the plugin should be ran against
	TargetEntity Target
	Config       []PluginConfig
	Param        []PluginParam
}

type PluginConfig struct {
	Name        string
	Value       string
	Description string
}

type PluginParam = PluginConfig
