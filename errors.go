package tengo

import (
	"errors"
	"fmt"
)

var (
	// ErrStackOverflow is a stack overflow error.
	ErrStackOverflow = errors.New("stack overflow")

	// ErrObjectAllocLimit is an objects allocation limit error.
	ErrObjectAllocLimit = errors.New("object allocation limit exceeded")

	// ErrIndexOutOfBounds is an error where a given index is out of the
	// bounds.
	ErrIndexOutOfBounds = errors.New("index out of bounds")

	// ErrInvalidIndexType represents an invalid index type.
	ErrInvalidIndexType = errors.New("invalid index type")

	// ErrInvalidIndexValueType represents an invalid index value type.
	ErrInvalidIndexValueType = errors.New("invalid index value type")

	// ErrInvalidIndexOnError represents an invalid index on error.
	ErrInvalidIndexOnError = errors.New("invalid index on error")

	// ErrInvalidOperator represents an error for invalid operator usage.
	ErrInvalidOperator = errors.New("invalid operator")

	// ErrWrongNumArguments represents a wrong number of arguments error.
	ErrWrongNumArguments = errors.New("wrong number of arguments")

	// ErrBytesLimit represents an error where the size of bytes value exceeds
	// the limit.
	ErrBytesLimit = errors.New("exceeding bytes size limit")

	// ErrStringLimit represents an error where the size of string value
	// exceeds the limit.
	ErrStringLimit = errors.New("exceeding string size limit")

	// ErrNotIndexable is an error where an Object is not indexable.
	ErrNotIndexable = errors.New("not indexable")

	// ErrNotIndexAssignable is an error where an Object is not index
	// assignable.
	ErrNotIndexAssignable = errors.New("not index-assignable")

	// ErrNotImplemented is an error where an Object has not implemented a
	// required method.
	ErrNotImplemented = errors.New("not implemented")

	// ErrInvalidRangeStep is an error where the step parameter is less than or equal to 0 when using builtin range function.
	ErrInvalidRangeStep = errors.New("range step must be greater than 0")

	// ErrMissingConstants represents an error where constants are required but not provided.
	ErrMissingConstants = errors.New("missing constants for function execution")

	// ErrMissingGlobals represents an error where globals are required but not provided.
	ErrMissingGlobals = errors.New("missing globals for function execution")

	// ErrIncompleteExecutionContext represents an error where execution context is incomplete.
	ErrIncompleteExecutionContext = errors.New("incomplete execution context")

	// ErrInvalidExecutionContext represents an error where execution context is invalid.
	ErrInvalidExecutionContext = errors.New("invalid execution context")
)

// ErrInvalidArgumentType represents an invalid argument value type error.
type ErrInvalidArgumentType struct {
	Name     string
	Expected string
	Found    string
}

func (e ErrInvalidArgumentType) Error() string {
	return fmt.Sprintf("invalid type for argument '%s': expected %s, found %s",
		e.Name, e.Expected, e.Found)
}

// ErrMissingExecutionContext represents an error where execution context is missing required components.
type ErrMissingExecutionContext struct {
	Function   string
	Missing    string
	Suggestion string
}

func (e ErrMissingExecutionContext) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("function '%s' requires %s for execution - %s",
			e.Function, e.Missing, e.Suggestion)
	}
	return fmt.Sprintf("function '%s' requires %s for execution",
		e.Function, e.Missing)
}

// ErrInvalidConstantsArray represents an error where constants array is invalid.
type ErrInvalidConstantsArray struct {
	Reason string
	Index  int
}

func (e ErrInvalidConstantsArray) Error() string {
	if e.Index >= 0 {
		return fmt.Sprintf("invalid constants array at index %d: %s", e.Index, e.Reason)
	}
	return fmt.Sprintf("invalid constants array: %s", e.Reason)
}

// ErrInvalidGlobalsArray represents an error where globals array is invalid.
type ErrInvalidGlobalsArray struct {
	Reason string
	Index  int
}

func (e ErrInvalidGlobalsArray) Error() string {
	if e.Index >= 0 {
		return fmt.Sprintf("invalid globals array at index %d: %s", e.Index, e.Reason)
	}
	return fmt.Sprintf("invalid globals array: %s", e.Reason)
}
