// Package fault provides structured error handling with error codes.
//
// This package implements a fault system that allows errors to carry
// additional metadata in the form of error codes. This enables better
// error categorization and handling throughout the application.
//
// Error codes are typed and can be extracted from error chains, making it
// easy to handle specific error conditions while preserving the original
// error information.
package fault

import (
	"errors"
	"fmt"
)

// Standard error codes used throughout the system.
const Ok = "OK"           // No error occurred
const Unknown = "UNKNOWN" // Unknown error type

// Code extracts the error code from an error, walking the error chain
// if necessary. Returns "OK" for nil errors and "UNKNOWN" for errors
// without a recognizable code.
func Code[T code](err error) T {
	if err == nil {
		return Ok
	}
	for err != nil {
		if c, ok := err.(*fault[T]); ok {
			return c.Code()
		}
		err = errors.Unwrap(err)
	}
	return Unknown
}

// RawCode extracts the raw (untyped) error code from an error.
// This is useful when you need to inspect error codes without
// knowing their specific type at compile time.
func RawCode(err error) interface{} {
	if err == nil {
		return nil
	}
	for err != nil {
		if rc, ok := err.(rawCoder); ok {
			return rc.RawCode()
		}
		err = errors.Unwrap(err)
	}
	return Unknown
}

func New[T code](
	msg string,
	code T,
) error {
	return &fault[T]{
		msg:  msg,
		code: code,
	}
}

func Wrap[T code](
	err error,
	prefix string,
	code T,
) error {
	if err == nil {
		return nil
	}
	result := &fault[T]{
		msg:     prefix,
		wrapped: err,
		code:    code,
	}
	return result
}

type code interface {
	~string
}

type rawCoder interface {
	RawCode() any
}

type fault[T code] struct {
	msg     string
	wrapped error
	code    T
}

func (e *fault[T]) Error() string {
	if e.wrapped == nil {
		return e.msg
	}
	if e.msg == "" {
		return e.wrapped.Error()
	}
	return fmt.Sprintf("%s: %s", e.msg, e.wrapped.Error())
}

func (e *fault[T]) Unwrap() error {
	return e.wrapped
}

func (err *fault[T]) Code() T {
	return err.code
}

func (err *fault[T]) RawCode() any {
	return err.code
}
