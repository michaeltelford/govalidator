package govalidator

import (
	"fmt"
	"testing"
)

func TestErrorsToString(t *testing.T) {
	t.Parallel()
	customErr := Error{Name: "Custom Error Name", Err: fmt.Errorf("stdlib error")}
	customErrWithCustomErrorMessage := Error{Name: "Custom Error Name 2", Err: fmt.Errorf("Bad stuff happened"), CustomErrorMessageExists: true}

	var tests = []struct {
		param1   Errors
		expected string
	}{
		{Errors{}, ""},
		{Errors{NewError(fmt.Errorf("Error 1"))}, "Error 1"},
		{Errors{NewError(fmt.Errorf("Error 1")), NewError(fmt.Errorf("Error 2"))}, "Error 1;Error 2"},
		{Errors{customErr, NewError(fmt.Errorf("Error 2"))}, "stdlib error;Error 2"},
		{Errors{NewError(fmt.Errorf("Error 123")), customErrWithCustomErrorMessage}, "Error 123;Bad stuff happened"},
	}
	for _, test := range tests {
		actual := test.param1.Error()
		if actual != test.expected {
			t.Errorf("Expected Error() to return '%v', got '%v'", test.expected, actual)
		}
	}
}
