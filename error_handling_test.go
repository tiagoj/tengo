package tengo

import (
	"testing"
)

func TestContextErrorTypes(t *testing.T) {
	// Test ErrMissingExecutionContext
	err := ErrMissingExecutionContext{
		Function:   "test-function",
		Missing:    "constants",
		Suggestion: "use ExecutionContext",
	}
	expected := "function 'test-function' requires constants for execution - use ExecutionContext"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}

	// Test ErrMissingExecutionContext without suggestion
	err2 := ErrMissingExecutionContext{
		Function: "test-function",
		Missing:  "constants",
	}
	expected2 := "function 'test-function' requires constants for execution"
	if err2.Error() != expected2 {
		t.Errorf("Expected error message '%s', got '%s'", expected2, err2.Error())
	}

	// Test ErrInvalidConstantsArray
	err3 := ErrInvalidConstantsArray{
		Reason: "nil constant",
		Index:  2,
	}
	expected3 := "invalid constants array at index 2: nil constant"
	if err3.Error() != expected3 {
		t.Errorf("Expected error message '%s', got '%s'", expected3, err3.Error())
	}

	// Test ErrInvalidConstantsArray without index
	err4 := ErrInvalidConstantsArray{
		Reason: "empty array",
		Index:  -1,
	}
	expected4 := "invalid constants array: empty array"
	if err4.Error() != expected4 {
		t.Errorf("Expected error message '%s', got '%s'", expected4, err4.Error())
	}

	// Test ErrInvalidGlobalsArray
	err5 := ErrInvalidGlobalsArray{
		Reason: "nil global",
		Index:  1,
	}
	expected5 := "invalid globals array at index 1: nil global"
	if err5.Error() != expected5 {
		t.Errorf("Expected error message '%s', got '%s'", expected5, err5.Error())
	}
}

func TestCompiledFunctionErrorHandling(t *testing.T) {
	// Create a simple compiled function
	fn := &CompiledFunction{
		Instructions:  []byte{0x01, 0x02}, // Some dummy instructions
		NumLocals:     1,
		NumParameters: 1,
		VarArgs:       false,
	}

	// Test calling without constants
	_, _, err := fn.CallWithGlobalsExAndConstants(nil, nil, &Int{Value: 42})
	if err == nil {
		t.Error("Expected error when calling function without constants")
	}

	// Check that it's the correct error type
	if missingCtxErr, ok := err.(ErrMissingExecutionContext); ok {
		if missingCtxErr.Function != "compiled-function" {
			t.Errorf("Expected function name 'compiled-function', got '%s'", missingCtxErr.Function)
		}
		if missingCtxErr.Missing != "constants from original compilation" {
			t.Errorf("Expected missing 'constants from original compilation', got '%s'", missingCtxErr.Missing)
		}
	} else {
		t.Errorf("Expected ErrMissingExecutionContext, got %T", err)
	}

	// Test with invalid constants (nil constant)
	constants := []Object{&Int{Value: 1}, nil, &String{Value: "test"}}
	_, _, err = fn.CallWithGlobalsExAndConstants(constants, nil, &Int{Value: 42})
	if err == nil {
		t.Error("Expected error when calling function with nil constant")
	}

	// Check that it's the correct error type
	if invalidConstsErr, ok := err.(ErrInvalidConstantsArray); ok {
		if invalidConstsErr.Index != 1 {
			t.Errorf("Expected index 1, got %d", invalidConstsErr.Index)
		}
		if invalidConstsErr.Reason != "nil constant" {
			t.Errorf("Expected reason 'nil constant', got '%s'", invalidConstsErr.Reason)
		}
	} else {
		t.Errorf("Expected ErrInvalidConstantsArray, got %T", err)
	}
}

func TestExecutionContextValidation(t *testing.T) {
	// Create a compiled script for testing
	script := NewScript([]byte("x := 42"))
	compiled, err := script.Compile()
	if err != nil {
		t.Fatalf("Failed to compile script: %v", err)
	}

	// Create a valid execution context
	ctx := NewExecutionContext(compiled)
	
	// Test valid context
	if err := ctx.Validate(); err != nil {
		t.Errorf("Expected valid context to pass validation, got error: %v", err)
	}

	// Test invalid context - nil source
	invalidCtx := &ExecutionContext{
		constants: []Object{&Int{Value: 1}},
		globals:   []Object{},
		source:    nil,
	}
	err = invalidCtx.Validate()
	if err != ErrInvalidExecutionContext {
		t.Errorf("Expected ErrInvalidExecutionContext for nil source, got %v", err)
	}

	// Test invalid context - nil constants
	invalidCtx2 := &ExecutionContext{
		constants: nil,
		globals:   []Object{},
		source:    compiled,
	}
	err = invalidCtx2.Validate()
	if err == nil {
		t.Error("Expected error for nil constants")
	}
	if missingCtxErr, ok := err.(ErrMissingExecutionContext); ok {
		if missingCtxErr.Function != "execution-context" {
			t.Errorf("Expected function name 'execution-context', got '%s'", missingCtxErr.Function)
		}
	} else {
		t.Errorf("Expected ErrMissingExecutionContext, got %T", err)
	}

	// Test invalid context - nil constant in array
	invalidCtx3 := &ExecutionContext{
		constants: []Object{&Int{Value: 1}, nil},
		globals:   []Object{},
		source:    compiled,
	}
	err = invalidCtx3.Validate()
	if err == nil {
		t.Error("Expected error for nil constant in array")
	}
	if invalidConstsErr, ok := err.(ErrInvalidConstantsArray); ok {
		if invalidConstsErr.Index != 1 {
			t.Errorf("Expected index 1, got %d", invalidConstsErr.Index)
		}
	} else {
		t.Errorf("Expected ErrInvalidConstantsArray, got %T", err)
	}

	// Test context with nil globals (should be valid - nil globals are allowed)
	validCtx4 := &ExecutionContext{
		constants: []Object{&Int{Value: 1}},
		globals:   []Object{&String{Value: "test"}, nil}, // nil globals are valid
		source:    compiled,
	}
	err = validCtx4.Validate()
	if err != nil {
		t.Errorf("Expected nil globals to be valid, got error: %v", err)
	}
}

func TestExecutionContextCallValidation(t *testing.T) {
	// Create a compiled script for testing
	script := NewScript([]byte("x := 42"))
	compiled, err := script.Compile()
	if err != nil {
		t.Fatalf("Failed to compile script: %v", err)
	}

	// Create a valid execution context
	ctx := NewExecutionContext(compiled)

	// Test calling with nil function
	_, err = ctx.Call(nil, &Int{Value: 42})
	if err == nil {
		t.Error("Expected error when calling with nil function")
	}
	if missingCtxErr, ok := err.(ErrMissingExecutionContext); ok {
		if missingCtxErr.Missing != "compiled function" {
			t.Errorf("Expected missing 'compiled function', got '%s'", missingCtxErr.Missing)
		}
	} else {
		t.Errorf("Expected ErrMissingExecutionContext, got %T", err)
	}

	// Test CallEx with nil function
	_, _, err = ctx.CallEx(nil, &Int{Value: 42})
	if err == nil {
		t.Error("Expected error when calling CallEx with nil function")
	}
	if missingCtxErr, ok := err.(ErrMissingExecutionContext); ok {
		if missingCtxErr.Missing != "compiled function" {
			t.Errorf("Expected missing 'compiled function', got '%s'", missingCtxErr.Missing)
		}
	} else {
		t.Errorf("Expected ErrMissingExecutionContext, got %T", err)
	}
}
