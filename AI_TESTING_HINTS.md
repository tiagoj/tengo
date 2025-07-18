# AI Testing Hints for Tengo Development

## Common Testing Mistakes and Solutions

### 1. Testing Framework Usage

#### ❌ Common Mistakes:
- Using `require.NoError(b, err)` with `*testing.B` in benchmarks
- Assuming `require.Contains()` exists when it doesn't
- Using `compiled.Get("function").(*tengo.CompiledFunction)` directly without checking the Variable wrapper

#### ✅ Correct Approaches:

**For Benchmarks (`*testing.B`):**
```go
// WRONG
require.NoError(b, err)

// CORRECT
if err != nil {
    b.Fatal(err)
}
```

**For Regular Tests (`*testing.T`):**
```go
// CORRECT
require.NoError(t, err)
```

**For String Containment Checks:**
```go
// WRONG (doesn't exist)
require.Contains(t, err.Error(), "expected text")

// CORRECT
require.Error(t, err)
if !strings.Contains(err.Error(), "expected text") {
    t.Errorf("Expected error to contain 'expected text', got: %v", err)
}

// OR use standard testing
require.Error(t, err)
if err != nil && !strings.Contains(err.Error(), "expected text") {
    t.Errorf("Expected error to contain 'expected text', got: %v", err)
}
```

### 2. Tengo-Specific Testing Patterns

#### Function Extraction from Compiled Scripts:
```go
// WRONG
fn := compiled.Get("function_name").(*tengo.CompiledFunction)

// CORRECT
fnVar := compiled.Get("function_name")
require.NotNil(t, fnVar)
fn, ok := fnVar.Value().(*tengo.CompiledFunction)
require.True(t, ok)
```

#### Available require Package Methods:
Based on `/require/require.go`, available methods are:
- `NoError(t, err, msg...)`
- `Error(t, err, msg...)`
- `Nil(t, v, msg...)`
- `NotNil(t, v, msg...)`
- `True(t, v, msg...)`
- `False(t, v, msg...)`
- `Equal(t, expected, actual, msg...)`
- `IsType(t, expected, actual, msg...)`
- `Fail(t, msg...)`

**NOT AVAILABLE:**
- `Contains()` - must implement manually
- `Greater()`, `Less()` - must implement manually
- `Len()` - must implement manually

### 3. Testing Best Practices for Tengo

#### Test Structure:
```go
func TestFeature_Scenario(t *testing.T) {
    // 1. Setup
    script := tengo.NewScript([]byte(`...`))
    compiled, err := script.Compile()
    require.NoError(t, err)
    
    err = compiled.Run()
    require.NoError(t, err)
    
    // 2. Test execution
    ctx := tengo.NewExecutionContext(compiled)
    
    // 3. Function extraction (safe pattern)
    fnVar := compiled.Get("function_name")
    require.NotNil(t, fnVar)
    fn, ok := fnVar.Value().(*tengo.CompiledFunction)
    require.True(t, ok)
    
    // 4. Test calls
    result, err := ctx.Call(fn, args...)
    require.NoError(t, err)
    
    // 5. Assertions
    require.Equal(t, expectedValue, result.(*tengo.Int).Value)
}
```

#### Error Testing Pattern:
```go
func TestError_Scenario(t *testing.T) {
    // Setup that should fail
    _, err := someOperation()
    
    // Check error exists
    require.Error(t, err)
    
    // Check error message (manual string check)
    if !strings.Contains(err.Error(), "expected substring") {
        t.Errorf("Expected error to contain 'expected substring', got: %v", err)
    }
    
    // Check error type
    if _, ok := err.(ExpectedErrorType); !ok {
        t.Errorf("Expected error of type %T, got %T", ExpectedErrorType{}, err)
    }
}
```

### 4. Benchmark Testing Patterns

#### Basic Benchmark:
```go
func BenchmarkFeature_Scenario(b *testing.B) {
    // Setup (outside timer)
    script := tengo.NewScript([]byte(`...`))
    compiled, err := script.Compile()
    if err != nil {
        b.Fatal(err)
    }
    
    err = compiled.Run()
    if err != nil {
        b.Fatal(err)
    }
    
    ctx := tengo.NewExecutionContext(compiled)
    fnVar := compiled.Get("function_name")
    if fnVar == nil {
        b.Fatal("function not found")
    }
    fn, ok := fnVar.Value().(*tengo.CompiledFunction)
    if !ok {
        b.Fatal("not a compiled function")
    }
    
    // Reset timer before the measured operation
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := ctx.Call(fn, args...)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

#### Concurrent Benchmark:
```go
func BenchmarkFeature_Concurrent(b *testing.B) {
    // Setup
    // ... setup code ...
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            // Test operation
            _, err := ctx.Call(fn, args...)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

### 5. Integration Testing Patterns

#### End-to-End Test:
```go
func TestIntegration_RealWorldScenario(t *testing.T) {
    // 1. Create realistic script
    script := tengo.NewScript([]byte(`
        // Complex realistic scenario
        state := {value: 0}
        
        export func operation(x) {
            state.value += x
            return state.value
        }
    `))
    
    // 2. Full compilation and execution
    compiled, err := script.Compile()
    require.NoError(t, err)
    
    err = compiled.Run()
    require.NoError(t, err)
    
    // 3. Test multiple operations
    ctx := tengo.NewExecutionContext(compiled)
    
    // 4. Multiple function calls showing stateful behavior
    // ... test multiple scenarios ...
}
```

### 6. Performance Testing Guidelines

#### Performance Expectations:
- Simple function calls: < 100μs per call
- Memory usage: < 10KB per context
- Context creation: Should be fast enough for concurrent use

#### Memory Testing:
```go
func TestMemoryUsage(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping memory test in short mode")
    }
    
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // ... create many objects ...
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    memoryUsed := m2.Alloc - m1.Alloc
    t.Logf("Memory used: %d bytes", memoryUsed)
    
    // Assert reasonable memory usage
    if memoryUsed > expectedLimit {
        t.Errorf("Memory usage too high: %d bytes > %d", memoryUsed, expectedLimit)
    }
}
```

### 7. Common Anti-Patterns to Avoid

#### ❌ Don't:
- Use `require` with `*testing.B`
- Assume methods exist without checking
- Extract functions without type checking
- Forget to check for nil values
- Use `require.Contains()` (doesn't exist)

#### ✅ Do:
- Use `if err != nil { b.Fatal(err) }` in benchmarks
- Always check return values and types
- Use safe function extraction patterns
- Implement string containment checks manually
- Use appropriate error checking for each test type

### 8. Test Organization

#### File Structure:
- `*_test.go` - Unit tests
- `integration_test.go` - Integration tests
- `benchmark_test.go` - Benchmarks (if separate)
- `error_handling_test.go` - Error scenario tests

#### Test Naming:
- `TestFeature_Scenario` - Unit tests
- `TestFeature_Integration` - Integration tests
- `BenchmarkFeature_Scenario` - Benchmarks
- `TestFeature_ErrorHandling` - Error tests

This guide should prevent the common mistakes I've been making and provide reliable patterns for testing Tengo functionality.
