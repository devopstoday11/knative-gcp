/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package volume

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/google/knative-gcp/pkg/broker/config"
	"google.golang.org/protobuf/proto"
)

const (
	defaultPath = "/var/run/cloud-run-events/broker/targets"
)

// Targets implements config.ReadonlyTargets with data
// loaded from a file.
// It also watches the file for any changes and will automatically
// refresh the in memory cache.
type Targets struct {
	config.CachedTargets
	path       string
	notifyChan chan<- struct{}
}

var _ config.ReadonlyTargets = (*Targets)(nil)

// NewTargetsFromFile initializes the targets config from a file.
func NewTargetsFromFile(opts ...Option) (config.ReadonlyTargets, error) {
	t := &Targets{
		CachedTargets: config.CachedTargets{},
		path:          defaultPath,
	}

	for _, opt := range opts {
		opt(t)
	}

	if err := t.sync(); err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := t.watchWith(watcher); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Targets) watchWith(watcher *fsnotify.Watcher) error {
	configFile := filepath.Clean(t.path)
	configDir, _ := filepath.Split(t.path)
	realConfigFile, _ := filepath.EvalSymlinks(t.path)
	if err := watcher.Add(configDir); err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					// 'Events' channel is closed.
					return
				}
				currentConfigFile, _ := filepath.EvalSymlinks(t.path)

				// Re-sync if the file was updated/created or
				// if the real file was replaced.
				const writeOrCreateMask = fsnotify.Write | fsnotify.Create
				if (filepath.Clean(event.Name) == configFile &&
					event.Op&writeOrCreateMask != 0) ||
					(currentConfigFile != "" && currentConfigFile != realConfigFile) {
					realConfigFile = currentConfigFile
					if err := t.sync(); err != nil {
						log.Printf("error syncing config: %v\n", err)
					} else if t.notifyChan != nil {
						// File got updated and notify the external channel.
						t.notifyChan <- struct{}{}
					}
				}

			case err, ok := <-watcher.Errors:
				if ok {
					log.Printf("watcher error: %v\n", err)
				}
				return
			}
		}
	}()
	return nil
}

func (t *Targets) sync() error {
	b, err := t.readFile()
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var val config.TargetsConfig
	if err := proto.Unmarshal(b, &val); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	t.Store(&val)
	return nil
}

func (t *Targets) readFile() ([]byte, error) {
	return ioutil.ReadFile(t.path)
}
