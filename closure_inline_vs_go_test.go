// Package tengo_test demonstrates the differences between calling closures inline vs from Go API.
// For a complete working example, see: examples/closure_calls/main.go
package tengo_test

import (
	"strings"
	"testing"

	"github.com/tiagoj/tengo/v2"
	"github.com/tiagoj/tengo/v2/require"
)

// TestClosureInlineVsGoAPI demonstrates the key differences between calling
// closures inline in Tengo scripts versus calling them from Go using the ExecutionContext API.
//
// Key differences:
// 1. Global variable access: inline calls access current globals, Go API calls can use isolated/custom globals
// 2. State management: inline calls share global state, Go API calls can have independent state
// 3. Error handling: inline errors propagate through script, Go API errors are returned to Go
// 4. Performance: inline calls are faster, Go API calls have overhead of crossing language boundary
func TestClosureInlineVsGoAPI(t *testing.T) {
	// Test script with closures that access global variables
	script := tengo.NewScript([]byte(`
		global_counter := 0
		global_multiplier := 10
		
		// Create a closure that captures global variables
		make_calculator := func(base_value) {
			return func(input) {
				global_counter += 1  // Modify global state
				return (input + base_value) * global_multiplier + global_counter
			}
		}
		
		// Create the calculator function
		calculator := make_calculator(5)
		
		// Test inline execution
		inline_result1 := calculator(3)  // (3 + 5) * 10 + 1 = 81
		inline_result2 := calculator(7)  // (7 + 5) * 10 + 2 = 122
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Verify inline results
	inlineResult1 := compiled.Get("inline_result1")
	require.Equal(t, int64(81), inlineResult1.Value())

	inlineResult2 := compiled.Get("inline_result2")
	require.Equal(t, int64(122), inlineResult2.Value())

	// Verify that global_counter was modified
	globalCounter := compiled.Get("global_counter")
	require.Equal(t, int64(2), globalCounter.Value())

	// Now test the same closure via Go API
	calculatorVar := compiled.Get("calculator")
	require.NotNil(t, calculatorVar)
	calculatorFn := calculatorVar.Value().(*tengo.CompiledFunction)

	// Test 1: Using same execution context (shared globals)
	ctx := tengo.NewExecutionContext(compiled)
	
	goResult1, err := ctx.Call(calculatorFn, &tengo.Int{Value: 3})
	require.NoError(t, err)
	require.Equal(t, int64(83), goResult1.(*tengo.Int).Value) // (3 + 5) * 10 + 3 = 83 (counter continues from 2)

	goResult2, err := ctx.Call(calculatorFn, &tengo.Int{Value: 7})
	require.NoError(t, err)
	require.Equal(t, int64(124), goResult2.(*tengo.Int).Value) // (7 + 5) * 10 + 4 = 124

	// Test 2: Using isolated globals (independent state)
	// Note: isolated globals start with a COPY of current globals, not reset to initial state
	isolatedCtx := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
	
	isolatedResult1, err := isolatedCtx.Call(calculatorFn, &tengo.Int{Value: 3})
	require.NoError(t, err)
	require.Equal(t, int64(83), isolatedResult1.(*tengo.Int).Value) // (3 + 5) * 10 + 3 = 83 (counter starts from copy of 2)

	isolatedResult2, err := isolatedCtx.Call(calculatorFn, &tengo.Int{Value: 7})
	require.NoError(t, err)
	require.Equal(t, int64(124), isolatedResult2.(*tengo.Int).Value) // (7 + 5) * 10 + 4 = 124

	// Test 3: Using custom globals
	customGlobals := []tengo.Object{
		&tengo.Int{Value: 100},  // global_counter = 100
		&tengo.Int{Value: 2},    // global_multiplier = 2
	}
	customCtx := ctx.WithGlobals(customGlobals)
	
	customResult, err := customCtx.Call(calculatorFn, &tengo.Int{Value: 3})
	require.NoError(t, err)
	require.Equal(t, int64(117), customResult.(*tengo.Int).Value) // (3 + 5) * 2 + 101 = 117
}

// TestClosureErrorHandlingInlineVsGoAPI demonstrates how error handling differs
// between inline calls and Go API calls
func TestClosureErrorHandlingInlineVsGoAPI(t *testing.T) {
	// Test script with a closure that can return errors
	script := tengo.NewScript([]byte(`
		global_error_threshold := 10
		
		make_validator := func(min_value) {
			return func(input) {
				if input < min_value {
					return error("Value too small: " + string(input) + " < " + string(min_value))
				}
				if input > global_error_threshold {
					return error("Value too large: " + string(input) + " > " + string(global_error_threshold))
				}
				return input * 2
			}
		}
		
		validator := make_validator(5)
		
		// Test inline - these would cause script errors if uncommented
		// inline_error := validator(3)  // Would cause script to fail
		inline_success := validator(7)  // Should work: 7 * 2 = 14
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Verify inline success case
	inlineSuccess := compiled.Get("inline_success")
	require.Equal(t, int64(14), inlineSuccess.Value())

	// Now test the same closure via Go API - errors are returned to Go
	validatorVar := compiled.Get("validator")
	require.NotNil(t, validatorVar)
	validatorFn := validatorVar.Value().(*tengo.CompiledFunction)

	ctx := tengo.NewExecutionContext(compiled)

	// Test error case - too small
	errorResult1, err := ctx.Call(validatorFn, &tengo.Int{Value: 3})
	require.NoError(t, err) // No Go error - Tengo error is returned as object
	errorObj1 := errorResult1.(*tengo.Error)
	errorStr1 := errorObj1.String()
	require.True(t, strings.Contains(errorStr1, "Value too small: 3 < 5"), "Error message should contain 'Value too small: 3 < 5', got: %s", errorStr1)

	// Test error case - too large
	errorResult2, err := ctx.Call(validatorFn, &tengo.Int{Value: 15})
	require.NoError(t, err) // No Go error - Tengo error is returned as object
	errorObj2 := errorResult2.(*tengo.Error)
	errorStr2 := errorObj2.String()
	require.True(t, strings.Contains(errorStr2, "Value too large: 15 > 10"), "Error message should contain 'Value too large: 15 > 10', got: %s", errorStr2)

	// Test success case
	successResult, err := ctx.Call(validatorFn, &tengo.Int{Value: 7})
	require.NoError(t, err)
	require.Equal(t, int64(14), successResult.(*tengo.Int).Value)
}

// TestClosureComplexStateManagement demonstrates complex state management differences
func TestClosureComplexStateManagement(t *testing.T) {
	// Test script with multiple closures sharing state
	script := tengo.NewScript([]byte(`
		shared_state := {
			count: 0,
			history: [],
			multiplier: 3
		}
		
		make_counter := func(increment) {
			return func(label) {
				shared_state.count += increment
				shared_state.history = append(shared_state.history, {
					count: shared_state.count,
					label: label,
					result: shared_state.count * shared_state.multiplier
				})
				return shared_state.count * shared_state.multiplier
			}
		}
		
		make_resetter := func() {
			return func() {
				shared_state.count = 0
				shared_state.history = []
				return "reset"
			}
		}
		
		counter := make_counter(1)
		fast_counter := make_counter(5)
		resetter := make_resetter()
		
		// Test inline execution
		inline_result1 := counter("first")      // count=1, result=3
		inline_result2 := fast_counter("fast")  // count=6, result=18
		inline_result3 := counter("second")     // count=7, result=21
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Verify inline results
	inlineResult1 := compiled.Get("inline_result1")
	require.Equal(t, int64(3), inlineResult1.Value())

	inlineResult2 := compiled.Get("inline_result2") 
	require.Equal(t, int64(18), inlineResult2.Value())

	inlineResult3 := compiled.Get("inline_result3")
	require.Equal(t, int64(21), inlineResult3.Value())

	// Check shared state
	sharedState := compiled.Get("shared_state")
	stateMap := sharedState.Object().(*tengo.Map)
	require.Equal(t, int64(7), stateMap.Value["count"].(*tengo.Int).Value)

	// Get closures for Go API testing
	counterVar := compiled.Get("counter")
	counterFn := counterVar.Value().(*tengo.CompiledFunction)
	
	resetterVar := compiled.Get("resetter")
	resetterFn := resetterVar.Value().(*tengo.CompiledFunction)

	// Test with shared execution context (continues from existing state)
	ctx := tengo.NewExecutionContext(compiled)
	
	goResult1, err := ctx.Call(counterFn, &tengo.String{Value: "from_go"})
	require.NoError(t, err)
	require.Equal(t, int64(24), goResult1.(*tengo.Int).Value) // count=8, 8*3=24

	// Test with isolated execution context (independent state)
	// Note: isolated context copies current state, so count starts at 7+1 (from previous go call), not 0
	isolatedCtx := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
	
	isolatedResult1, err := isolatedCtx.Call(counterFn, &tengo.String{Value: "isolated"})
	require.NoError(t, err)
	require.Equal(t, int64(27), isolatedResult1.(*tengo.Int).Value) // count=9, 9*3=27 (copies state after first go call)

	// Reset in isolated context
	_, err = isolatedCtx.Call(resetterFn)
	require.NoError(t, err)

	isolatedResult2, err := isolatedCtx.Call(counterFn, &tengo.String{Value: "after_reset"})
	require.NoError(t, err)
	require.Equal(t, int64(3), isolatedResult2.(*tengo.Int).Value) // count=1, 1*3=3 (reset worked)

	// Verify original context still has its state
	goResult2, err := ctx.Call(counterFn, &tengo.String{Value: "still_there"})
	require.NoError(t, err)
	require.Equal(t, int64(27), goResult2.(*tengo.Int).Value) // count=9, 9*3=27
}

// TestClosurePerformanceComparison demonstrates performance characteristics
// Note: This is more of a conceptual test - actual performance would need benchmarks
func TestClosurePerformanceComparison(t *testing.T) {
	// Test script with a simple closure
	script := tengo.NewScript([]byte(`
		make_adder := func(x) {
			return func(y) {
				return x + y
			}
		}
		
		add_ten := make_adder(10)
		
		// Inline calls - direct VM execution
		inline_results := []
		for i := 0; i < 100; i++ {
			inline_results = append(inline_results, add_ten(i))
		}
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Verify inline results
	inlineResults := compiled.Get("inline_results")
	resultsArray := inlineResults.Object().(*tengo.Array)
	require.Equal(t, 100, len(resultsArray.Value))
	require.Equal(t, int64(10), resultsArray.Value[0].(*tengo.Int).Value) // 10 + 0
	require.Equal(t, int64(109), resultsArray.Value[99].(*tengo.Int).Value) // 10 + 99

	// Go API calls - require marshaling and context switching
	addTenVar := compiled.Get("add_ten")
	addTenFn := addTenVar.Value().(*tengo.CompiledFunction)
	
	ctx := tengo.NewExecutionContext(compiled)
	
	goResults := make([]tengo.Object, 100)
	for i := 0; i < 100; i++ {
		result, err := ctx.Call(addTenFn, &tengo.Int{Value: int64(i)})
		require.NoError(t, err)
		goResults[i] = result
	}

	// Verify Go API results match inline results
	require.Equal(t, int64(10), goResults[0].(*tengo.Int).Value)
	require.Equal(t, int64(109), goResults[99].(*tengo.Int).Value)

	// Both approaches should produce identical results
	for i := 0; i < 100; i++ {
		expected := resultsArray.Value[i].(*tengo.Int).Value
		actual := goResults[i].(*tengo.Int).Value
		require.Equal(t, expected, actual, "Results should be identical at index %d", i)
	}
}
