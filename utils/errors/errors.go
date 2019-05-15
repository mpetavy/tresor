package errors

import "fmt"

type ErrProcessState struct {
	State int
}

func (e *ErrProcessState) Error() string { return fmt.Sprintf("Error process state: %d", e.State) }
