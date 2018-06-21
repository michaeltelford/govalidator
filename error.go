package govalidator

import "strings"

// Error encapsulates a name, an error and whether there's a custom error message or not.
type Error struct {
	Name                     string
	Err                      error
	CustomErrorMessageExists bool

	// Validator indicates the name of the validator that failed
	Validator string
}

func (e Error) Error() string {
	return strings.Trim(e.Err.Error(), ` `)
}

// NewError from existing error.
func NewError(err error) Error {
	return Error{
		Err: err,
		CustomErrorMessageExists: true,
	}
}

// Errors is an array of multiple errors and conforms to the error interface.
type Errors []Error

// Errors returns itself.
func (es Errors) Errors() (errs []error) {
	for _, err := range es {
		errs = append(errs, err.Err)
	}
	return
}

func (es Errors) Error() string {
	var errs []string
	for _, e := range es {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, ";")
}
