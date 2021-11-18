package secretstore

import (
	"os"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/stretchr/testify/assert"
)

func TestKongAdminAPI_Setup(t *testing.T) {

	configuration := &config.ConfigurationStruct{}
	_ = di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	k := NewKongAdminAPI(config.KongAdminInfo{
		ConfigTemplatePath: "../../../cmd/security-secretstore-setup/res/kong-admin-config.template.yml",
		ConfigFilePath:     "/tmp/kong/kong.yml",
		ConfigJWTPath:      "/tmp/kong/kong-admin-jwt",
		ConfigJWTDuration:  "1h",
	})

	err := k.Setup()
	assert.NoError(t, err)

	// check file permission of jwt
	info, err := os.Stat(k.paths.jwt)
	assert.NoError(t, err)

	fileMode := info.Mode().String()
	assert.Equal(t, "-rw-------", fileMode, "The jwt should be read+write for owner")

	// Clean up files created
	os.Remove(k.paths.config)
	os.Remove(k.paths.jwt)
}
