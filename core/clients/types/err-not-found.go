package types

type ErrNotFound struct {}

func(e ErrNotFound) Error() string {
	return "item not found"
}
