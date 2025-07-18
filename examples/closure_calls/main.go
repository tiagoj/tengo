package main

import (
	"fmt"
	"log"

	"github.com/d5/tengo/v2"
)

func main() {
	// Example demonstrating the difference between calling closures inline vs from Go
	code := `
	global_counter := 0
	
	// Create a closure that accesses and modifies global state
	make_incrementer := func(step) {
		return func() {
			global_counter += step
			return global_counter
		}
	}
	
	incrementer := make_incrementer(5)
	
	// Inline calls within the script
	inline_result1 := incrementer()  // global_counter = 5
	inline_result2 := incrementer()  // global_counter = 10
	`

	script := tengo.NewScript([]byte(code))
	compiled, err := script.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// Run the script to execute inline calls
	if err := compiled.Run(); err != nil {
		log.Fatal(err)
	}

	// Check results of inline execution
	inlineResult1 := compiled.Get("inline_result1")
	inlineResult2 := compiled.Get("inline_result2")
	globalCounter := compiled.Get("global_counter")

	fmt.Println("=== Inline Execution ===")
	fmt.Printf("Inline result 1: %d\n", inlineResult1.Value())
	fmt.Printf("Inline result 2: %d\n", inlineResult2.Value())
	fmt.Printf("Global counter after inline calls: %d\n", globalCounter.Value())
	fmt.Println()

	// Now call the same closure from Go using ExecutionContext
	incrementerVar := compiled.Get("incrementer")
	incrementerFn := incrementerVar.Value().(*tengo.CompiledFunction)

	fmt.Println("=== Go API Calls (Shared Context) ===")
	
	// Using shared execution context (continues from existing state)
	ctx := tengo.NewExecutionContext(compiled)
	
	goResult1, err := ctx.Call(incrementerFn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Go result 1: %d\n", goResult1.(*tengo.Int).Value) // Should be 15

	goResult2, err := ctx.Call(incrementerFn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Go result 2: %d\n", goResult2.(*tengo.Int).Value) // Should be 20
	fmt.Println()

	fmt.Println("=== Go API Calls (Isolated Context) ===")
	
	// Using isolated execution context (independent copy of globals)
	isolatedCtx := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
	
	isolatedResult1, err := isolatedCtx.Call(incrementerFn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Isolated result 1: %d\n", isolatedResult1.(*tengo.Int).Value) // Should be 25 (starts from current state copy)

	isolatedResult2, err := isolatedCtx.Call(incrementerFn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Isolated result 2: %d\n", isolatedResult2.(*tengo.Int).Value) // Should be 30
	fmt.Println()

	fmt.Println("=== Go API Calls (Custom Globals) ===")
	
	// Using custom globals (reset counter to specific value)
	customGlobals := []tengo.Object{
		&tengo.Int{Value: 100}, // Reset global_counter to 100
	}
	customCtx := ctx.WithGlobals(customGlobals)
	
	customResult1, err := customCtx.Call(incrementerFn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Custom result 1: %d\n", customResult1.(*tengo.Int).Value) // Should be 105

	customResult2, err := customCtx.Call(incrementerFn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Custom result 2: %d\n", customResult2.(*tengo.Int).Value) // Should be 110
	fmt.Println()

	// Show final state of original context
	finalState := ctx.Globals()[0].(*tengo.Int).Value
	fmt.Printf("Final state of shared context: %d\n", finalState)
}
