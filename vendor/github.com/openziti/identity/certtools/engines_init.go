// +build engines

/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package certtools

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

var supportedEngines = []string{
	"siometrics.so",
	"authenta.so",
}

var loaded = false
var loadMu = sync.Mutex{}

func loadEngines() {
	if loaded {
		return
	}

	loadMu.Lock()
	defer loadMu.Unlock()

	if loaded {
		return
	}

	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	log.Debug("execpath ", execDir)

	for _, lib := range supportedEngines {
		libPath := filepath.Join(execDir, lib)
		log.Debug("trying", libPath)
		l, err := plugin.Open(libPath)

		if err != nil {
			log.Debugf("failed load engine %s: %s\n", libPath, err.Error())
			continue
		}

		e, err := l.Lookup("Engine")
		if err != nil {
			log.Warnln("Symbol 'Engine' not found in ", lib, libPath)
			continue
		}

		if engine, ok := e.(Engine); ok {
			log.Debugf("found engine '%s' in %s(%s)\n", engine.Id(), lib, libPath)
			engines[engine.Id()] = engine
		} else {
			log.Warnf("Engine instance is not valid in (%s)\n", lib)
			continue
		}
	}

	loaded = true
}
