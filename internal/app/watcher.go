package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Ollinar/scuff/internal/plugin"

	"github.com/fsnotify/fsnotify"
	"github.com/gabriel-vasile/mimetype"
)

func watcherHandler(ap *App) {
	wt := ap.watcher
	lg := ap.logger.WithGroup("fsnotify")

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

mainloop:
	for {
		select {
		case err, ok := <-wt.Errors:
			if !ok {
				lg.Debug("watcher closed")
				return
			}
			if err != nil {
				lg.Error("unexpected error reading watcher error channel", slog.String("error", err.Error()))
			}
		case e, ok := <-wt.Events:
			if !ok {
				return
			}

			lg.Debug("event recieved", slog.Any("event", e.String()))
			if e.Has(fsnotify.Create) {
				finf, err := os.Stat(e.Name)
				if err != nil {
					lg.Error("failed to get the stat of new file", slog.String("fileName", e.Name))
					continue
				}
				if finf.IsDir() {
					err = wt.Add(e.Name)
					if err != nil {
						lg.Error("failed to add directory to watcher", slog.String("dir", e.Name))
					}
					continue
				}

				lastStat, err := os.Stat(e.Name)
				if err != nil {
					lg.Error("failed to get the stat of new file", slog.String("fileName", e.Name))
					continue
				}

				// wait for any writes to finish, if after 30 seconds, the size and modtime hasnt changed, assume its done writing.
				unchangedCounter := 0
				for {
					time.Sleep(5 * time.Second)
					newStat, err := os.Stat(e.Name)
					if err != nil {
						lg.Error("failed to get the stat of new file", slog.String("fileName", e.Name))
						continue mainloop
					}

					if lastStat.Size() == newStat.Size() && lastStat.ModTime().Equal(newStat.ModTime()) {
						unchangedCounter++
						if unchangedCounter == 6 {
							break
						}
					} else {
						lastStat = newStat
						// reset counter
						unchangedCounter = 0
					}
				}

				m, err := mimetype.DetectFile(e.Name)
				if err != nil {
					lg.Error("failed to detect the mime of new file", slog.String("fileName", e.Name))
					continue
				}
				if !m.Is("application/zip") {
					continue
				}
				arc, isNew, err := ap.Archive().AddZipToArchives(ctx, e.Name)
				if err != nil {
					lg.Error("failed to add the new file to Archives",
						slog.String("fileName", e.Name),
						slog.Any("error", err))
					continue
				}
				if isNew {
					plugins, err := ap.Plugin().LoadAutoRuns(ctx, plugin.TargetArchive)
					if err != nil {
						lg.Error("failed to run auto run plugin", slog.Int64("archive id", arc.ID), slog.Any("error", err))
						continue
					}
					for _, plug := range plugins {
						plug.QueueUp(arc.ID)
					}
				}

				lg.Debug("added archive", slog.Any("archive", arc))
			}

		}
	}
}
