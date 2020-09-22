package redis

import dataInterfaces "github.com/edgexfoundry/edgex-go/internal/core/data/v2/infrastructure/interfaces"
import metadataInterfaces "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/infrastructure/interfaces"

// Check the implementation of Redis satisfies the DB client
var _ dataInterfaces.DBClient = &Client{}
var _ metadataInterfaces.DBClient = &Client{}
