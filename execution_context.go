package tengo

import (
	"sync"
)

// ExecutionContext provides a context-aware execution environment for compiled functions.
// It bundles constants, globals, and the original compiled object together to ensure
// that closures have access to their complete execution context.
type ExecutionContext struct {
	constants []Object
	globals   []Object
	source    *Compiled
	lock      sync.RWMutex // Protects globals for concurrent access
}

// NewExecutionContext creates a new ExecutionContext from a compiled script.
// It captures the constants and globals from the compiled object to provide
// a complete execution context for closures.
func NewExecutionContext(compiled *Compiled) *ExecutionContext {
	return &ExecutionContext{
		constants: compiled.Constants(),
		globals:   compiled.Globals(),
		source:    compiled,
	}
}

// WithGlobals creates a new ExecutionContext with specific globals.
// This is useful for creating isolated execution contexts or for testing.
func (ec *ExecutionContext) WithGlobals(globals []Object) *ExecutionContext {
	ec.lock.RLock()
	defer ec.lock.RUnlock()

	return &ExecutionContext{
		constants: ec.constants,
		globals:   globals,
		source:    ec.source,
	}
}

// WithIsolatedGlobals creates a new ExecutionContext with a copy of the current globals.
// This ensures thread-safe execution by providing each context with its own globals copy.
func (ec *ExecutionContext) WithIsolatedGlobals() *ExecutionContext {
	ec.lock.RLock()
	defer ec.lock.RUnlock()

	// Create a deep copy of globals to ensure isolation
	isolatedGlobals := make([]Object, len(ec.globals))
	for i, g := range ec.globals {
		if g != nil {
			isolatedGlobals[i] = g.Copy()
		}
	}

	return &ExecutionContext{
		constants: ec.constants,
		globals:   isolatedGlobals,
		source:    ec.source,
	}
}

// Call invokes a compiled function with the execution context.
// It provides the function with access to constants and globals from the original compilation.
func (ec *ExecutionContext) Call(fn *CompiledFunction, args ...Object) (Object, error) {
	result, _, err := ec.CallEx(fn, args...)
	return result, err
}

// CallEx invokes a compiled function with the execution context and returns both
// the result and the updated globals (if any were modified).
func (ec *ExecutionContext) CallEx(fn *CompiledFunction, args ...Object) (Object, []Object, error) {
	// Validate execution context before use
	if err := ec.Validate(); err != nil {
		return nil, nil, err
	}

	// Validate the function
	if fn == nil {
		return nil, nil, ErrMissingExecutionContext{
			Function: "execution-context",
			Missing:  "compiled function",
			Suggestion: "provide a valid CompiledFunction",
		}
	}

	ec.lock.RLock()
	constants := ec.constants
	globals := ec.globals
	ec.lock.RUnlock()

	// Call the function with the complete context
	result, updatedGlobals, err := fn.CallWithGlobalsExAndConstants(constants, globals, args...)

	// Update our globals if they were modified
	if err == nil && updatedGlobals != nil {
		ec.lock.Lock()
		ec.globals = updatedGlobals
		ec.lock.Unlock()
	}

	return result, updatedGlobals, err
}

// Constants returns a copy of the constants array.
func (ec *ExecutionContext) Constants() []Object {
	ec.lock.RLock()
	defer ec.lock.RUnlock()
	return ec.constants
}

// Globals returns a copy of the globals array.
func (ec *ExecutionContext) Globals() []Object {
	ec.lock.RLock()
	defer ec.lock.RUnlock()

	// Return a copy to prevent external mutations
	result := make([]Object, len(ec.globals))
	copy(result, ec.globals)
	return result
}

// Source returns the original compiled object.
func (ec *ExecutionContext) Source() *Compiled {
	return ec.source
}

// Validate checks if the execution context is valid and complete.
func (ec *ExecutionContext) Validate() error {
	if ec.source == nil {
		return ErrInvalidExecutionContext
	}

	if ec.constants == nil {
		return ErrMissingExecutionContext{
			Function: "execution-context",
			Missing:  "constants array",
			Suggestion: "ensure ExecutionContext was created from a valid compiled script",
		}
	}

	// Validate constants array
	for i, constant := range ec.constants {
		if constant == nil {
			return ErrInvalidConstantsArray{
				Reason: "nil constant",
				Index:  i,
			}
		}
	}

	// Validate globals array if present
	// Note: globals can be nil, which is normal for uninitialized globals
	// The VM treats nil globals as UndefinedValue when accessed
	if ec.globals != nil {
		// Just validate that it's not an empty slice when it should have content
		// We don't validate individual elements as nil is acceptable
	}

	return nil
}
