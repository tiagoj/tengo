package tengo_test

import (
	"strings"
	"testing"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/require"
)

// TestClosureBasicFunctionality tests basic closure functionality in inline scripts
func TestClosureBasicFunctionality(t *testing.T) {
	// Test basic closure with free variables
	script := tengo.NewScript([]byte(`
		make_counter := func() {
			count := 0
			return func() {
				count += 1
				return count
			}
		}
		
		counter := make_counter()
		first := counter()
		second := counter()
		third := counter()
		
		// Test that multiple closures have independent state
		counter2 := make_counter()
		first2 := counter2()
		second2 := counter2()
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Verify the results
	first := compiled.Get("first")
	require.Equal(t, int64(1), first.Value())

	second := compiled.Get("second")
	require.Equal(t, int64(2), second.Value())

	third := compiled.Get("third")
	require.Equal(t, int64(3), third.Value())

	// Verify independent state
	first2 := compiled.Get("first2")
	require.Equal(t, int64(1), first2.Value())

	second2 := compiled.Get("second2")
	require.Equal(t, int64(2), second2.Value())
}

// TestClosureWithGlobalsInline tests closures accessing global variables inline
func TestClosureWithGlobalsInline(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_var := 100
		
		make_adder := func(x) {
			return func(y) {
				return x + y + global_var
			}
		}
		
		add_five := make_adder(5)
		result1 := add_five(10) // should be 5 + 10 + 100 = 115
		
		// Modify global and test again
		global_var = 200
		result2 := add_five(10) // should be 5 + 10 + 200 = 215
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	result1 := compiled.Get("result1")
	require.Equal(t, int64(115), result1.Value())

	result2 := compiled.Get("result2")
	require.Equal(t, int64(215), result2.Value())
}

// TestClosureFromGoAPI tests calling closures from Go API using ExecutionContext
func TestClosureFromGoAPI(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_multiplier := 10
		
		make_multiplier := func(factor) {
			return func(value) {
				return value * factor * global_multiplier
			}
		}
		
		multiply_by_three := make_multiplier(3)
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Get the closure function
	multiplyByThreeVar := compiled.Get("multiply_by_three")
	require.NotNil(t, multiplyByThreeVar)

	multiplyByThreeFn, ok := multiplyByThreeVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)

	// Test via ExecutionContext
	ctx := tengo.NewExecutionContext(compiled)
	result, err := ctx.Call(multiplyByThreeFn, &tengo.Int{Value: 5})
	require.NoError(t, err)
	require.Equal(t, int64(150), result.(*tengo.Int).Value) // 5 * 3 * 10 = 150

	// Test with modified globals (adjust global_multiplier to 20)
	globals := compiled.Globals()
	modifiedGlobals := make([]tengo.Object, len(globals))
	copy(modifiedGlobals, globals)
	// global_multiplier is typically at index 0
	modifiedGlobals[0] = &tengo.Int{Value: 20}
	
	customCtx := ctx.WithGlobals(modifiedGlobals)
	result2, err := customCtx.Call(multiplyByThreeFn, &tengo.Int{Value: 5})
	require.NoError(t, err)
	require.Equal(t, int64(300), result2.(*tengo.Int).Value) // 5 * 3 * 20 = 300
}

// TestClosureWithDifferentTypesOfVariables tests closures with different types of captured variables
func TestClosureWithDifferentTypesOfVariables(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_string := "global"
		global_array := [1, 2, 3]
		global_map := {key: "value"}
		
		make_complex_closure := func(local_string, local_number) {
			local_array := [4, 5, 6]
			
			return func(param) {
				return {
					global_string: global_string,
					global_array: global_array,
					global_map: global_map,
					local_string: local_string,
					local_number: local_number,
					local_array: local_array,
					param: param
				}
			}
		}
		
		complex_closure := make_complex_closure("local", 42)
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Test inline execution first
	script2 := tengo.NewScript([]byte(`
		global_string := "global"
		global_array := [1, 2, 3]
		global_map := {key: "value"}
		
		make_complex_closure := func(local_string, local_number) {
			local_array := [4, 5, 6]
			
			return func(param) {
				return {
					global_string: global_string,
					global_array: global_array,
					global_map: global_map,
					local_string: local_string,
					local_number: local_number,
					local_array: local_array,
					param: param
				}
			}
		}
		
		complex_closure := make_complex_closure("local", 42)
		inline_result := complex_closure("test_param")
	`))

	compiled2, err := script2.Compile()
	require.NoError(t, err)

	err = compiled2.Run()
	require.NoError(t, err)

	inlineResult := compiled2.Get("inline_result")
	require.NotNil(t, inlineResult)
	inlineMap := inlineResult.Object().(*tengo.Map)
	require.Equal(t, "global", inlineMap.Value["global_string"].(*tengo.String).Value)

	// Test via Go API
	complexClosureVar := compiled.Get("complex_closure")
	require.NotNil(t, complexClosureVar)

	complexClosureFn, ok := complexClosureVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)

	ctx := tengo.NewExecutionContext(compiled)
	result, err := ctx.Call(complexClosureFn, &tengo.String{Value: "test_param"})
	require.NoError(t, err)

	resultMap := result.(*tengo.Map)
	require.Equal(t, "global", resultMap.Value["global_string"].(*tengo.String).Value)
	require.Equal(t, "local", resultMap.Value["local_string"].(*tengo.String).Value)
	require.Equal(t, int64(42), resultMap.Value["local_number"].(*tengo.Int).Value)
	require.Equal(t, "test_param", resultMap.Value["param"].(*tengo.String).Value)
}

// TestNestedClosures tests deeply nested closures
func TestNestedClosures(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_base := 1000
		
		make_nested_closure := func(level1) {
			return func(level2) {
				return func(level3) {
					return func(level4) {
						return global_base + level1 + level2 + level3 + level4
					}
				}
			}
		}
		
		nested_fn := make_nested_closure(100)
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Test inline execution
	script2 := tengo.NewScript([]byte(`
		global_base := 1000
		
		make_nested_closure := func(level1) {
			return func(level2) {
				return func(level3) {
					return func(level4) {
						return global_base + level1 + level2 + level3 + level4
					}
				}
			}
		}
		
		nested_fn := make_nested_closure(100)
		inline_result := nested_fn(200)(300)(400)
	`))

	compiled2, err := script2.Compile()
	require.NoError(t, err)

	err = compiled2.Run()
	require.NoError(t, err)

	inlineResult := compiled2.Get("inline_result")
	require.Equal(t, int64(2000), inlineResult.Value()) // 1000 + 100 + 200 + 300 + 400

	// Test via Go API
	nestedFnVar := compiled.Get("nested_fn")
	require.NotNil(t, nestedFnVar)

	nestedFn, ok := nestedFnVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)

	ctx := tengo.NewExecutionContext(compiled)
	
	// Call level 2
	level2Result, err := ctx.Call(nestedFn, &tengo.Int{Value: 200})
	require.NoError(t, err)
	level2Fn := level2Result.(*tengo.CompiledFunction)

	// Call level 3
	level3Result, err := ctx.Call(level2Fn, &tengo.Int{Value: 300})
	require.NoError(t, err)
	level3Fn := level3Result.(*tengo.CompiledFunction)

	// Call level 4
	finalResult, err := ctx.Call(level3Fn, &tengo.Int{Value: 400})
	require.NoError(t, err)
	require.Equal(t, int64(2000), finalResult.(*tengo.Int).Value)
}

// TestClosureWithIsolatedGlobals tests that isolated globals work correctly
func TestClosureWithIsolatedGlobals(t *testing.T) {
	script := tengo.NewScript([]byte(`
		counter := 0
		
		increment := func() {
			counter += 1
			return counter
		}
		
		get_counter := func() {
			return counter
		}
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Get functions
	incrementVar := compiled.Get("increment")
	require.NotNil(t, incrementVar)
	incrementFn := incrementVar.Value().(*tengo.CompiledFunction)

	getCounterVar := compiled.Get("get_counter")
	require.NotNil(t, getCounterVar)
	getCounterFn := getCounterVar.Value().(*tengo.CompiledFunction)

	// Create two isolated contexts
	ctx1 := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
	ctx2 := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()

	// Test that they have independent state
	result1, err := ctx1.Call(incrementFn)
	require.NoError(t, err)
	require.Equal(t, int64(1), result1.(*tengo.Int).Value)

	result2, err := ctx2.Call(incrementFn)
	require.NoError(t, err)
	require.Equal(t, int64(1), result2.(*tengo.Int).Value) // Should be 1, not 2

	// Increment ctx1 again
	result3, err := ctx1.Call(incrementFn)
	require.NoError(t, err)
	require.Equal(t, int64(2), result3.(*tengo.Int).Value)

	// Check that ctx2 is still at 1
	result4, err := ctx2.Call(getCounterFn)
	require.NoError(t, err)
	require.Equal(t, int64(1), result4.(*tengo.Int).Value)
}

// TestClosureWithConstantsAndErrors tests error handling with closures
func TestClosureWithConstantsAndErrors(t *testing.T) {
	script := tengo.NewScript([]byte(`
		error_msg := "Custom error message"
		
		make_error_closure := func(should_error) {
			return func(value) {
				if should_error {
					return error(error_msg + ": " + string(value))
				}
				return value * 2
			}
		}
		
		error_closure := make_error_closure(true)
		success_closure := make_error_closure(false)
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Get closures
	errorClosureVar := compiled.Get("error_closure")
	require.NotNil(t, errorClosureVar)
	errorClosureFn := errorClosureVar.Value().(*tengo.CompiledFunction)

	successClosureVar := compiled.Get("success_closure")
	require.NotNil(t, successClosureVar)
	successClosureFn := successClosureVar.Value().(*tengo.CompiledFunction)

	ctx := tengo.NewExecutionContext(compiled)

	// Test error case
	errorResult, err := ctx.Call(errorClosureFn, &tengo.Int{Value: 123})
	require.NoError(t, err)
	errorObj := errorResult.(*tengo.Error)
	require.True(t, strings.Contains(errorObj.String(), "Custom error message: 123"))

	// Test success case
	successResult, err := ctx.Call(successClosureFn, &tengo.Int{Value: 123})
	require.NoError(t, err)
	require.Equal(t, int64(246), successResult.(*tengo.Int).Value)
}

// TestClosureCompatibilityBetweenInlineAndGoAPI tests that closures behave identically
func TestClosureCompatibilityBetweenInlineAndGoAPI(t *testing.T) {
	// Test that the same closure behaves identically when called inline vs via Go API
	script := tengo.NewScript([]byte(`
		global_factor := 5
		
		make_test_closure := func(multiplier) {
			local_addition := 10
			return func(value) {
				return (value * multiplier + local_addition) * global_factor
			}
		}
		
		test_closure := make_test_closure(3)
		
		// Test inline
		inline_result := test_closure(7) // (7 * 3 + 10) * 5 = 155
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Get inline result
	inlineResult := compiled.Get("inline_result")
	require.Equal(t, int64(155), inlineResult.Value())

	// Test via Go API
	testClosureVar := compiled.Get("test_closure")
	require.NotNil(t, testClosureVar)
	testClosureFn := testClosureVar.Value().(*tengo.CompiledFunction)

	ctx := tengo.NewExecutionContext(compiled)
	goAPIResult, err := ctx.Call(testClosureFn, &tengo.Int{Value: 7})
	require.NoError(t, err)
	require.Equal(t, int64(155), goAPIResult.(*tengo.Int).Value)

	// They should be identical
	require.Equal(t, inlineResult.Value(), goAPIResult.(*tengo.Int).Value)
}

// TestClosureWithDirectAPICall tests the direct CallWithGlobalsExAndConstants method
func TestClosureWithDirectAPICall(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_value := 42
		
		make_direct_closure := func(param) {
			return func(input) {
				return global_value + param + input
			}
		}
		
		direct_closure := make_direct_closure(10)
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Get the closure
	directClosureVar := compiled.Get("direct_closure")
	require.NotNil(t, directClosureVar)
	directClosureFn := directClosureVar.Value().(*tengo.CompiledFunction)

	// Test direct call
	constants := compiled.Constants()
	globals := compiled.Globals()

	result, updatedGlobals, err := directClosureFn.CallWithGlobalsExAndConstants(
		constants, globals, &tengo.Int{Value: 5})
	require.NoError(t, err)
	require.Equal(t, int64(57), result.(*tengo.Int).Value) // 42 + 10 + 5 = 57
	require.NotNil(t, updatedGlobals)
}
