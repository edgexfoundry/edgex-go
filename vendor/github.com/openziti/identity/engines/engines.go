package engines

import (
	"crypto"
	"net/url"
)

func RegisterEngine(e Engine) {
	engines[e.Id()] = e
}

type Engine interface {
	Id() string
	LoadKey(key *url.URL) (crypto.PrivateKey, error)
}

var engines = map[string]Engine{}

func ListEngines() []string {
	res := make([]string, 0, len(engines))
	for k := range engines {
		res = append(res, k)
	}
	return res
}

func GetEngine(id string) (Engine, bool) {
	e, found := engines[id]
	return e, found
}
