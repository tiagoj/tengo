package tengo

import (
	"fmt"
	"testing"
)

func TestDebugFrameIssue(t *testing.T) {
	script := NewScript([]byte(`
		add := func(x, y) {
			return x + y
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

	addVar := compiled.Get("add")
	if addVar == nil {
		t.Fatal("add function not found")
	}

	addFn, ok := addVar.Value().(*CompiledFunction)
	if !ok {
		t.Fatal("add is not a CompiledFunction")
	}

	// Print debugging info
	fmt.Printf("Function instructions length: %d\n", len(addFn.Instructions))
	fmt.Printf("Function NumLocals: %d\n", addFn.NumLocals)
	fmt.Printf("Function NumParameters: %d\n", addFn.NumParameters)
	fmt.Printf("Function VarArgs: %v\n", addFn.VarArgs)
	fmt.Printf("Function Free vars: %d\n", len(addFn.Free))

	// Print the bytecode
	fmt.Println("Bytecode:")
	for i, b := range addFn.Instructions {
		fmt.Printf("%d: %02x\n", i, b)
	}

	constants := compiled.Constants()
	globals := compiled.Globals()

	fmt.Printf("Constants length: %d\n", len(constants))
	fmt.Printf("Globals length: %d\n", len(globals))

	// Try to understand what happens when we create the VM
	vm := &VM{
		constants:   constants,
		sp:          0,
		globals:     globals,
		fileSet:     nil,
		framesIndex: 1,
		ip:          -1,
		maxAllocs:   -1,
	}

	// Set up the function frame
	vm.curFrame = &vm.frames[0]
	vm.curFrame.fn = addFn
	vm.curFrame.freeVars = addFn.Free
	vm.curFrame.ip = -1
	vm.curFrame.basePointer = vm.sp
	vm.curInsts = addFn.Instructions
	vm.ip = -1

	fmt.Printf("Initial VM state:\n")
	fmt.Printf("  sp: %d\n", vm.sp)
	fmt.Printf("  framesIndex: %d\n", vm.framesIndex)
	fmt.Printf("  ip: %d\n", vm.ip)
	fmt.Printf("  curFrame.basePointer: %d\n", vm.curFrame.basePointer)

	// Add arguments to stack
	args := []Object{&Int{Value: 10}, &Int{Value: 20}}
	for _, arg := range args {
		vm.stack[vm.sp] = arg
		vm.sp++
	}

	fmt.Printf("After adding args:\n")
	fmt.Printf("  sp: %d\n", vm.sp)

	// Allocate space for local variables
	for i := len(args); i < addFn.NumLocals; i++ {
		vm.stack[vm.sp] = UndefinedValue
		vm.sp++
	}

	fmt.Printf("After allocating locals:\n")
	fmt.Printf("  sp: %d\n", vm.sp)

	// This is where the panic will occur
	fmt.Println("About to run VM...")
	err = vm.Run()
	if err != nil {
		fmt.Printf("VM error: %v\n", err)
	}
}
