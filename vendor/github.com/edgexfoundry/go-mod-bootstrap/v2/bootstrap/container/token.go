/*******************************************************************************
 * Copyright 2020 Intel Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package container

import (
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/authtokenloader"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

//// FileIoPerformerInterfaceName contains the name of the fileioperformer.FileIoPerformer implementation in the DIC.
//var FileIoPerformerInterfaceName = di.TypeInstanceToName((*fileioperformer.FileIoPerformer)(nil))
//
//// FileIoPerformerFrom helper function queries the DIC and returns the fileioperformer.FileIoPerformer implementation.
//func FileIoPerformerFrom(get di.Get) fileioperformer.FileIoPerformer {
//	fileIo := get(FileIoPerformerInterfaceName)
//	if fileIo != nil {
//		return fileIo.(fileioperformer.FileIoPerformer)
//	}
//	return (fileioperformer.FileIoPerformer)(nil)
//}

// AuthTokenLoaderInterfaceName contains the name of the authtokenloader.AuthTokenLoader implementation in the DIC.
var AuthTokenLoaderInterfaceName = di.TypeInstanceToName((*authtokenloader.AuthTokenLoader)(nil))

// AuthTokenLoaderFrom helper function queries the DIC and returns the authtokenloader.AuthTokenLoader implementation.
func AuthTokenLoaderFrom(get di.Get) authtokenloader.AuthTokenLoader {
	loader, ok := get(AuthTokenLoaderInterfaceName).(authtokenloader.AuthTokenLoader)
	if !ok {
		return nil
	}

	return loader
}
