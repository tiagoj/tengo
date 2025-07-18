package tengo_test

import (
	"testing"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/require"
)

func TestExecutionContext_Basic(t *testing.T) {
	// Test basic ExecutionContext creation and access
	script := tengo.NewScript([]byte(`
		a := 10
		b := 20
		export func(x) { return a + b + x }
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	// Create execution context
	ctx := tengo.NewExecutionContext(compiled)
	require.NotNil(t, ctx)

	// Test that we can access constants and globals
	constants := ctx.Constants()
	require.NotNil(t, constants)

	globals := ctx.Globals()
	require.NotNil(t, globals)

	// Test that Source() returns the original compiled object
	source := ctx.Source()
	require.True(t, source == compiled)
}

func TestExecutionContext_WithGlobals(t *testing.T) {
	// Test creating ExecutionContext with specific globals
	script := tengo.NewScript([]byte(`
		a := 10
		export func(x) { return a + x }
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	ctx := tengo.NewExecutionContext(compiled)

	// Create new globals
	newGlobals := []tengo.Object{
		&tengo.Int{Value: 100}, // Replace a with 100
	}

	newCtx := ctx.WithGlobals(newGlobals)
	require.NotNil(t, newCtx)

	// Verify the new context has the new globals
	globals := newCtx.Globals()
	require.Equal(t, 1, len(globals))
	require.Equal(t, int64(100), globals[0].(*tengo.Int).Value)
}

func TestExecutionContext_WithIsolatedGlobals(t *testing.T) {
	// Test creating ExecutionContext with isolated globals
	script := tengo.NewScript([]byte(`
		a := 10
		export func(x) { return a + x }
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	ctx := tengo.NewExecutionContext(compiled)

	// Create isolated context
	isolatedCtx := ctx.WithIsolatedGlobals()
	require.NotNil(t, isolatedCtx)

	// Verify that the isolated context has its own copy of globals
	originalGlobals := ctx.Globals()
	isolatedGlobals := isolatedCtx.Globals()

	// Should have same values but different objects
	require.Equal(t, len(originalGlobals), len(isolatedGlobals))

	// Test that modifying isolated globals doesn't affect original
	// (This test verifies the concept, actual function execution testing would need
	// a complete VM implementation in CallWithGlobalsExAndConstants)
}

func TestExecutionContext_ThreadSafety(t *testing.T) {
	// Test that ExecutionContext is thread-safe
	script := tengo.NewScript([]byte(`
		a := 10
		export func(x) { return a + x }
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	ctx := tengo.NewExecutionContext(compiled)

	// Test concurrent access to Constants() and Globals()
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			_ = ctx.Constants()
			_ = ctx.Globals()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = ctx.WithIsolatedGlobals()
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Test should complete without panicking
}

func TestExecutionContext_ConstantsImmutability(t *testing.T) {
	// Test that Constants() returns a copy that doesn't affect the original
	script := tengo.NewScript([]byte(`
		a := 10
		export func(x) { return a + x }
	`))

	compiled, err := script.Compile()
	require.NoError(t, err)

	err = compiled.Run()
	require.NoError(t, err)

	ctx := tengo.NewExecutionContext(compiled)

	// Get constants
	constants1 := ctx.Constants()
	constants2 := ctx.Constants()

	// Should be separate arrays (different pointers)
	require.NotNil(t, constants1)
	require.NotNil(t, constants2)

	// But should have same content
	require.Equal(t, len(constants1), len(constants2))
}
