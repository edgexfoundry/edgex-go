package ziti

import "github.com/openziti/edge-api/rest_model"

type Precedence byte

func (p Precedence) String() string {
	if p == PrecedenceRequired {
		return PrecedenceRequiredLabel
	}
	if p == PrecedenceFailed {
		return PrecedenceFailedLabel
	}
	return PrecedenceDefaultLabel
}

const (
	PrecedenceDefault  Precedence = 0
	PrecedenceRequired Precedence = 1
	PrecedenceFailed   Precedence = 2

	PrecedenceDefaultLabel  = string(rest_model.TerminatorPrecedenceDefault)
	PrecedenceRequiredLabel = string(rest_model.TerminatorPrecedenceRequired)
	PrecedenceFailedLabel   = string(rest_model.TerminatorPrecedenceFailed)
)

func GetPrecedenceForLabel(p string) Precedence {
	if p == PrecedenceRequiredLabel {
		return PrecedenceRequired
	}
	if p == PrecedenceFailedLabel {
		return PrecedenceFailed
	}
	return PrecedenceDefault
}
