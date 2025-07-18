package tengo_test

import (
	"strings"
	"testing"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/require"
)

// TestClosureWithGlobals_BasicIntegration tests basic closure functionality with globals
func TestClosureWithGlobals_BasicIntegration(t *testing.T) {
	// Test simple closure with global variable access
	script := tengo.NewScript([]byte(`
		global_var := 42
		
		get_global := func() {
			return global_var
		}
		
		modify_global := func(x) {
			global_var = x
			return global_var
		}
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Test through ExecutionContext
	ctx := tengo.NewExecutionContext(compiled)
	
	// Get the functions
	getGlobalVar := compiled.Get("get_global")
	require.NotNil(t, getGlobalVar)
	
	getGlobalFn, ok := getGlobalVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)
	
	// Test basic call
	result, err := ctx.Call(getGlobalFn)
	require.NoError(t, err)
	require.Equal(t, int64(42), result.(*tengo.Int).Value)
	
	// Test with modified global
	modifyGlobalVar := compiled.Get("modify_global")
	require.NotNil(t, modifyGlobalVar)
	
	modifyGlobalFn, ok := modifyGlobalVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)
	
	// Modify global
	result2, err := ctx.Call(modifyGlobalFn, &tengo.Int{Value: 100})
	require.NoError(t, err)
	require.Equal(t, int64(100), result2.(*tengo.Int).Value)
	
	// Verify global was updated
	result3, err := ctx.Call(getGlobalFn)
	require.NoError(t, err)
	require.Equal(t, int64(100), result3.(*tengo.Int).Value)
}

// TestClosureWithGlobals_IsolatedContexts tests isolated execution contexts
func TestClosureWithGlobals_IsolatedContexts(t *testing.T) {
	script := tengo.NewScript([]byte(`
		counter := 0
		
		increment := func() {
			counter += 1
			return counter
		}
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Create two isolated contexts
	ctx1 := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
	ctx2 := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()

	incrementVar := compiled.Get("increment")
	require.NotNil(t, incrementVar)
	
	incrementFn, ok := incrementVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)

	// Call increment in both contexts
	result1, err := ctx1.Call(incrementFn)
	require.NoError(t, err)
	require.Equal(t, int64(1), result1.(*tengo.Int).Value)

	result2, err := ctx2.Call(incrementFn)
	require.NoError(t, err)
	require.Equal(t, int64(1), result2.(*tengo.Int).Value) // Should be 1, not 2

	// Call again in ctx1 - should be 2
	result3, err := ctx1.Call(incrementFn)
	require.NoError(t, err)
	require.Equal(t, int64(2), result3.(*tengo.Int).Value)

	// Call again in ctx2 - should be 2 (isolated from ctx1)
	result4, err := ctx2.Call(incrementFn)
	require.NoError(t, err)
	require.Equal(t, int64(2), result4.(*tengo.Int).Value)
}

// TestClosureWithGlobals_CustomGlobals tests custom globals
func TestClosureWithGlobals_CustomGlobals(t *testing.T) {
	script := tengo.NewScript([]byte(`
		multiplier := 2
		
		multiply := func(x) {
			return multiplier * x
		}
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Create contexts with different globals
	ctx := tengo.NewExecutionContext(compiled)
	
	customGlobals := []tengo.Object{
		&tengo.Int{Value: 10}, // multiplier = 10
	}
	customCtx := ctx.WithGlobals(customGlobals)

	multiplyVar := compiled.Get("multiply")
	require.NotNil(t, multiplyVar)
	
	multiplyFn, ok := multiplyVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)

	// Test with original context (multiplier = 2)
	result1, err := ctx.Call(multiplyFn, &tengo.Int{Value: 5})
	require.NoError(t, err)
	require.Equal(t, int64(10), result1.(*tengo.Int).Value) // 2 * 5

	// Test with custom context (multiplier = 10)
	result2, err := customCtx.Call(multiplyFn, &tengo.Int{Value: 5})
	require.NoError(t, err)
	require.Equal(t, int64(50), result2.(*tengo.Int).Value) // 10 * 5
}

// TestClosureWithGlobals_DirectCallMethod tests direct CallWithGlobalsExAndConstants
func TestClosureWithGlobals_DirectCallMethod(t *testing.T) {
	script := tengo.NewScript([]byte(`
		add := func(x, y) {
			return x + y
		}
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	addVar := compiled.Get("add")
	require.NotNil(t, addVar)
	
	addFn, ok := addVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)

	// Test direct call with constants and globals
	constants := compiled.Constants()
	globals := compiled.Globals()

	result, updatedGlobals, err := addFn.CallWithGlobalsExAndConstants(
		constants, globals, 
		&tengo.Int{Value: 10}, 
		&tengo.Int{Value: 20})
	require.NoError(t, err)
	require.Equal(t, int64(30), result.(*tengo.Int).Value)
	require.NotNil(t, updatedGlobals)
}

// TestClosureWithGlobals_ErrorScenarios tests error handling scenarios
func TestClosureWithGlobals_ErrorScenarios(t *testing.T) {
	script := tengo.NewScript([]byte(`x := 42`))
	compiled, err := script.Compile()
	require.NoError(t, err)
	
	err = compiled.Run()
	require.NoError(t, err)
	
	ctx := tengo.NewExecutionContext(compiled)
	
	// Test calling nil function
	_, err = ctx.Call(nil)
	require.Error(t, err)
	if !strings.Contains(err.Error(), "compiled function") {
		t.Errorf("Expected error to contain 'compiled function', got: %v", err)
	}
	
	// Test function with invalid constants
	invalidFn := &tengo.CompiledFunction{
		Instructions:  []byte{0x01, 0x02},
		NumLocals:     0,
		NumParameters: 0,
		VarArgs:       false,
	}
	
	_, _, err = invalidFn.CallWithGlobalsExAndConstants(nil, nil)
	require.Error(t, err)
	if !strings.Contains(err.Error(), "constants") {
		t.Errorf("Expected error to contain 'constants', got: %v", err)
	}
	
	// Test function with invalid constants array
	invalidConstants := []tengo.Object{&tengo.Int{Value: 1}, nil}
	_, _, err = invalidFn.CallWithGlobalsExAndConstants(invalidConstants, nil)
	require.Error(t, err)
	if !strings.Contains(err.Error(), "invalid constants array") {
		t.Errorf("Expected error to contain 'invalid constants array', got: %v", err)
	}
}
