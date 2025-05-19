package utils

import "errors"

// WrapError creates a new error to wrap around another error.
// Useful in cases where you have an error you need to bring up through multiple levels of error returns.
func WrapError(wrapErrorMsg string, errorToWrap error) error {
	return errors.Join(errors.New(wrapErrorMsg), errorToWrap)
}
