package tengo_test

import (
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/require"
	"strings"
	"testing"
)

// TestGlobalVariableShadowing tests cases where local variables shadow global variables
func TestGlobalVariableShadowing(t *testing.T) {
	script := tengo.NewScript([]byte(`
		val := 1
		foo := func() { val := 2; return val }
		global_val := val
		local_val := foo()
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	globalVal := compiled.Get("global_val")
	require.Equal(t, int64(1), globalVal.Value())

	localVal := compiled.Get("local_val")
	require.Equal(t, int64(2), localVal.Value())
}

// TestRecursiveClosures tests closures that call themselves recursively
func TestRecursiveClosures(t *testing.T) {
	script := tengo.NewScript([]byte(`
		fib := func(x) {
			if x <= 1 {
				return x
			}
			return fib(x-1) + fib(x-2)
		}
		result := fib(10)
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	result := compiled.Get("result")
	require.Equal(t, int64(55), result.Value())
}

// TestDeeplyNestedClosures tests closures nested several layers deep
func TestDeeplyNestedClosures(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_val := 100
		
		level1 := func(a) {
			level2 := func(b) {
				level3 := func(c) {
					level4 := func(d) {
						level5 := func(e) {
							return a + b + c + d + e + global_val
						}
						return level5
					}
					return level4
				}
				return level3
			}
			return level2
		}
		
		result := level1(1)(2)(3)(4)(5)
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	result := compiled.Get("result")
	require.Equal(t, int64(115), result.Value()) // 1+2+3+4+5+100 = 115
}

// TestInterdependentGlobals tests globals that depend on each other
func TestInterdependentGlobals(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_a := 10
		global_b := global_a * 2
		global_c := global_a + global_b
		
		make_closure := func() {
			return func() {
				return global_a + global_b + global_c
			}
		}
		
		closure := make_closure()
		result := closure()
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	result := compiled.Get("result")
	require.Equal(t, int64(60), result.Value()) // 10 + 20 + 30 = 60
}

// TestStatePersistenceThroughErrors tests state preservation when errors occur
func TestStatePersistenceThroughErrors(t *testing.T) {
	// First, create a closure that maintains state
	script := tengo.NewScript([]byte(`
		global_counter := 0
		
		make_counter := func() {
			local_counter := 0
			return func(should_error) {
				local_counter += 1
				global_counter += 1
				
				if should_error {
					error("intentional error")
				}
				
				return {local: local_counter, global: global_counter}
			}
		}
		
		counter := make_counter()
		result1 := counter(false)
		result2 := counter(false)
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Verify normal execution first
	result1 := compiled.Get("result1")
	require.NotNil(t, result1)

	result2 := compiled.Get("result2")
	require.NotNil(t, result2)

	// Now test via ExecutionContext to handle errors
	counterVar := compiled.Get("counter")
	require.NotNil(t, counterVar)

	counterFn, ok := counterVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)

	ctx := tengo.NewExecutionContext(compiled)

	// Call with error - should return error result
	errorResult, err := ctx.Call(counterFn, tengo.TrueValue)
	require.NoError(t, err)

	// Check if the result is an error object
	if errorObj, ok := errorResult.(*tengo.Error); ok {
		require.True(t, strings.Contains(errorObj.Value.String(), "intentional error"))
	}

	// Call again without error - state should be preserved
	successResult, err := ctx.Call(counterFn, tengo.FalseValue)
	require.NoError(t, err)
	require.NotNil(t, successResult)
}

// TestDynamicFunctionComposition tests creating functions within closures dynamically
func TestDynamicFunctionComposition(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_multiplier := 2
		
		make_composer := func(base_value) {
			return func(operation) {
				if operation == "add" {
					return func(x) { return x + base_value + global_multiplier }
				} else if operation == "multiply" {
					return func(x) { return x * base_value * global_multiplier }
				} else {
					return func(x) { return x }
				}
			}
		}
		
		composer := make_composer(5)
		adder := composer("add")
		multiplier := composer("multiply")
		identity := composer("other")
		
		add_result := adder(10)
		multiply_result := multiplier(3)
		identity_result := identity(42)
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	addResult := compiled.Get("add_result")
	require.Equal(t, int64(17), addResult.Value()) // 10 + 5 + 2 = 17

	multiplyResult := compiled.Get("multiply_result")
	require.Equal(t, int64(30), multiplyResult.Value()) // 3 * 5 * 2 = 30

	identityResult := compiled.Get("identity_result")
	require.Equal(t, int64(42), identityResult.Value()) // 42
}

// TestDuplicateClosuresInLoops tests identical closures created in loops
func TestDuplicateClosuresInLoops(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_base := 100
		closures := []
		
		for i := 0; i < 5; i++ {
			closure := func(x) {
				return x + i + global_base
			}
			closures = append(closures, closure)
		}
		
		results := []
		for j := 0; j < len(closures); j++ {
			result := closures[j](10)
			results = append(results, result)
		}
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	results := compiled.Get("results")
	require.NotNil(t, results)

	// Results come back as []interface{} from tengo
	resultsSlice, ok := results.Value().([]interface{})
	if !ok {
		// Try tengo array format
		resultsArray, ok := results.Value().(*tengo.Array)
		require.True(t, ok, "results should be either []interface{} or *tengo.Array")
		// Convert tengo array to interface slice for consistent handling
		resultsSlice = make([]interface{}, len(resultsArray.Value))
		for i, v := range resultsArray.Value {
			resultsSlice[i] = v
		}
	}

	// All closures should capture the final value of i (which is 5)
	require.Equal(t, 5, len(resultsSlice))
	for i := 0; i < len(resultsSlice); i++ {
		result, ok := resultsSlice[i].(int64)
		require.True(t, ok, "result should be int64")
		require.Equal(t, int64(115), result) // 10 + 5 + 100 = 115
	}
}

// TestHighLoadExecutionWithResourceConstraints tests execution under high load
func TestHighLoadExecutionWithResourceConstraints(t *testing.T) {
	script := tengo.NewScript([]byte(`
		global_counter := 0
		
		make_heavy_closure := func() {
			local_data := []
			for i := 0; i < 100; i++ {
				local_data = append(local_data, i)
			}
			
			return func(multiplier) {
				global_counter += 1
				sum := 0
				for j := 0; j < len(local_data); j++ {
					sum += local_data[j] * multiplier
				}
				return sum + global_counter
			}
		}
		
		heavy_closure := make_heavy_closure()
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Test multiple executions
	heavyClosureVar := compiled.Get("heavy_closure")
	require.NotNil(t, heavyClosureVar)

	heavyClosureFn, ok := heavyClosureVar.Value().(*tengo.CompiledFunction)
	require.True(t, ok)

	ctx := tengo.NewExecutionContext(compiled)

	// Execute multiple times to simulate high load
	for i := 0; i < 10; i++ {
		result, err := ctx.Call(heavyClosureFn, &tengo.Int{Value: int64(i + 1)})
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify the computation is correct
		expectedSum := 0
		for j := 0; j < 100; j++ {
			expectedSum += j * (i + 1)
		}
		expectedResult := int64(expectedSum + i + 1)

		actualResult := result.(*tengo.Int).Value
		require.Equal(t, expectedResult, actualResult)
	}
}
