# Rollback Instructions for Plan B

## Current State Analysis

After our investigation, we have several working components:
1. ✅ `Constants()` method added to `Compiled` struct
2. ✅ `CallWithGlobalsExAndConstants()` method added to `CompiledFunction`
3. ❌ Complex VM setup that's causing nil pointer dereferences

## Recommended Rollback Actions

### 1. Simplify the Integration Test

Instead of trying to call closures without globals in isolation, let's focus on the primary use case: calling closures WITH the proper context.

```go
// Simplify integration_test.go - focus on the main use case
func TestClosureWithGlobals_Integration(t *testing.T) {
    // Test calling closures with proper context (constants + globals)
    // Test state isolation
    // Test error handling
    // Skip the complex isolated VM test for now
}
```

### 2. Keep the Constants Infrastructure

The `Constants()` method and `CallWithGlobalsExAndConstants()` method are valuable and should be kept.

### 3. Fix the VM Context Setup

The issue is in the VM setup - we need to ensure:
- Constants are properly passed
- Stack is properly initialized
- Free variables are properly set up

### 4. Start with Simpler Implementation

Instead of trying to create a fully isolated VM, let's start with a simpler approach:

```go
func (o *CompiledFunction) CallWithGlobalsExAndConstants(constants []Object, globals []Object, args ...Object) (Object, []Object, error) {
    // Validate arguments
    if o.VarArgs {
        if len(args) < o.NumParameters-1 {
            return nil, nil, fmt.Errorf("wrong number of arguments: want>=%d, got=%d", o.NumParameters-1, len(args))
        }
    } else {
        if len(args) != o.NumParameters {
            return nil, nil, fmt.Errorf("wrong number of arguments: want=%d, got=%d", o.NumParameters, len(args))
        }
    }

    // For now, return an error if constants are needed but not provided
    if constants == nil && len(o.Instructions) > 0 {
        // Check if the function actually needs constants by scanning instructions
        needsConstants := false
        for i := 0; i < len(o.Instructions); i++ {
            if o.Instructions[i] == parser.OpConstant {
                needsConstants = true
                break
            }
        }
        
        if needsConstants {
            return nil, nil, fmt.Errorf("function requires constants but none provided")
        }
    }

    // Continue with existing implementation but with better error handling
    // ...
}
```

## Implementation Strategy for Plan B

### Phase 1: Foundation (This Week)
1. Keep current `Constants()` and `CallWithGlobalsExAndConstants()` methods
2. Fix the VM context setup issues
3. Create simple integration tests that work
4. Focus on the primary use case: calling closures with proper context

### Phase 2: ExecutionContext Wrapper (Next Week)
1. Create the `ExecutionContext` struct
2. Add factory methods
3. Improve error handling
4. Add more comprehensive tests

### Phase 3: Advanced Features (Later)
1. Automatic context resolution
2. Performance optimizations  
3. Developer experience improvements

## Key Principles for Plan B

1. **Start Simple**: Begin with the most basic working version
2. **Incremental**: Add complexity only when needed
3. **Test-Driven**: Each change should be validated with tests
4. **Backward Compatible**: Don't break existing functionality
5. **Clear Errors**: Provide helpful error messages when things go wrong

## Next Steps

1. Simplify the integration test to focus on working use cases
2. Fix the VM context setup to handle constants properly
3. Add basic error handling for missing context
4. Validate that the core functionality works
5. Then proceed with the ExecutionContext wrapper

This approach is more sustainable and aligns with Tengo's existing architecture while still solving the core problem.
