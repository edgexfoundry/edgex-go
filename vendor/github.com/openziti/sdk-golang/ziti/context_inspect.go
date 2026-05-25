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

package ziti

import (
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/inspect"
	"github.com/openziti/sdk-golang/ziti/edge"
)

func (context *ContextImpl) Inspect() *inspect.ContextInspectResult {
	result := &inspect.ContextInspectResult{
		ContextId: context.Id,
	}

	// Identity
	result.Identity = &inspect.ContextInspectIdentity{}
	if cachedIdentity := context.cachedIdentity.Load(); cachedIdentity != nil {
		if cachedIdentity.ID != nil {
			result.Identity.Id = *cachedIdentity.ID
		}
		if cachedIdentity.Name != nil {
			result.Identity.Name = *cachedIdentity.Name
		}
	}

	toStrSlice := func(dba rest_model.DialBindArray) []string {
		var s []string
		for _, v := range dba {
			s = append(s, string(v))
		}
		return s
	}

	// Services
	context.services.IterCb(func(key string, svc *rest_model.ServiceDetail) {
		result.Services = append(result.Services, &inspect.ContextInspectService{
			Id:          *svc.ID,
			Name:        *svc.Name,
			Permissions: toStrSlice(svc.Permissions),
		})
	})

	// Sessions
	context.sessions.IterCb(func(key string, session *rest_model.SessionDetail) {
		result.Sessions = append(result.Sessions, &inspect.ContextInspectSession{
			Id:        *session.ID,
			ServiceId: session.Service.ID,
			Type:      string(*session.Type),
		})
	})

	// Router Connections
	context.routerConnections.IterCb(func(key string, conn edge.RouterConn) {
		result.RouterConnections = append(result.RouterConnections, conn.Inspect())
	})

	// Listener Managers
	context.listenerManagers.IterCb(func(key string, mgr *listenerManager) {
		result.Listeners = append(result.Listeners, mgr.InspectListener())
	})

	return result
}
