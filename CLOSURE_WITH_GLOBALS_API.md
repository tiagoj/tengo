# Closure with Globals API Documentation

## Overview

The Closure with Globals API solves a specific limitation in the original Tengo implementation: **closures executed from Go code could not access global variables**, while closures executed inline within Tengo scripts worked perfectly.

### The Problem (Original Tengo)

In the original Tengo implementation:
- ✅ **Inline execution**: Closures called within Tengo scripts had full access to globals
- ❌ **Go code execution**: Closures extracted and called from Go code lost access to globals

This created a significant limitation when building Go applications that needed to call Tengo closures with access to global state.

### The Solution (This Fork)

This enhanced fork provides APIs that enable calling Tengo closures from Go code while preserving their complete execution context, including access to global variables and constants.

## Before vs After Comparison

### ❌ Original Tengo - What Didn't Work

```go
package main

import (
    "fmt"
    "github.com/d5/tengo/v2"
)

func main() {
    script := tengo.NewScript([]byte(`
        counter := 0  // Global variable
        
        increment := func(step) {
            counter += step    // ❌ This won't work when called from Go!
            return counter
        }
        
        // ✅ This works fine - inline execution within Tengo
        inline_result := increment(5)  // counter becomes 5
    `))
    
    compiled, _ := script.RunContext(context.Background())
    
    // Extract the closure to call from Go
    incrementVar := compiled.Get("increment")
    incrementFn := incrementVar.Value().(*tengo.CompiledFunction)
    
    // ❌ This call will fail - closure can't access 'counter'
    result, err := incrementFn.Call(&tengo.Int{Value: 10})
    if err != nil {
        fmt.Println("Error:", err) // Runtime error - undefined variable 'counter'
    }
}
```

### ✅ Enhanced Fork - What Now Works

```go
package main

import (
    "fmt"
    "github.com/d5/tengo/v2"
)

func main() {
    script := tengo.NewScript([]byte(`
        counter := 0  // Global variable
        
        increment := func(step) {
            counter += step    // ✅ This now works from Go too!
            return counter
        }
        
        // ✅ This still works - inline execution within Tengo
        inline_result := increment(5)  // counter becomes 5
    `))
    
    compiled, _ := script.RunContext(context.Background())
    
    // Create execution context to preserve globals
    ctx := tengo.NewExecutionContext(compiled)
    
    // Extract the closure to call from Go
    incrementVar := compiled.Get("increment")
    incrementFn := incrementVar.Value().(*tengo.CompiledFunction)
    
    // ✅ This now works perfectly - closure can access 'counter'
    result, err := ctx.Call(incrementFn, &tengo.Int{Value: 10})
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result.(*tengo.Int).Value) // 15 (5 + 10)
    }
    
    // ✅ Global state is properly maintained
    counterVar := compiled.Get("counter")
    fmt.Println("Final counter:", counterVar.Value().(*tengo.Int).Value) // 15
}
```

## Key Concepts

### ExecutionContext
An `ExecutionContext` bundles together the constants, globals, and source compilation context needed to execute closures correctly from Go code.

### Context Isolation
The API provides isolated execution contexts that prevent race conditions when calling closures from multiple goroutines.

### Backward Compatibility
All existing APIs continue to work unchanged. The new functionality is additive.

## API Reference

### ExecutionContext Type

```go
type ExecutionContext struct {
    // Private fields - access via methods
}
```

The `ExecutionContext` provides thread-safe access to closure execution with proper context preservation.

### Factory Methods

#### NewExecutionContext
```go
func NewExecutionContext(compiled *Compiled) *ExecutionContext
```

Creates a new execution context from a compiled script.

**Parameters:**
- `compiled`: A compiled Tengo script

**Returns:**
- `*ExecutionContext`: A new execution context with the script's constants and globals

**Example:**
```go
script := tengo.NewScript([]byte(`
    global_var := 42
    my_closure := func(x) { return x + global_var }
`))
compiled, _ := script.Compile()
compiled.Run()

ctx := tengo.NewExecutionContext(compiled)
```

#### WithGlobals
```go
func (ec *ExecutionContext) WithGlobals(globals []Object) *ExecutionContext
```

Creates a new execution context with custom globals.

**Parameters:**
- `globals`: Array of objects to use as global variables

**Returns:**
- `*ExecutionContext`: New execution context with the specified globals

**Example:**
```go
customGlobals := []tengo.Object{
    &tengo.Int{Value: 100}, // global_var = 100
}
customCtx := ctx.WithGlobals(customGlobals)
```

#### WithIsolatedGlobals
```go
func (ec *ExecutionContext) WithIsolatedGlobals() *ExecutionContext
```

Creates a new execution context with isolated copies of the current globals.

**Returns:**
- `*ExecutionContext`: New execution context with isolated globals

**Example:**
```go
isolatedCtx := ctx.WithIsolatedGlobals()
// Changes to globals in isolatedCtx won't affect other contexts
```

### Execution Methods

#### Call
```go
func (ec *ExecutionContext) Call(fn *CompiledFunction, args ...Object) (Object, error)
```

Calls a compiled function with the execution context.

**Parameters:**
- `fn`: The compiled function to call
- `args`: Arguments to pass to the function

**Returns:**
- `Object`: The result of the function call
- `error`: Any error that occurred during execution

**Example:**
```go
closureVar := compiled.Get("my_closure")
closureFn := closureVar.Value().(*tengo.CompiledFunction)

result, err := ctx.Call(closureFn, &tengo.Int{Value: 10})
if err != nil {
    // Handle error
}
fmt.Println(result.(*tengo.Int).Value) // 52 (10 + 42)
```

#### CallEx
```go
func (ec *ExecutionContext) CallEx(fn *CompiledFunction, args ...Object) (Object, []Object, error)
```

Calls a compiled function and returns both the result and updated globals.

**Parameters:**
- `fn`: The compiled function to call
- `args`: Arguments to pass to the function

**Returns:**
- `Object`: The result of the function call
- `[]Object`: Updated globals after the function call
- `error`: Any error that occurred during execution

**Example:**
```go
result, updatedGlobals, err := ctx.CallEx(closureFn, &tengo.Int{Value: 10})
if err != nil {
    // Handle error
}
// Use result and updatedGlobals
```

### Utility Methods

#### Constants
```go
func (ec *ExecutionContext) Constants() []Object
```

Returns a copy of the constants array.

**Returns:**
- `[]Object`: Copy of the constants array

#### Globals
```go
func (ec *ExecutionContext) Globals() []Object
```

Returns a copy of the globals array.

**Returns:**
- `[]Object`: Copy of the globals array

#### Source
```go
func (ec *ExecutionContext) Source() *Compiled
```

Returns the original compiled object.

**Returns:**
- `*Compiled`: The source compiled object

### Direct API Methods

#### CallWithGlobalsExAndConstants
```go
func (fn *CompiledFunction) CallWithGlobalsExAndConstants(
    constants []Object, 
    globals []Object, 
    args ...Object,
) (Object, []Object, error)
```

Directly calls a compiled function with explicit constants and globals.

**Parameters:**
- `constants`: Constants array from the original compilation
- `globals`: Globals array for the execution context
- `args`: Arguments to pass to the function

**Returns:**
- `Object`: The result of the function call
- `[]Object`: Updated globals after the function call
- `error`: Any error that occurred during execution

**Example:**
```go
constants := compiled.Constants()
globals := compiled.Globals()

result, updatedGlobals, err := closureFn.CallWithGlobalsExAndConstants(
    constants, globals, &tengo.Int{Value: 10})
```

## Error Handling

The API provides specific error types for different failure scenarios:

### ErrMissingExecutionContext
Returned when required execution context is missing.

### ErrInvalidConstantsArray
Returned when the constants array is invalid.

### ErrInvalidGlobalsArray
Returned when the globals array is invalid.

## Thread Safety

All `ExecutionContext` methods are thread-safe and can be called concurrently from multiple goroutines. Use `WithIsolatedGlobals()` to ensure complete isolation between concurrent executions.

## Performance Characteristics

- **Inline Execution**: ~151,283 ns/op (6,612 ops/sec)
- **Go API Execution**: ~11,075,277 ns/op (90 ops/sec)
- **Memory Usage**: Go API uses ~790x more memory than inline execution
- **ExecutionContext Overhead**: <1% between different Go API methods

The performance trade-off is expected for isolated execution environments that provide safety guarantees.

## Best Practices

1. **Reuse ExecutionContext**: Create once and reuse for multiple calls
2. **Use Isolated Contexts**: Use `WithIsolatedGlobals()` for concurrent execution
3. **Check Errors**: Always check for errors when calling closures
4. **Consider Performance**: For high-performance scenarios, prefer inline execution
5. **Thread Safety**: Use appropriate context isolation for your concurrency needs

## Migration from Legacy API

The new API is fully backward compatible. Existing code continues to work unchanged. To migrate to the new context-aware execution:

```go
// Old way (still works)
result, err := closureFn.Call(args...)

// New way (with proper context)
ctx := tengo.NewExecutionContext(compiled)
result, err := ctx.Call(closureFn, args...)
```

## See Also

- [Usage Examples](CLOSURE_WITH_GLOBALS_EXAMPLES.md)
- [Migration Guide](CLOSURE_WITH_GLOBALS_MIGRATION.md)
- [Performance Analysis](PERFORMANCE_BENCHMARK_RESULTS.md)
