//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import "github.com/edgexfoundry/edgex-go/export/mongo"

var repo *mongo.Repository

// InitMongoRepository - Init Mongo DB
func InitMongoRepository(r *mongo.Repository) {
	repo = r
	return
}
