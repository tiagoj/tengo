# Tengo Enhancement Plan B: Context-Aware Closure Execution

## Overview

This document outlines **Plan B** for enhancing Tengo's closure-with-globals functionality. After investigating the original approach, we discovered that the core issue is not just about passing globals, but about ensuring that `CompiledFunction` objects have access to their complete execution context (constants, globals, and free variables) when called via the Go API.

## Problem Statement

### Original Issue
When closures are extracted from a compiled Tengo script and called via Go API methods like `CallWithGlobals()`, they lose access to:
1. **Constants** from the original compilation (literals like numbers, strings)
2. **Globals** from the script's execution context
3. **Free variables** captured during closure creation

### Root Cause Discovery
During investigation, we found that `CompiledFunction` objects are designed to execute within a VM context that provides:
- **Constants array** from the original bytecode compilation
- **Globals array** representing the script's global state
- **Free variables** for closure scope

When these functions are called in isolation, they fail because they can't access the constants they reference.

## Requirements (Unchanged)

### Functional Requirements
1. **Global Access**: Closures should retain access to global variables when called via Go API
2. **Context Preservation**: The execution context should be maintained across calls
3. **State Isolation**: Each call should have its own copy of globals to prevent race conditions
4. **Backward Compatibility**: Existing code should continue to work without modification
5. **Performance**: Minimal overhead for the new functionality

### Non-Functional Requirements
1. **Thread Safety**: Multiple goroutines should be able to call closures concurrently
2. **Memory Efficiency**: Avoid unnecessary duplication of data
3. **API Consistency**: New methods should follow existing Tengo patterns
4. **Error Handling**: Clear error messages for context-related failures

## Plan B Architecture

### Core Concept
Instead of creating fully isolated VM execution contexts, we'll ensure that `CompiledFunction` objects have access to their original compilation context when needed.

### Key Components

1. **Context-Aware Execution**: Extend `CompiledFunction` with methods that accept execution context
2. **Constants Management**: Provide access to constants from the original compilation
3. **Execution Context Wrapper**: Create a lightweight wrapper that bundles constants, globals, and function together
4. **Backward Compatibility Layer**: Ensure existing APIs continue working

## Implementation Plan

### Phase 1: Core Context Infrastructure (Week 1)

#### 1.1 Add Constants Access Method
- Add `Constants()` method to `Compiled` struct to expose constants array
- Ensure thread-safe access to constants

#### 1.2 Extend CompiledFunction API
- Add `CallWithGlobalsExAndConstants()` method that accepts constants parameter
- Modify existing `CallWithGlobalsEx()` to use the new method internally
- Maintain backward compatibility

#### 1.3 Fix VM Context Setup
- Ensure VM is created with proper constants array
- Fix isolated VM execution to handle constants correctly
- Address nil pointer dereference issues in VM.run()

### Phase 2: Enhanced Context Management (Week 2)

#### 2.1 Create ExecutionContext Wrapper
```go
type ExecutionContext struct {
    constants []Object
    globals   []Object
    compiled  *Compiled
}

func (ec *ExecutionContext) Call(fn *CompiledFunction, args ...Object) (Object, error)
func (ec *ExecutionContext) CallEx(fn *CompiledFunction, args ...Object) (Object, []Object, error)
```

#### 2.2 Add Context Factory Methods
- Add `NewExecutionContext()` to create context from `Compiled` objects
- Add `WithGlobals()` method to create context with specific globals
- Add `WithIsolatedGlobals()` for thread-safe execution

#### 2.3 Improve Error Handling
- Add specific error types for context-related failures
- Provide clear error messages for missing constants/globals
- Add validation for execution context completeness

### Phase 3: Advanced Features (Week 3)

#### 3.1 Automatic Context Resolution
- Implement automatic context detection where possible
- Add heuristics to determine if constants are needed
- Provide warnings for incomplete context

#### 3.2 Performance Optimizations
- Implement lazy context creation
- Add context caching for frequently called functions
- Optimize memory usage for context storage

#### 3.3 Developer Experience Improvements
- Add helper methods for common use cases
- Provide examples and documentation
- Add debugging utilities for context inspection

### Phase 4: Testing and Validation (Week 4)

#### 4.1 Comprehensive Testing
- Unit tests for all new methods
- Integration tests for end-to-end scenarios
- Performance benchmarks
- Concurrency tests

#### 4.2 Documentation
- Update API documentation
- Create usage examples
- Write migration guide for existing code
- Add troubleshooting guide

## Detailed Implementation Steps

### Step 1: Constants Access Infrastructure

```go
// In script.go
func (c *Compiled) Constants() []Object {
    c.lock.RLock()
    defer c.lock.RUnlock()
    return c.bytecode.Constants
}
```

### Step 2: Enhanced CompiledFunction Methods

```go
// In objects.go
func (o *CompiledFunction) CallWithGlobalsExAndConstants(
    constants []Object, 
    globals []Object, 
    args ...Object,
) (Object, []Object, error) {
    // Implementation with proper context setup
}
```

### Step 3: ExecutionContext Wrapper

```go
// New file: execution_context.go
type ExecutionContext struct {
    constants []Object
    globals   []Object
    source    *Compiled
}

func NewExecutionContext(compiled *Compiled) *ExecutionContext {
    return &ExecutionContext{
        constants: compiled.Constants(),
        globals:   compiled.Globals(),
        source:    compiled,
    }
}
```

### Step 4: Integration Tests

```go
// Enhanced integration test
func TestClosureWithGlobals_ContextAware(t *testing.T) {
    // Test with automatic context resolution
    // Test with explicit context
    // Test with isolated context
    // Test error cases
}
```

## Risk Mitigation

### Technical Risks
1. **Performance Impact**: Monitor execution overhead and optimize if needed
2. **Memory Usage**: Ensure context objects don't cause memory leaks
3. **Thread Safety**: Thorough testing of concurrent access patterns
4. **Backward Compatibility**: Extensive testing of existing code

### Mitigation Strategies
1. **Incremental Implementation**: Roll out changes in phases
2. **Comprehensive Testing**: Unit, integration, and performance tests
3. **Documentation**: Clear migration paths and examples
4. **Community Feedback**: Early feedback from users

## Success Metrics

### Functional Metrics
- ✅ All closure-with-globals use cases work correctly
- ✅ No regressions in existing functionality
- ✅ Performance impact < 5% for typical use cases
- ✅ Memory usage increase < 10% for context storage

### Quality Metrics
- ✅ 100% test coverage for new functionality
- ✅ Zero critical bugs in context handling
- ✅ Clear error messages for all failure cases
- ✅ Complete documentation and examples

## Timeline

### Week 1: Core Infrastructure
- [ ] Add Constants() method to Compiled
- [ ] Implement CallWithGlobalsExAndConstants()
- [ ] Fix VM context setup issues
- [ ] Basic integration tests

### Week 2: Enhanced Context Management
- [ ] Create ExecutionContext wrapper
- [ ] Add context factory methods
- [ ] Improve error handling
- [ ] Advanced integration tests

### Week 3: Advanced Features
- [ ] Automatic context resolution
- [ ] Performance optimizations
- [ ] Developer experience improvements
- [ ] Comprehensive testing

### Week 4: Polish and Documentation
- [ ] Performance benchmarks
- [ ] Complete documentation
- [ ] Migration guide
- [ ] Community feedback integration

## Conclusion

Plan B takes a more systematic approach to the closure-with-globals problem by addressing the underlying execution context issues. Instead of trying to create fully isolated execution environments, we're extending the existing architecture to properly handle context propagation while maintaining backward compatibility.

This approach is more aligned with Tengo's existing design patterns and should be more maintainable and performant in the long run. The phased implementation allows for continuous validation and course correction as we learn more about the system's behavior.

## Key Differences from Original Plan

1. **Focus on Context**: Emphasis on proper execution context rather than isolated VM execution
2. **Incremental Approach**: Smaller, more focused changes rather than large architectural shifts
3. **Backward Compatibility**: Stronger emphasis on maintaining existing functionality
4. **Performance Awareness**: More attention to performance implications of context management
5. **Error Handling**: Better error handling and debugging capabilities

This plan builds on the lessons learned from the initial investigation while providing a more pragmatic path forward for the enhancement.
