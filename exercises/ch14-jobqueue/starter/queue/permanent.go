package queue

import "errors"

// permanentError marks a failure that retrying cannot fix (bad input, 4xx from a
// receiver, a deleted account). Fail buries such a job at once.
type permanentError struct{ err error }

func (e permanentError) Error() string { return "permanent: " + e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

// Permanent wraps err so that Fail moves the job straight to dead.
func Permanent(err error) error { return permanentError{err} }

// IsPermanent reports whether err, or anything it wraps, was made by Permanent.
func IsPermanent(err error) bool {
	_, ok := errors.AsType[permanentError](err)
	return ok
}
