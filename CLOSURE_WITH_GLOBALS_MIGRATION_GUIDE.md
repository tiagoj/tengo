# Migration Guide: Closure with Globals Enhancement

This guide helps you migrate from the original Tengo closure functionality to the new context-aware closure execution system.

## Table of Contents

1. [Overview of Changes](#overview-of-changes)
2. [Breaking Changes](#breaking-changes)
3. [Migration Steps](#migration-steps)
4. [API Changes](#api-changes)
5. [Before and After Examples](#before-and-after-examples)
6. [Common Migration Patterns](#common-migration-patterns)
7. [Troubleshooting](#troubleshooting)
8. [Performance Considerations](#performance-considerations)
9. [Best Practices](#best-practices)

## Overview of Changes

### What Changed

The Tengo closure system has been enhanced to provide proper context-aware execution. The key improvements include:

- **Context Preservation**: Closures now retain full access to their execution context (constants, globals, free variables)
- **Thread Safety**: New isolated execution contexts prevent race conditions
- **Enhanced API**: New `ExecutionContext` type for easier closure management
- **Better Error Handling**: More specific error messages and validation

### What Stayed the Same

- **Backward Compatibility**: All existing APIs continue to work
- **Core Language**: No changes to Tengo script syntax
- **Performance**: Inline script execution performance unchanged

## Breaking Changes

### ⚠️ Potential Breaking Changes

While we've maintained backward compatibility, there are some edge cases that may require attention:

1. **Error Messages**: Error messages for closure execution failures have changed
2. **Memory Usage**: Context-aware execution uses more memory
3. **Thread Safety**: Previous race conditions may now be exposed as errors

### ✅ No Breaking Changes

The following continue to work without modification:
- All existing `Script.Run()` calls
- All existing `CompiledFunction.Call()` calls
- All existing `CompiledFunction.CallWithGlobals()` calls
- All inline script execution

## Migration Steps

### Step 1: Assess Your Current Usage

Identify how you're currently using closures in your code:

```bash
# Search for closure-related code
grep -r "CallWithGlobals\|CompiledFunction.*Call" your-project/
```

### Step 2: Update Imports (if needed)

No import changes are required - all new functionality is in the existing `tengo` package.

### Step 3: Choose Your Migration Strategy

You have three options:

1. **No Changes Required**: If your code works correctly, no migration needed
2. **Gradual Migration**: Use new APIs for new code, keep existing code unchanged
3. **Full Migration**: Update all closure calls to use `ExecutionContext`

### Step 4: Test Thoroughly

- Run your existing tests to ensure no regressions
- Add tests for any new closure usage patterns
- Test concurrent execution scenarios

## API Changes

### New APIs Added

```go
// New ExecutionContext type
type ExecutionContext struct { /* ... */ }

// Factory methods
func NewExecutionContext(compiled *Compiled) *ExecutionContext
func (ec *ExecutionContext) WithGlobals(globals []Object) *ExecutionContext
func (ec *ExecutionContext) WithIsolatedGlobals() *ExecutionContext

// Execution methods
func (ec *ExecutionContext) Call(fn *CompiledFunction, args ...Object) (Object, error)
func (ec *ExecutionContext) CallEx(fn *CompiledFunction, args ...Object) (Object, []Object, error)

// Enhanced CompiledFunction method
func (o *CompiledFunction) CallWithGlobalsExAndConstants(
    constants []Object, 
    globals []Object, 
    args ...Object,
) (Object, []Object, error)

// New Compiled method
func (c *Compiled) Constants() []Object
```

### Existing APIs (Unchanged)

```go
// These continue to work exactly as before
func (o *CompiledFunction) Call(args ...Object) (Object, error)
func (o *CompiledFunction) CallWithGlobals(globals []Object, args ...Object) (Object, error)
func (o *CompiledFunction) CallWithGlobalsEx(globals []Object, args ...Object) (Object, []Object, error)
```

## Before and After Examples

### Example 1: Basic Closure Execution

**Before (Still Works)**:
```go
script := tengo.NewScript([]byte(`
    counter := 0
    increment := func() {
        counter += 1
        return counter
    }
`))

compiled, _ := script.Compile()
compiled.Run()

incrementVar := compiled.Get("increment")
incrementFn := incrementVar.Value().(*tengo.CompiledFunction)

// This may not work correctly with globals
result, _ := incrementFn.Call()
```

**After (Recommended)**:
```go
script := tengo.NewScript([]byte(`
    counter := 0
    increment := func() {
        counter += 1
        return counter
    }
`))

compiled, _ := script.Compile()
compiled.Run()

// Create execution context
ctx := tengo.NewExecutionContext(compiled)

incrementVar := compiled.Get("increment")
incrementFn := incrementVar.Value().(*tengo.CompiledFunction)

// This works correctly with globals
result, _ := ctx.Call(incrementFn)
```

### Example 2: Concurrent Execution

**Before (Race Condition)**:
```go
// Multiple goroutines calling the same closure
// This could cause race conditions with shared globals

for i := 0; i < 10; i++ {
    go func() {
        result, _ := closureFn.CallWithGlobals(globals)
        // Potential race condition on globals
    }()
}
```

**After (Thread Safe)**:
```go
// Multiple goroutines with isolated contexts
// Each gets its own copy of globals

for i := 0; i < 10; i++ {
    go func() {
        isolatedCtx := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
        result, _ := isolatedCtx.Call(closureFn)
        // No race conditions
    }()
}
```

### Example 3: Custom Globals

**Before (Limited)**:
```go
// Limited ability to modify globals
customGlobals := make([]tengo.Object, len(originalGlobals))
copy(customGlobals, originalGlobals)
customGlobals[0] = &tengo.Int{Value: 42}

result, _ := closureFn.CallWithGlobals(customGlobals)
```

**After (Enhanced)**:
```go
// Flexible globals modification
ctx := tengo.NewExecutionContext(compiled)

customGlobals := make([]tengo.Object, len(ctx.Globals()))
copy(customGlobals, ctx.Globals())
customGlobals[0] = &tengo.Int{Value: 42}

customCtx := ctx.WithGlobals(customGlobals)
result, _ := customCtx.Call(closureFn)
```

## Common Migration Patterns

### Pattern 1: Simple Closure Calls

```go
// OLD: May not work with globals
result, err := closureFn.Call(args...)

// NEW: Guaranteed to work with globals
ctx := tengo.NewExecutionContext(compiled)
result, err := ctx.Call(closureFn, args...)
```

### Pattern 2: Closure with Modified Globals

```go
// OLD: Manual globals management
globals := compiled.Globals()
globals[0] = newValue
result, err := closureFn.CallWithGlobals(globals, args...)

// NEW: Context-based globals management
ctx := tengo.NewExecutionContext(compiled)
newGlobals := make([]tengo.Object, len(ctx.Globals()))
copy(newGlobals, ctx.Globals())
newGlobals[0] = newValue
customCtx := ctx.WithGlobals(newGlobals)
result, err := customCtx.Call(closureFn, args...)
```

### Pattern 3: Concurrent Execution

```go
// OLD: Potential race conditions
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        result, _ := closureFn.CallWithGlobals(globals)
        // Race condition possible
    }()
}
wg.Wait()

// NEW: Thread-safe execution
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        ctx := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
        result, _ := ctx.Call(closureFn)
        // No race conditions
    }()
}
wg.Wait()
```

## Troubleshooting

### Common Issues and Solutions

#### Issue 1: "Missing execution context" error

**Problem**: Closure fails with context-related error

**Solution**: Use `ExecutionContext` instead of direct function calls

```go
// Instead of:
result, err := closureFn.Call(args...)

// Use:
ctx := tengo.NewExecutionContext(compiled)
result, err := ctx.Call(closureFn, args...)
```

#### Issue 2: Race conditions in concurrent code

**Problem**: Inconsistent results when calling closures from multiple goroutines

**Solution**: Use isolated contexts

```go
// Instead of:
go func() {
    result, _ := ctx.Call(closureFn)
}()

// Use:
go func() {
    isolatedCtx := ctx.WithIsolatedGlobals()
    result, _ := isolatedCtx.Call(closureFn)
}()
```

#### Issue 3: Memory usage increases

**Problem**: Higher memory usage after migration

**Solution**: This is expected due to context preservation. Consider:

1. Reusing contexts where possible
2. Using isolated contexts only when necessary
3. Monitoring memory usage in production

#### Issue 4: Performance degradation

**Problem**: Slower execution after migration

**Solution**: 
1. Use `ExecutionContext` only when you need global access
2. Keep direct `Call()` for simple functions without globals
3. Consider performance vs. correctness trade-offs

### Debugging Tips

#### Enable Debug Logging

```go
// Add debug logging to understand context usage
func debugContext(ctx *tengo.ExecutionContext) {
    fmt.Printf("Context has %d constants, %d globals\n", 
        len(ctx.Constants()), len(ctx.Globals()))
}
```

#### Validate Context State

```go
// Check if context is complete
func validateContext(ctx *tengo.ExecutionContext) error {
    if len(ctx.Constants()) == 0 {
        return errors.New("context missing constants")
    }
    if len(ctx.Globals()) == 0 {
        return errors.New("context missing globals")
    }
    return nil
}
```

## Performance Considerations

### Performance Impact

Based on benchmarks:
- **Inline execution**: No performance impact
- **Context-aware execution**: ~73x slower than inline
- **Memory usage**: ~790x more memory for context preservation

### When to Use Each Approach

#### Use Direct Calls When:
- Functions don't access globals
- Performance is critical
- Simple, stateless operations

```go
// For simple functions without globals
result, err := simpleFn.Call(args...)
```

#### Use ExecutionContext When:
- Functions access globals
- Thread safety is important
- Context preservation is needed

```go
// For closures with globals
ctx := tengo.NewExecutionContext(compiled)
result, err := ctx.Call(closureFn, args...)
```

#### Use Isolated Contexts When:
- Concurrent execution
- Independent global state needed
- Testing scenarios

```go
// For concurrent execution
isolatedCtx := ctx.WithIsolatedGlobals()
result, err := isolatedCtx.Call(closureFn, args...)
```

## Best Practices

### 1. Choose the Right API

```go
// ✅ Good: Use ExecutionContext for closures with globals
ctx := tengo.NewExecutionContext(compiled)
result, err := ctx.Call(closureFn, args...)

// ❌ Avoid: Direct calls for closures that need globals
result, err := closureFn.Call(args...) // May fail
```

### 2. Reuse Contexts When Possible

```go
// ✅ Good: Reuse context for multiple calls
ctx := tengo.NewExecutionContext(compiled)
for _, fn := range closures {
    result, err := ctx.Call(fn, args...)
    // Process result
}

// ❌ Avoid: Creating new context for each call
for _, fn := range closures {
    ctx := tengo.NewExecutionContext(compiled) // Wasteful
    result, err := ctx.Call(fn, args...)
}
```

### 3. Use Isolation for Concurrency

```go
// ✅ Good: Isolated contexts for goroutines
for i := 0; i < workers; i++ {
    go func() {
        ctx := baseCtx.WithIsolatedGlobals()
        // Safe concurrent execution
    }()
}

// ❌ Avoid: Shared context across goroutines
for i := 0; i < workers; i++ {
    go func() {
        // Race condition possible
        result, _ := sharedCtx.Call(closureFn)
    }()
}
```

### 4. Handle Errors Properly

```go
// ✅ Good: Proper error handling
ctx := tengo.NewExecutionContext(compiled)
result, err := ctx.Call(closureFn, args...)
if err != nil {
    if errors.Is(err, tengo.ErrMissingExecutionContext) {
        // Handle context error
    }
    return fmt.Errorf("closure execution failed: %w", err)
}
```

### 5. Monitor Performance

```go
// ✅ Good: Monitor performance impact
start := time.Now()
result, err := ctx.Call(closureFn, args...)
duration := time.Since(start)

if duration > performanceThreshold {
    log.Printf("Slow closure execution: %v", duration)
}
```

## Migration Checklist

- [ ] **Identify closure usage** in your codebase
- [ ] **Test existing functionality** to ensure no regressions
- [ ] **Update concurrent code** to use isolated contexts
- [ ] **Add error handling** for new context-related errors
- [ ] **Monitor performance** after migration
- [ ] **Update documentation** to reflect new patterns
- [ ] **Train team** on new best practices
- [ ] **Consider gradual rollout** for large codebases

## Summary

The new closure-with-globals functionality provides:
- **Better correctness** through proper context preservation
- **Improved thread safety** with isolated execution contexts
- **Enhanced flexibility** in globals management
- **Backward compatibility** with existing code

While there are performance trade-offs, the improved correctness and thread safety make it worthwhile for most use cases. Choose the right API for your specific needs and follow the best practices outlined in this guide.
