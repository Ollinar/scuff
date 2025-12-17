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
	Load(string) (Plugin, error)
	FileName(PluginInfo) string
}

type Plugin interface {
	Close() error
	Run(ctx context.Context, config, param map[string]string, id model.ID) error
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
