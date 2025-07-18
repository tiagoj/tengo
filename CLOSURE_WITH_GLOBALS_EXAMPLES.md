# Closure with Globals Usage Examples

This document provides practical examples of using the Closure with Globals API in various scenarios.

## Table of Contents

1. [Basic Closure with Globals](#basic-closure-with-globals)
2. [Isolated Context Execution](#isolated-context-execution)
3. [Custom Globals Modification](#custom-globals-modification)
4. [Error Handling Best Practices](#error-handling-best-practices)
5. [Concurrent Execution](#concurrent-execution)
6. [Complex Data Types](#complex-data-types)
7. [Nested Closures](#nested-closures)
8. [Direct API Usage](#direct-api-usage)

## Basic Closure with Globals

### Example: Simple Counter with Global State

```go
package main

import (
    "fmt"
    "github.com/d5/tengo/v2"
)

func main() {
    // Define Tengo script with global variable and closure
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

    // Compile and run the script
    compiled, err := script.Compile()
    if err != nil {
        panic(err)
    }
    
    err = compiled.Run()
    if err != nil {
        panic(err)
    }

    // Create execution context
    ctx := tengo.NewExecutionContext(compiled)

    // Get closures from the script
    incrementVar := compiled.Get("increment")
    incrementFn := incrementVar.Value().(*tengo.CompiledFunction)
    
    getCounterVar := compiled.Get("get_counter")
    getCounterFn := getCounterVar.Value().(*tengo.CompiledFunction)

    // Call closures with access to globals
    result1, err := ctx.Call(incrementFn)
    if err != nil {
        panic(err)
    }
    fmt.Printf("First increment: %d\n", result1.(*tengo.Int).Value) // 1

    result2, err := ctx.Call(incrementFn)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Second increment: %d\n", result2.(*tengo.Int).Value) // 2

    counter, err := ctx.Call(getCounterFn)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Current counter: %d\n", counter.(*tengo.Int).Value) // 2
}
```

## Isolated Context Execution

### Example: Multiple Independent Counters

```go
package main

import (
    "fmt"
    "sync"
    "github.com/d5/tengo/v2"
)

func main() {
    // Script with shared counter logic
    script := tengo.NewScript([]byte(`
        counter := 0
        
        increment := func(amount) {
            counter += amount
            return counter
        }
    `))

    compiled, err := script.Compile()
    if err != nil {
        panic(err)
    }
    
    err = compiled.Run()
    if err != nil {
        panic(err)
    }

    // Get the increment function
    incrementVar := compiled.Get("increment")
    incrementFn := incrementVar.Value().(*tengo.CompiledFunction)

    // Create multiple isolated contexts
    ctx1 := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
    ctx2 := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()

    var wg sync.WaitGroup
    
    // Goroutine 1: Increment by 1, five times
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 5; i++ {
            result, err := ctx1.Call(incrementFn, &tengo.Int{Value: 1})
            if err != nil {
                panic(err)
            }
            fmt.Printf("Context 1: %d\n", result.(*tengo.Int).Value)
        }
    }()

    // Goroutine 2: Increment by 2, five times
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 5; i++ {
            result, err := ctx2.Call(incrementFn, &tengo.Int{Value: 2})
            if err != nil {
                panic(err)
            }
            fmt.Printf("Context 2: %d\n", result.(*tengo.Int).Value)
        }
    }()

    wg.Wait()
    // Context 1 will show: 1, 2, 3, 4, 5
    // Context 2 will show: 2, 4, 6, 8, 10
    // They remain isolated from each other
}
```

## Custom Globals Modification

### Example: Calculator with Different Precision Settings

```go
package main

import (
    "fmt"
    "github.com/d5/tengo/v2"
)

func main() {
    // Script with configurable precision
    script := tengo.NewScript([]byte(`
        precision := 2
        
        round_result := func(value) {
            multiplier := 1
            for i := 0; i < precision; i++ {
                multiplier *= 10
            }
            return int(value * multiplier + 0.5) / multiplier
        }
        
        calculate := func(a, b, operation) {
            result := 0.0
            if operation == "add" {
                result = a + b
            } else if operation == "multiply" {
                result = a * b
            } else if operation == "divide" {
                result = a / b
            }
            return round_result(result)
        }
    `))

    compiled, err := script.Compile()
    if err != nil {
        panic(err)
    }
    
    err = compiled.Run()
    if err != nil {
        panic(err)
    }

    // Get the calculate function
    calculateVar := compiled.Get("calculate")
    calculateFn := calculateVar.Value().(*tengo.CompiledFunction)

    // Create context with default precision (2)
    ctx := tengo.NewExecutionContext(compiled)
    
    // Create context with higher precision (4)
    globals := compiled.Globals()
    customGlobals := make([]tengo.Object, len(globals))
    copy(customGlobals, globals)
    customGlobals[0] = &tengo.Int{Value: 4} // precision = 4
    
    highPrecisionCtx := ctx.WithGlobals(customGlobals)

    // Test division with different precisions
    a := &tengo.Float{Value: 10.0}
    b := &tengo.Float{Value: 3.0}
    op := &tengo.String{Value: "divide"}

    // Default precision (2 decimal places)
    result1, err := ctx.Call(calculateFn, a, b, op)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Default precision: %v\n", result1) // 3.33

    // High precision (4 decimal places)
    result2, err := highPrecisionCtx.Call(calculateFn, a, b, op)
    if err != nil {
        panic(err)
    }
    fmt.Printf("High precision: %v\n", result2) // 3.3333
}
```

## Error Handling Best Practices

### Example: Robust Closure Execution with Error Handling

```go
package main

import (
    "fmt"
    "log"
    "github.com/d5/tengo/v2"
)

type ClosureExecutor struct {
    ctx      *tengo.ExecutionContext
    function *tengo.CompiledFunction
}

func NewClosureExecutor(script string, functionName string) (*ClosureExecutor, error) {
    // Compile script
    s := tengo.NewScript([]byte(script))
    compiled, err := s.Compile()
    if err != nil {
        return nil, fmt.Errorf("compile error: %w", err)
    }
    
    // Run script
    err = compiled.Run()
    if err != nil {
        return nil, fmt.Errorf("runtime error: %w", err)
    }

    // Get function
    funcVar := compiled.Get(functionName)
    if funcVar == nil {
        return nil, fmt.Errorf("function '%s' not found", functionName)
    }

    function, ok := funcVar.Value().(*tengo.CompiledFunction)
    if !ok {
        return nil, fmt.Errorf("'%s' is not a function", functionName)
    }

    // Create execution context
    ctx := tengo.NewExecutionContext(compiled)

    return &ClosureExecutor{
        ctx:      ctx,
        function: function,
    }, nil
}

func (ce *ClosureExecutor) Execute(args ...tengo.Object) (tengo.Object, error) {
    result, err := ce.ctx.Call(ce.function, args...)
    if err != nil {
        return nil, fmt.Errorf("execution error: %w", err)
    }
    return result, nil
}

func (ce *ClosureExecutor) ExecuteWithIsolation(args ...tengo.Object) (tengo.Object, error) {
    isolatedCtx := ce.ctx.WithIsolatedGlobals()
    result, err := isolatedCtx.Call(ce.function, args...)
    if err != nil {
        return nil, fmt.Errorf("isolated execution error: %w", err)
    }
    return result, nil
}

func main() {
    script := `
        config := {
            multiplier: 2,
            offset: 10
        }
        
        process := func(value) {
            if value < 0 {
                return error("negative values not allowed")
            }
            return value * config.multiplier + config.offset
        }
    `

    executor, err := NewClosureExecutor(script, "process")
    if err != nil {
        log.Fatal("Setup error:", err)
    }

    // Test with valid input
    result, err := executor.Execute(&tengo.Int{Value: 5})
    if err != nil {
        log.Printf("Execution error: %v", err)
    } else {
        fmt.Printf("Result: %v\n", result) // 20 (5 * 2 + 10)
    }

    // Test with invalid input (should trigger error in script)
    result, err = executor.Execute(&tengo.Int{Value: -1})
    if err != nil {
        log.Printf("Expected error: %v", err) // Error from script
    }

    // Test isolated execution
    result, err = executor.ExecuteWithIsolation(&tengo.Int{Value: 3})
    if err != nil {
        log.Printf("Isolation error: %v", err)
    } else {
        fmt.Printf("Isolated result: %v\n", result) // 16 (3 * 2 + 10)
    }
}
```

## Concurrent Execution

### Example: Thread-Safe Task Processing

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/d5/tengo/v2"
)

func main() {
    // Script that processes tasks with global configuration
    script := tengo.NewScript([]byte(`
        config := {
            timeout: 1000,
            retries: 3,
            batch_size: 10
        }
        
        task_count := 0
        
        process_task := func(task_id, data) {
            task_count += 1
            
            // Simulate processing
            processed_data := data * 2
            
            return {
                id: task_id,
                processed_data: processed_data,
                config: config,
                total_processed: task_count
            }
        }
    `))

    compiled, err := script.Compile()
    if err != nil {
        panic(err)
    }
    
    err = compiled.Run()
    if err != nil {
        panic(err)
    }

    // Get the process function
    processVar := compiled.Get("process_task")
    processFn := processVar.Value().(*tengo.CompiledFunction)

    // Create multiple workers with isolated contexts
    const numWorkers = 3
    const tasksPerWorker = 5

    var wg sync.WaitGroup
    results := make(chan string, numWorkers*tasksPerWorker)

    for workerID := 0; workerID < numWorkers; workerID++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // Each worker gets its own isolated context
            ctx := tengo.NewExecutionContext(compiled).WithIsolatedGlobals()
            
            for taskNum := 0; taskNum < tasksPerWorker; taskNum++ {
                taskID := id*tasksPerWorker + taskNum
                
                result, err := ctx.Call(processFn, 
                    &tengo.Int{Value: int64(taskID)}, 
                    &tengo.Int{Value: int64(taskID * 10)})
                
                if err != nil {
                    results <- fmt.Sprintf("Worker %d, Task %d: ERROR %v", 
                        id, taskID, err)
                } else {
                    resultMap := result.(*tengo.Map)
                    processedData := resultMap.Value["processed_data"].(*tengo.Int).Value
                    totalProcessed := resultMap.Value["total_processed"].(*tengo.Int).Value
                    
                    results <- fmt.Sprintf("Worker %d, Task %d: processed_data=%d, total_in_worker=%d", 
                        id, taskID, processedData, totalProcessed)
                }
                
                time.Sleep(10 * time.Millisecond) // Simulate work
            }
        }(workerID)
    }

    // Close results channel when all workers are done
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect and display results
    for result := range results {
        fmt.Println(result)
    }
    
    // Each worker maintains its own task_count due to isolation
}
```

## Complex Data Types

### Example: Working with Maps and Arrays

```go
package main

import (
    "fmt"
    "github.com/d5/tengo/v2"
)

func main() {
    script := tengo.NewScript([]byte(`
        users := []
        user_count := 0
        
        add_user := func(user_data) {
            user_data.id = user_count
            user_count += 1
            
            users = append(users, user_data)
            
            return {
                success: true,
                user_id: user_data.id,
                total_users: len(users)
            }
        }
        
        get_user := func(user_id) {
            for user in users {
                if user.id == user_id {
                    return user
                }
            }
            return error("user not found")
        }
        
        get_all_users := func() {
            return {
                users: users,
                count: len(users)
            }
        }
    `))

    compiled, err := script.Compile()
    if err != nil {
        panic(err)
    }
    
    err = compiled.Run()
    if err != nil {
        panic(err)
    }

    ctx := tengo.NewExecutionContext(compiled)

    // Get functions
    addUserVar := compiled.Get("add_user")
    addUserFn := addUserVar.Value().(*tengo.CompiledFunction)
    
    getUserVar := compiled.Get("get_user")
    getUserFn := getUserVar.Value().(*tengo.CompiledFunction)
    
    getAllUsersVar := compiled.Get("get_all_users")
    getAllUsersFn := getAllUsersVar.Value().(*tengo.CompiledFunction)

    // Create user data
    userData1 := &tengo.Map{
        Value: map[string]tengo.Object{
            "name":  &tengo.String{Value: "Alice"},
            "email": &tengo.String{Value: "alice@example.com"},
            "age":   &tengo.Int{Value: 30},
        },
    }

    userData2 := &tengo.Map{
        Value: map[string]tengo.Object{
            "name":  &tengo.String{Value: "Bob"},
            "email": &tengo.String{Value: "bob@example.com"},
            "age":   &tengo.Int{Value: 25},
        },
    }

    // Add users
    result1, err := ctx.Call(addUserFn, userData1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Added user 1: %v\n", result1)

    result2, err := ctx.Call(addUserFn, userData2)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Added user 2: %v\n", result2)

    // Get specific user
    user, err := ctx.Call(getUserFn, &tengo.Int{Value: 0})
    if err != nil {
        panic(err)
    }
    fmt.Printf("Retrieved user: %v\n", user)

    // Get all users
    allUsers, err := ctx.Call(getAllUsersFn)
    if err != nil {
        panic(err)
    }
    fmt.Printf("All users: %v\n", allUsers)
}
```

## Nested Closures

### Example: Factory Functions Creating Specialized Processors

```go
package main

import (
    "fmt"
    "github.com/d5/tengo/v2"
)

func main() {
    script := tengo.NewScript([]byte(`
        global_config := {
            version: "1.0",
            debug: false
        }
        
        create_processor := func(processor_type, config) {
            // This returns a closure that captures both parameters and globals
            return func(data) {
                result := {
                    type: processor_type,
                    config: config,
                    global_config: global_config,
                    processed_data: []
                }
                
                if processor_type == "transformer" {
                    for item in data {
                        result.processed_data = append(result.processed_data, item * config.multiplier)
                    }
                } else if processor_type == "filter" {
                    for item in data {
                        if item > config.threshold {
                            result.processed_data = append(result.processed_data, item)
                        }
                    }
                } else if processor_type == "aggregator" {
                    sum := 0
                    for item in data {
                        sum += item
                    }
                    result.processed_data = [sum / len(data)] // average
                }
                
                return result
            }
        }
    `))

    compiled, err := script.Compile()
    if err != nil {
        panic(err)
    }
    
    err = compiled.Run()
    if err != nil {
        panic(err)
    }

    ctx := tengo.NewExecutionContext(compiled)

    // Get the factory function
    createProcessorVar := compiled.Get("create_processor")
    createProcessorFn := createProcessorVar.Value().(*tengo.CompiledFunction)

    // Create different types of processors
    
    // 1. Transformer processor
    transformerConfig := &tengo.Map{
        Value: map[string]tengo.Object{
            "multiplier": &tengo.Int{Value: 2},
        },
    }
    
    transformer, err := ctx.Call(createProcessorFn, 
        &tengo.String{Value: "transformer"}, transformerConfig)
    if err != nil {
        panic(err)
    }
    transformerFn := transformer.(*tengo.CompiledFunction)

    // 2. Filter processor
    filterConfig := &tengo.Map{
        Value: map[string]tengo.Object{
            "threshold": &tengo.Int{Value: 5},
        },
    }
    
    filter, err := ctx.Call(createProcessorFn, 
        &tengo.String{Value: "filter"}, filterConfig)
    if err != nil {
        panic(err)
    }
    filterFn := filter.(*tengo.CompiledFunction)

    // 3. Aggregator processor
    aggregatorConfig := &tengo.Map{
        Value: map[string]tengo.Object{},
    }
    
    aggregator, err := ctx.Call(createProcessorFn, 
        &tengo.String{Value: "aggregator"}, aggregatorConfig)
    if err != nil {
        panic(err)
    }
    aggregatorFn := aggregator.(*tengo.CompiledFunction)

    // Test data
    testData := &tengo.Array{
        Value: []tengo.Object{
            &tengo.Int{Value: 1},
            &tengo.Int{Value: 3},
            &tengo.Int{Value: 7},
            &tengo.Int{Value: 9},
            &tengo.Int{Value: 2},
        },
    }

    // Process with transformer
    transformResult, err := ctx.Call(transformerFn, testData)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Transformer result: %v\n", transformResult)

    // Process with filter
    filterResult, err := ctx.Call(filterFn, testData)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Filter result: %v\n", filterResult)

    // Process with aggregator
    aggResult, err := ctx.Call(aggregatorFn, testData)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Aggregator result: %v\n", aggResult)
}
```

## Direct API Usage

### Example: Using CallWithGlobalsExAndConstants Directly

```go
package main

import (
    "fmt"
    "github.com/d5/tengo/v2"
)

func main() {
    script := tengo.NewScript([]byte(`
        multiplier := 5
        
        calculate := func(base, increment) {
            return (base + increment) * multiplier
        }
    `))

    compiled, err := script.Compile()
    if err != nil {
        panic(err)
    }
    
    err = compiled.Run()
    if err != nil {
        panic(err)
    }

    // Get the function
    calculateVar := compiled.Get("calculate")
    calculateFn := calculateVar.Value().(*tengo.CompiledFunction)

    // Get constants and globals
    constants := compiled.Constants()
    globals := compiled.Globals()

    fmt.Println("=== Direct API Usage ===")

    // Method 1: Use original globals
    result1, updatedGlobals1, err := calculateFn.CallWithGlobalsExAndConstants(
        constants, globals, 
        &tengo.Int{Value: 10}, &tengo.Int{Value: 5})
    if err != nil {
        panic(err)
    }
    fmt.Printf("With original globals (multiplier=5): %d\n", 
        result1.(*tengo.Int).Value) // 75 ((10+5)*5)

    // Method 2: Use modified globals
    customGlobals := make([]tengo.Object, len(globals))
    copy(customGlobals, globals)
    customGlobals[0] = &tengo.Int{Value: 10} // Change multiplier to 10

    result2, updatedGlobals2, err := calculateFn.CallWithGlobalsExAndConstants(
        constants, customGlobals,
        &tengo.Int{Value: 10}, &tengo.Int{Value: 5})
    if err != nil {
        panic(err)
    }
    fmt.Printf("With custom globals (multiplier=10): %d\n", 
        result2.(*tengo.Int).Value) // 150 ((10+5)*10)

    // Method 3: Check if globals were updated
    fmt.Printf("Original globals modified: %v\n", len(updatedGlobals1) > 0)
    fmt.Printf("Custom globals modified: %v\n", len(updatedGlobals2) > 0)

    // Compare with ExecutionContext approach
    fmt.Println("\n=== ExecutionContext Comparison ===")
    
    ctx := tengo.NewExecutionContext(compiled)
    result3, err := ctx.Call(calculateFn, 
        &tengo.Int{Value: 10}, &tengo.Int{Value: 5})
    if err != nil {
        panic(err)
    }
    fmt.Printf("ExecutionContext result: %d\n", 
        result3.(*tengo.Int).Value) // Should match result1
}
```

## Summary

These examples demonstrate the versatility and power of the Closure with Globals API:

- **Basic Usage**: Simple closure execution with global state preservation
- **Isolation**: Thread-safe execution with independent contexts  
- **Customization**: Modifying globals for different execution environments
- **Error Handling**: Robust error handling patterns
- **Concurrency**: Safe concurrent execution patterns
- **Complex Types**: Working with maps, arrays, and nested structures
- **Advanced Patterns**: Nested closures and factory functions
- **Direct API**: Low-level API usage for maximum control

The API provides both high-level convenience methods (`ExecutionContext`) and low-level control (`CallWithGlobalsExAndConstants`) to suit different use cases and performance requirements.
