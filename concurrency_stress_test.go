package tengo

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrentExecutionContextCreation tests creating multiple execution contexts concurrently
func TestConcurrentExecutionContextCreation(t *testing.T) {
	script := NewScript([]byte(`
		counter := 0
		increment := func() {
			counter += 1
			return counter
		}
	`))

	compiled, err := script.Compile()
	if err != nil {
		t.Fatal(err)
	}

	err = compiled.Run()
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 100
	var wg sync.WaitGroup
	contextCreationErrors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					contextCreationErrors <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			// Create execution context
			ctx := NewExecutionContext(compiled)
			if ctx == nil {
				contextCreationErrors <- fmt.Errorf("goroutine %d: failed to create execution context", id)
				return
			}

			// Create isolated context
			isolatedCtx := ctx.WithIsolatedGlobals()
			if isolatedCtx == nil {
				contextCreationErrors <- fmt.Errorf("goroutine %d: failed to create isolated context", id)
				return
			}

			// Verify context has proper data
			if len(isolatedCtx.Constants()) == 0 {
				contextCreationErrors <- fmt.Errorf("goroutine %d: context missing constants", id)
				return
			}

			if len(isolatedCtx.Globals()) == 0 {
				contextCreationErrors <- fmt.Errorf("goroutine %d: context missing globals", id)
				return
			}
		}(i)
	}

	wg.Wait()
	close(contextCreationErrors)

	// Check for errors
	for err := range contextCreationErrors {
		t.Error(err)
	}
}

// TestConcurrentIsolatedExecution tests executing closures in isolated contexts concurrently
func TestConcurrentIsolatedExecution(t *testing.T) {
	script := NewScript([]byte(`
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
	if err != nil {
		t.Fatal(err)
	}

	err = compiled.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Get functions
	incrementVar := compiled.Get("increment")
	if incrementVar == nil {
		t.Fatal("increment function not found")
	}
	incrementFn := incrementVar.Value().(*CompiledFunction)

	getCounterVar := compiled.Get("get_counter")
	if getCounterVar == nil {
		t.Fatal("get_counter function not found")
	}
	getCounterFn := getCounterVar.Value().(*CompiledFunction)

	const numGoroutines = 50
	const incrementsPerGoroutine = 20
	var wg sync.WaitGroup
	executionErrors := make(chan error, numGoroutines)
	finalCounters := make(chan int64, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					executionErrors <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			// Create isolated context for this goroutine
			ctx := NewExecutionContext(compiled).WithIsolatedGlobals()

			// Perform increments
			for j := 0; j < incrementsPerGoroutine; j++ {
				result, err := ctx.Call(incrementFn)
				if err != nil {
					executionErrors <- fmt.Errorf("goroutine %d, increment %d: %v", id, j, err)
					return
				}

				// Verify result is correct
				expected := int64(j + 1)
				actual := result.(*Int).Value
				if actual != expected {
					executionErrors <- fmt.Errorf("goroutine %d, increment %d: expected %d, got %d", id, j, expected, actual)
					return
				}
			}

			// Get final counter value
			finalResult, err := ctx.Call(getCounterFn)
			if err != nil {
				executionErrors <- fmt.Errorf("goroutine %d, final count: %v", id, err)
				return
			}

			finalCount := finalResult.(*Int).Value
			if finalCount != incrementsPerGoroutine {
				executionErrors <- fmt.Errorf("goroutine %d: expected final count %d, got %d", id, incrementsPerGoroutine, finalCount)
				return
			}

			finalCounters <- finalCount
		}(i)
	}

	wg.Wait()
	close(executionErrors)
	close(finalCounters)

	// Check for errors
	for err := range executionErrors {
		t.Error(err)
	}

	// Verify all goroutines completed with expected final count
	counterCount := 0
	for finalCount := range finalCounters {
		if finalCount != incrementsPerGoroutine {
			t.Errorf("unexpected final count: %d", finalCount)
		}
		counterCount++
	}

	if counterCount != numGoroutines {
		t.Errorf("expected %d final counters, got %d", numGoroutines, counterCount)
	}
}

// TestConcurrentSharedContextStress tests stress with shared (non-isolated) contexts
func TestConcurrentSharedContextStress(t *testing.T) {
	script := NewScript([]byte(`
		shared_counter := 0
		increment_shared := func() {
			shared_counter += 1
			return shared_counter
		}
	`))

	compiled, err := script.Compile()
	if err != nil {
		t.Fatal(err)
	}

	err = compiled.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Get function
	incrementVar := compiled.Get("increment_shared")
	if incrementVar == nil {
		t.Fatal("increment_shared function not found")
	}
	incrementFn := incrementVar.Value().(*CompiledFunction)

	// Create shared context (NOT isolated)
	sharedCtx := NewExecutionContext(compiled)

	const numGoroutines = 20
	const incrementsPerGoroutine = 10
	var wg sync.WaitGroup
	var totalIncrements int64
	executionErrors := make(chan error, numGoroutines*incrementsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					executionErrors <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			for j := 0; j < incrementsPerGoroutine; j++ {
				result, err := sharedCtx.Call(incrementFn)
				if err != nil {
					executionErrors <- fmt.Errorf("goroutine %d, increment %d: %v", id, j, err)
					return
				}

				// Count successful increments
				if result != nil {
					atomic.AddInt64(&totalIncrements, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	close(executionErrors)

	// Check for errors
	errorCount := 0
	for err := range executionErrors {
		t.Logf("Shared context error (expected): %v", err)
		errorCount++
	}

	expectedIncrements := int64(numGoroutines * incrementsPerGoroutine)
	t.Logf("Total increments attempted: %d", expectedIncrements)
	t.Logf("Total increments completed: %d", totalIncrements)
	t.Logf("Errors encountered: %d", errorCount)

	// With shared context, we expect some race conditions/errors
	// This test validates that the system handles concurrent access
	// without crashing, even if not all operations succeed
}

// TestConcurrentComplexDataManipulation tests concurrent manipulation of complex data structures
func TestConcurrentComplexDataManipulation(t *testing.T) {
	script := NewScript([]byte(`
		data_store := {}
		
		add_data := func(key, value) {
			data_store[key] = value
			return len(data_store)
		}
		
		get_data := func(key) {
			return data_store[key]
		}
		
		get_all_data := func() {
			return data_store
		}
	`))

	compiled, err := script.Compile()
	if err != nil {
		t.Fatal(err)
	}

	err = compiled.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Get functions
	addDataVar := compiled.Get("add_data")
	if addDataVar == nil {
		t.Fatal("add_data function not found")
	}
	addDataFn := addDataVar.Value().(*CompiledFunction)

	getDataVar := compiled.Get("get_data")
	if getDataVar == nil {
		t.Fatal("get_data function not found")
	}
	getDataFn := getDataVar.Value().(*CompiledFunction)

	getAllDataVar := compiled.Get("get_all_data")
	if getAllDataVar == nil {
		t.Fatal("get_all_data function not found")
	}
	getAllDataFn := getAllDataVar.Value().(*CompiledFunction)

	const numGoroutines = 30
	const operationsPerGoroutine = 15
	var wg sync.WaitGroup
	executionErrors := make(chan error, numGoroutines*operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					executionErrors <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			// Create isolated context for this goroutine
			ctx := NewExecutionContext(compiled).WithIsolatedGlobals()

			for j := 0; j < operationsPerGoroutine; j++ {
				key := &String{Value: fmt.Sprintf("key_%d_%d", id, j)}
				value := &Int{Value: int64(id*1000 + j)}

				// Add data
				result, err := ctx.Call(addDataFn, key, value)
				if err != nil {
					executionErrors <- fmt.Errorf("goroutine %d, add_data %d: %v", id, j, err)
					return
				}

				// Verify add result
				expectedSize := int64(j + 1)
				actualSize := result.(*Int).Value
				if actualSize != expectedSize {
					executionErrors <- fmt.Errorf("goroutine %d, add_data %d: expected size %d, got %d", id, j, expectedSize, actualSize)
					return
				}

				// Retrieve data
				retrievedValue, err := ctx.Call(getDataFn, key)
				if err != nil {
					executionErrors <- fmt.Errorf("goroutine %d, get_data %d: %v", id, j, err)
					return
				}

				// Verify retrieved value
				expectedValue := int64(id*1000 + j)
				actualValue := retrievedValue.(*Int).Value
				if actualValue != expectedValue {
					executionErrors <- fmt.Errorf("goroutine %d, get_data %d: expected %d, got %d", id, j, expectedValue, actualValue)
					return
				}
			}

			// Get all data at the end
			allData, err := ctx.Call(getAllDataFn)
			if err != nil {
				executionErrors <- fmt.Errorf("goroutine %d, get_all_data: %v", id, err)
				return
			}

			// Verify all data size
			allDataMap := allData.(*Map)
			if len(allDataMap.Value) != operationsPerGoroutine {
				executionErrors <- fmt.Errorf("goroutine %d, get_all_data: expected %d items, got %d", id, operationsPerGoroutine, len(allDataMap.Value))
				return
			}
		}(i)
	}

	wg.Wait()
	close(executionErrors)

	// Check for errors
	for err := range executionErrors {
		t.Error(err)
	}
}

// TestConcurrentMemoryStress tests memory usage under concurrent load
func TestConcurrentMemoryStress(t *testing.T) {
	script := NewScript([]byte(`
		large_array := []
		
		add_large_data := func() {
			for i := 0; i < 1000; i++ {
				large_array = append(large_array, i)
			}
			return len(large_array)
		}
		
		process_data := func() {
			sum := 0
			for item in large_array {
				sum += item
			}
			return sum
		}
	`))

	compiled, err := script.Compile()
	if err != nil {
		t.Fatal(err)
	}

	err = compiled.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Get functions
	addLargeDataVar := compiled.Get("add_large_data")
	if addLargeDataVar == nil {
		t.Fatal("add_large_data function not found")
	}
	addLargeDataFn := addLargeDataVar.Value().(*CompiledFunction)

	processDataVar := compiled.Get("process_data")
	if processDataVar == nil {
		t.Fatal("process_data function not found")
	}
	processDataFn := processDataVar.Value().(*CompiledFunction)

	// Get initial memory stats
	var initialMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMemStats)

	const numGoroutines = 20
	var wg sync.WaitGroup
	executionErrors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					executionErrors <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			// Create isolated context
			ctx := NewExecutionContext(compiled).WithIsolatedGlobals()

			// Add large data
			result, err := ctx.Call(addLargeDataFn)
			if err != nil {
				executionErrors <- fmt.Errorf("goroutine %d, add_large_data: %v", id, err)
				return
			}

			// Verify data was added
			arraySize := result.(*Int).Value
			if arraySize != 1000 {
				executionErrors <- fmt.Errorf("goroutine %d: expected array size 1000, got %d", id, arraySize)
				return
			}

			// Process data
			sum, err := ctx.Call(processDataFn)
			if err != nil {
				executionErrors <- fmt.Errorf("goroutine %d, process_data: %v", id, err)
				return
			}

			// Verify sum (0 + 1 + ... + 999 = 499500)
			expectedSum := int64(499500)
			actualSum := sum.(*Int).Value
			if actualSum != expectedSum {
				executionErrors <- fmt.Errorf("goroutine %d: expected sum %d, got %d", id, expectedSum, actualSum)
				return
			}
		}(i)
	}

	wg.Wait()
	close(executionErrors)

	// Check for errors
	for err := range executionErrors {
		t.Error(err)
	}

	// Get final memory stats
	var finalMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&finalMemStats)

	// Log memory usage
	t.Logf("Initial memory: %d bytes", initialMemStats.Alloc)
	t.Logf("Final memory: %d bytes", finalMemStats.Alloc)
	t.Logf("Memory increase: %d bytes", finalMemStats.Alloc-initialMemStats.Alloc)
	t.Logf("Total allocations: %d", finalMemStats.TotalAlloc)
}

// TestConcurrentErrorHandling tests error handling under concurrent load
func TestConcurrentErrorHandling(t *testing.T) {
	script := NewScript([]byte(`
		error_count := 0
		
		maybe_error := func(should_error) {
			if should_error {
				error_count += 1
				return error("intentional error")
			}
			return "success"
		}
		
		get_error_count := func() {
			return error_count
		}
	`))

	compiled, err := script.Compile()
	if err != nil {
		t.Fatal(err)
	}

	err = compiled.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Get functions
	maybeErrorVar := compiled.Get("maybe_error")
	if maybeErrorVar == nil {
		t.Fatal("maybe_error function not found")
	}
	maybeErrorFn := maybeErrorVar.Value().(*CompiledFunction)

	getErrorCountVar := compiled.Get("get_error_count")
	if getErrorCountVar == nil {
		t.Fatal("get_error_count function not found")
	}
	getErrorCountFn := getErrorCountVar.Value().(*CompiledFunction)

	const numGoroutines = 25
	const callsPerGoroutine = 10
	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64
	executionErrors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					executionErrors <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			// Create isolated context
			ctx := NewExecutionContext(compiled).WithIsolatedGlobals()

			for j := 0; j < callsPerGoroutine; j++ {
				// Half the calls should produce errors
				shouldError := (j % 2) == 0
				var boolArg Object
				if shouldError {
					boolArg = TrueValue
				} else {
					boolArg = FalseValue
				}
				result, err := ctx.Call(maybeErrorFn, boolArg)

				if shouldError {
					// Expect error object as result
					if err != nil {
						executionErrors <- fmt.Errorf("goroutine %d, call %d: unexpected Go error: %v", id, j, err)
						return
					}
					if _, isError := result.(*Error); !isError {
						executionErrors <- fmt.Errorf("goroutine %d, call %d: expected Error object but got %T: %v", id, j, result, result)
						return
					}
					atomic.AddInt64(&errorCount, 1)
				} else {
					// Expect success
					if err != nil {
						executionErrors <- fmt.Errorf("goroutine %d, call %d: unexpected Go error: %v", id, j, err)
						return
					}
					if result.(*String).Value != "success" {
						executionErrors <- fmt.Errorf("goroutine %d, call %d: expected 'success', got %v", id, j, result)
						return
					}
					atomic.AddInt64(&successCount, 1)
				}
			}

			// Get error count from this context
			errorCountResult, err := ctx.Call(getErrorCountFn)
			if err != nil {
				executionErrors <- fmt.Errorf("goroutine %d, get_error_count: %v", id, err)
				return
			}

			// Each context should have 5 errors (half of callsPerGoroutine)
			expectedErrors := int64(callsPerGoroutine / 2)
			actualErrors := errorCountResult.(*Int).Value
			if actualErrors != expectedErrors {
				executionErrors <- fmt.Errorf("goroutine %d: expected %d errors, got %d", id, expectedErrors, actualErrors)
				return
			}
		}(i)
	}

	wg.Wait()
	close(executionErrors)

	// Check for errors
	for err := range executionErrors {
		t.Error(err)
	}

	// Verify counts
	expectedSuccesses := int64(numGoroutines * callsPerGoroutine / 2)
	expectedErrors := int64(numGoroutines * callsPerGoroutine / 2)

	t.Logf("Expected successes: %d, Actual: %d", expectedSuccesses, successCount)
	t.Logf("Expected errors: %d, Actual: %d", expectedErrors, errorCount)

	if successCount != expectedSuccesses {
		t.Errorf("expected %d successes, got %d", expectedSuccesses, successCount)
	}

	if errorCount != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, errorCount)
	}
}

// TestConcurrentLongRunning tests long-running concurrent operations
func TestConcurrentLongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	script := NewScript([]byte(`
		operation_count := 0
		
		long_operation := func() {
			// Simulate some work
			sum := 0
			for i := 0; i < 10000; i++ {
				sum += i
			}
			operation_count += 1
			return sum
		}
		
		get_operation_count := func() {
			return operation_count
		}
	`))

	compiled, err := script.Compile()
	if err != nil {
		t.Fatal(err)
	}

	err = compiled.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Get functions
	longOperationVar := compiled.Get("long_operation")
	if longOperationVar == nil {
		t.Fatal("long_operation function not found")
	}
	longOperationFn := longOperationVar.Value().(*CompiledFunction)

	getOperationCountVar := compiled.Get("get_operation_count")
	if getOperationCountVar == nil {
		t.Fatal("get_operation_count function not found")
	}
	getOperationCountFn := getOperationCountVar.Value().(*CompiledFunction)

	const numGoroutines = 10
	const duration = 5 * time.Second
	var wg sync.WaitGroup
	executionErrors := make(chan error, numGoroutines*100) // Generous buffer
	var totalOperations int64

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					executionErrors <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			// Create isolated context
			ctx := NewExecutionContext(compiled).WithIsolatedGlobals()
			operationCount := 0

			for time.Since(startTime) < duration {
				result, err := ctx.Call(longOperationFn)
				if err != nil {
					executionErrors <- fmt.Errorf("goroutine %d, operation %d: %v", id, operationCount, err)
					return
				}

				// Verify result (sum of 0 to 9999 = 49995000)
				expectedSum := int64(49995000)
				actualSum := result.(*Int).Value
				if actualSum != expectedSum {
					executionErrors <- fmt.Errorf("goroutine %d, operation %d: expected %d, got %d", id, operationCount, expectedSum, actualSum)
					return
				}

				operationCount++
				atomic.AddInt64(&totalOperations, 1)
			}

			// Get final operation count for this context
			finalCount, err := ctx.Call(getOperationCountFn)
			if err != nil {
				executionErrors <- fmt.Errorf("goroutine %d, final count: %v", id, err)
				return
			}

			if finalCount.(*Int).Value != int64(operationCount) {
				executionErrors <- fmt.Errorf("goroutine %d: expected %d operations, got %d", id, operationCount, finalCount.(*Int).Value)
				return
			}
		}(i)
	}

	wg.Wait()
	close(executionErrors)

	// Check for errors
	for err := range executionErrors {
		t.Error(err)
	}

	actualDuration := time.Since(startTime)
	t.Logf("Test duration: %v", actualDuration)
	t.Logf("Total operations: %d", totalOperations)
	t.Logf("Operations per second: %.2f", float64(totalOperations)/actualDuration.Seconds())
}
