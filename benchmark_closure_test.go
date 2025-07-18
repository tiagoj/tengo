package tengo_test

import (
	"testing"

	"github.com/d5/tengo/v2"
)

// BenchmarkClosureInlineExecution benchmarks closure execution performance inline in script.
func BenchmarkClosureInlineExecution(b *testing.B) {
	script := tengo.NewScript([]byte(`
		global_var := 10
		
		make_adder := func(x) {
			return func(y) {
				return x + y + global_var
			}
		}
		
		add_five := make_adder(5)
		for i := 0; i < 1000; i++ {
			result := add_five(i)
		}
	`))

	compiled, err := script.Compile()
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = compiled.Run()
		if err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}

// BenchmarkClosureGoAPIExecution benchmarks closure execution via Go API using ExecutionContext.
func BenchmarkClosureGoAPIExecution(b *testing.B) {
	script := tengo.NewScript([]byte(`
		global_var := 10
		
		make_adder := func(x) {
			return func(y) {
				return x + y + global_var
			}
		}
		
		add_five := make_adder(5)
	`))

	compiled, err := script.Compile()
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	err = compiled.Run()
	if err != nil {
		b.Fatalf("run error: %v", err)
	}

	// Get the closure
	addFiveVar := compiled.Get("add_five")
	addFiveFn := addFiveVar.Value().(*tengo.CompiledFunction)

	// Create execution context
	ctx := tengo.NewExecutionContext(compiled)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < 1000; i++ {
			_, err := ctx.Call(addFiveFn, &tengo.Int{Value: int64(i)})
			if err != nil {
				b.Fatalf("call error: %v", err)
			}
		}
	}
}

// BenchmarkClosureDirectAPIExecution benchmarks closure execution via direct API call.
func BenchmarkClosureDirectAPIExecution(b *testing.B) {
	script := tengo.NewScript([]byte(`
		global_var := 10
		
		make_adder := func(x) {
			return func(y) {
				return x + y + global_var
			}
		}
		
		add_five := make_adder(5)
	`))

	compiled, err := script.Compile()
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	err = compiled.Run()
	if err != nil {
		b.Fatalf("run error: %v", err)
	}

	// Get the closure
	addFiveVar := compiled.Get("add_five")
	addFiveFn := addFiveVar.Value().(*tengo.CompiledFunction)

	// Get constants and globals
	constants := compiled.Constants()
	globals := compiled.Globals()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < 1000; i++ {
			_, _, err := addFiveFn.CallWithGlobalsExAndConstants(
				constants, globals, &tengo.Int{Value: int64(i)})
			if err != nil {
				b.Fatalf("call error: %v", err)
			}
		}
	}
}

// BenchmarkClosureIsolatedContext benchmarks closure execution with isolated contexts.
func BenchmarkClosureIsolatedContext(b *testing.B) {
	script := tengo.NewScript([]byte(`
		global_var := 10
		
		make_adder := func(x) {
			return func(y) {
				return x + y + global_var
			}
		}
		
		add_five := make_adder(5)
	`))

	compiled, err := script.Compile()
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	err = compiled.Run()
	if err != nil {
		b.Fatalf("run error: %v", err)
	}

	// Get the closure
	addFiveVar := compiled.Get("add_five")
	addFiveFn := addFiveVar.Value().(*tengo.CompiledFunction)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		// Create isolated context for each iteration
		ctx := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
		for i := 0; i < 1000; i++ {
			_, err := ctx.Call(addFiveFn, &tengo.Int{Value: int64(i)})
			if err != nil {
				b.Fatalf("call error: %v", err)
			}
		}
	}
}

// BenchmarkNestedClosures benchmarks deeply nested closure execution.
func BenchmarkNestedClosures(b *testing.B) {
	script := tengo.NewScript([]byte(`
		global_base := 1000
		
		make_nested := func(level1) {
			return func(level2) {
				return func(level3) {
					return func(level4) {
						return global_base + level1 + level2 + level3 + level4
					}
				}
			}
		}
		
		nested_fn := make_nested(100)
	`))

	compiled, err := script.Compile()
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	err = compiled.Run()
	if err != nil {
		b.Fatalf("run error: %v", err)
	}

	// Get the nested closure
	nestedFnVar := compiled.Get("nested_fn")
	nestedFn := nestedFnVar.Value().(*tengo.CompiledFunction)

	// Create execution context
	ctx := tengo.NewExecutionContext(compiled)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < 100; i++ {
			// Call nested closure chain
			level2Result, err := ctx.Call(nestedFn, &tengo.Int{Value: 200})
			if err != nil {
				b.Fatalf("call error: %v", err)
			}
			level2Fn := level2Result.(*tengo.CompiledFunction)

			level3Result, err := ctx.Call(level2Fn, &tengo.Int{Value: 300})
			if err != nil {
				b.Fatalf("call error: %v", err)
			}
			level3Fn := level3Result.(*tengo.CompiledFunction)

			_, err = ctx.Call(level3Fn, &tengo.Int{Value: 400})
			if err != nil {
				b.Fatalf("call error: %v", err)
			}
		}
	}
}

// BenchmarkClosureComplexDataTypes benchmarks closure execution with complex data types.
func BenchmarkClosureComplexDataTypes(b *testing.B) {
	script := tengo.NewScript([]byte(`
		global_map := {a: 1, b: 2, c: 3}
		global_array := [1, 2, 3, 4, 5]
		
		make_processor := func(local_data) {
			return func(param) {
				return {
					global_map: global_map,
					global_array: global_array,
					local_data: local_data,
					param: param
				}
			}
		}
		
		processor := make_processor({x: 10, y: 20})
	`))

	compiled, err := script.Compile()
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	err = compiled.Run()
	if err != nil {
		b.Fatalf("run error: %v", err)
	}

	// Get the processor closure
	processorVar := compiled.Get("processor")
	processorFn := processorVar.Value().(*tengo.CompiledFunction)

	// Create execution context
	ctx := tengo.NewExecutionContext(compiled)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < 1000; i++ {
			_, err := ctx.Call(processorFn, &tengo.String{Value: "test"})
			if err != nil {
				b.Fatalf("call error: %v", err)
			}
		}
	}
}

// BenchmarkBasicFunctionCall benchmarks basic function calls (baseline).
func BenchmarkBasicFunctionCall(b *testing.B) {
	script := tengo.NewScript([]byte(`
		add := func(x, y) {
			return x + y
		}
		
		for i := 0; i < 1000; i++ {
			result := add(i, 5)
		}
	`))

	compiled, err := script.Compile()
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = compiled.Run()
		if err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}
