package errorz

import (
	"fmt"
	"strings"
)

type MultipleErrors []error

func (e MultipleErrors) Error() string {
	if len(e) == 0 {
		return "no errors occurred"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	buf := strings.Builder{}
	buf.WriteString("multiple errors occurred")
	for idx, err := range e {
		buf.WriteString(fmt.Sprintf(" %v: %v\n", idx, err))
	}
	return buf.String()
}

func (e MultipleErrors) ToError() error {
	if len(e) == 0 {
		return nil
	}
	if len(e) == 1 {
		return e[0]
	}
	return e
}
