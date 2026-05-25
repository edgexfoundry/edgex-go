package versions

import (
	"runtime"
	"runtime/debug"
	"time"
)

type DefaultVersionProvider struct {
	VersionInfo
}

func (v DefaultVersionProvider) EncoderDecoder() VersionEncDec {
	return &StdVersionEncDec
}

func (v DefaultVersionProvider) Version() string {
	return v.VersionInfo.Version
}

func (v DefaultVersionProvider) BuildDate() string {
	return v.VersionInfo.BuildDate
}

func (v DefaultVersionProvider) Revision() string {
	return v.VersionInfo.Revision
}

func (v DefaultVersionProvider) AsVersionInfo() *VersionInfo {
	return &v.VersionInfo
}

func NewDefaultVersionProvider() VersionProvider {
	buildInfo, ok := debug.ReadBuildInfo()
	rev := ""
	if ok {
		for _, kv := range buildInfo.Settings {
			if kv.Key == "vcs.revision" {
				rev = kv.Value
			}
		}
	}
	return &DefaultVersionProvider{
		VersionInfo: VersionInfo{
			Version:   "v0.0.0",
			Revision:  rev,
			BuildDate: time.Now().String(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		},
	}
}
