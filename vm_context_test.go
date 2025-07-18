package tengo

import (
	"testing"
)

// TestVMContext_DirectFunctionCall tests direct function calls with globals and constants
func TestVMContext_DirectFunctionCall(t *testing.T) {
	// Create a simple script that uses a closure with globals
	script := `
		global_var := 42
		
		closure := func() {
			return global_var + 10
		}
		
		result := closure()
		`

	// Compile the script
	s := NewScript([]byte(script))
	compiled, err := s.Compile()
	if err != nil {
		t.Fatal(err)
	}

	// Run the script
	err = compiled.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Get the result
	result := compiled.Get("result")
	if result.IsUndefined() {
		t.Fatal("result variable not found")
	}

	// Verify the result is 52 (42 + 10)
	if intVal, ok := result.Value().(int64); ok {
		if intVal != 52 {
			t.Errorf("Expected result 52, got %d", intVal)
		}
	} else {
		t.Errorf("Expected int64 result, got %T", result.Value())
	}

	// Test direct function call
	closureVar := compiled.Get("closure")
	if closureVar.IsUndefined() {
		t.Fatal("closure variable not found")
	}

	// Try to call the closure function directly
	if fn, ok := closureVar.Value().(*CompiledFunction); ok {
		// Get the constants and globals from the compiled script
		constants := compiled.Constants()
		globals := compiled.Globals()

		// Call the function with constants and globals
		directResult, _, err := fn.CallWithGlobalsExAndConstants(constants, globals)
		if err != nil {
			t.Fatal(err)
		}

		// Verify the direct call result is also 52
		if intResult, ok := directResult.(*Int); ok {
			if intResult.Value != 52 {
				t.Errorf("Expected direct call result 52, got %d", intResult.Value)
			}
		} else {
			t.Errorf("Expected Int result from direct call, got %T", directResult)
		}
	} else {
		t.Fatal("closure is not a compiled function")
	}
}

// TestVMContext_FunctionWithArguments tests direct function calls with arguments
func TestVMContext_FunctionWithArguments(t *testing.T) {
	script := `
		multiplier := 2
		
		multiply := func(x) {
			return multiplier * x
		}
	`

	s := NewScript([]byte(script))
	compiled, err := s.Compile()
	if err != nil {
		t.Fatal(err)
	}

	err = compiled.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Get the function
	multiplyVar := compiled.Get("multiply")
	if multiplyVar.IsUndefined() {
		t.Fatal("multiply variable not found")
	}

	multiplyFn, ok := multiplyVar.Value().(*CompiledFunction)
	if !ok {
		t.Fatal("multiply is not a compiled function")
	}

	// Test with original globals (multiplier = 2)
	constants := compiled.Constants()
	globals := compiled.Globals()

	result1, _, err := multiplyFn.CallWithGlobalsExAndConstants(constants, globals, &Int{Value: 5})
	if err != nil {
		t.Fatal(err)
	}

	if intResult, ok := result1.(*Int); ok {
		if intResult.Value != 10 { // 2 * 5
			t.Errorf("Expected result 10, got %d", intResult.Value)
		}
	} else {
		t.Errorf("Expected Int result, got %T", result1)
	}

	// Test with modified globals (multiplier = 10)
	modifiedGlobals := make([]Object, len(globals))
	copy(modifiedGlobals, globals)
	modifiedGlobals[0] = &Int{Value: 10}

	result2, _, err := multiplyFn.CallWithGlobalsExAndConstants(constants, modifiedGlobals, &Int{Value: 5})
	if err != nil {
		t.Fatal(err)
	}

	if intResult, ok := result2.(*Int); ok {
		if intResult.Value != 50 { // 10 * 5
			t.Errorf("Expected result 50, got %d", intResult.Value)
		}
	} else {
		t.Errorf("Expected Int result, got %T", result2)
	}
}
