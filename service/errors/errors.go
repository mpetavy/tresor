package errors

import "fmt"

type ErrUnknownService struct {
	Service string
}

func (e *ErrUnknownService) Error() string { return fmt.Sprintf("Unknown service: %s", e.Service) }

type ErrUnknownDriver struct {
	Driver string
}

func (e *ErrUnknownDriver) Error() string { return fmt.Sprintf("Unknown driver: %s", e.Driver) }

type ErrServiceCreate struct {
	Service string
}

func (e *ErrServiceCreate) Error() string { return fmt.Sprintf("Cannot create service: %s", e.Service) }
