### High-Level Plan for Branching and Implementing Tengo Closure Changes

## AI Hints
- **Current Working Directory**: `/Users/james/GitRepos/Public/tengo`
- **Branch**: `feat/closure-with-globals`
- **Key Architecture Discovery**: CompiledFunction doesn't have a Call method - it's handled directly in VM.run() OpCall case
- **Important Files**:
  - `vm.go` - Main VM implementation, OpCall handling (lines 573-669)
  - `objects.go` - Object definitions, UserFunction has Call method (lines 1605-1607)
  - `tengo.go` - CallableFunc type definition (line 35)
  - `builtins.go` - Builtin functions examples
  - `compiler.go` - Compilation logic, closure creation (lines 389-475)
  - `bytecode.go` - Bytecode operations
- **Key Insights**:
  - VM constructor: `NewVM(bytecode, globals, maxAllocs)` in vm.go:38
  - Globals are `[]Object` slice in VM struct (vm.go:24)
  - CompiledFunction execution bypasses Call method, goes directly through VM frames
  - UserFunction uses Call method pattern we want to emulate
  - Closures created with OpClosure instruction (vm.go:747-782)
- **Current Challenge**: Need to add CallWithGlobals to CompiledFunction and modify VM execution

To address the closure context loss issue (where invoking a closure via the Go API creates an isolated VM without the original globals), we'll extend Tengo by adding support for passing globals explicitly when calling closures. This will be done in a new method to avoid breaking existing behavior (e.g., `CallWithGlobals`). The plan emphasizes incremental steps, each with isolated changes, unit tests, and integration tests to ensure stability. We'll use Go's standard testing tools and Tengo's existing test suite as a foundation.

Assume you have Go installed (v1.18+ for Tengo compatibility) and Git. The total effort might take 4-8 hours for an experienced Go developer, spread over steps.

#### 1. **Setup and Branching (Preparation Phase)** - ✅ COMPLETE
   - **Actions:**
     - ✅ Fork the official Tengo repository (github.com/d5/tengo) on GitHub to your account.
     - ✅ Clone the fork locally: `git clone https://github.com/yourusername/tengo.git`.
     - ✅ Checkout the latest stable tag or master: `git checkout master` (or e.g., `v2.12.0` if pinning a version).
     - ✅ Create a feature branch: `git checkout -b feat/closure-with-globals`.
     - ✅ Set up the dev environment: Run `go mod tidy` to fetch dependencies, and build/test the baseline: `go test ./...` to confirm the repo is healthy (all tests should pass).
   - **Testing:**
     - ✅ Verify baseline: Ensure existing tests pass without changes.
   - **Milestone:** ✅ A clean, branched repo ready for modifications. Commit: "Initial branch setup."

#### 2. **Analysis and Design (Planning Phase)** - ✅ COMPLETE
   - **Actions:**
     - ✅ Review key files:
       - ✅ `objects.go`: Focus on the `CompiledFunction` struct (lines 570-621) - NO existing Call method
       - ✅ `vm.go`: Examine `NewVM` constructor (line 38) and OpCall handling (lines 573-669)
       - ✅ `compiler.go`: Understand how globals are compiled and closure creation (lines 389-475)
       - ✅ `tengo.go`: CallableFunc type definition (line 35)
       - ✅ Note: Globals are a `[]Object` slice in VM struct (line 24), indexed by compiled symbol IDs.
     - ✅ Design the change:
       - ✅ Add `Call(args ...Object) (Object, error)` method to `CompiledFunction` (following UserFunction pattern)
       - ✅ Add `CallWithGlobals(globals []Object, args ...Object) (Object, error)` method to `CompiledFunction`
       - ✅ Create isolated VM execution similar to UserFunction.Call but with globals support
       - ✅ Ensure free variables (`Free []*ObjectPtr`) and args are handled correctly
       - ✅ Make globals copying efficient (e.g., use `copy` to avoid mutation issues)
   - **Testing:** ✅ No code changes yet; manually inspect with `go vet` or static analysis tools.
   - **Milestone:** ✅ Document the design in a README note or commit message. Commit: "Design notes for closure globals support."

#### 3. **Incremental Implementation (Development Phase)**
   Break into small, testable commits. Each step builds on the previous, with changes limited to 1-2 files.

   - **Step 3.1: Add Call Method to CompiledFunction (Basic Implementation)** - ⏳ NEXT
     - **Actions:**
       - ⏳ In `objects.go`, add basic `Call` method to `CompiledFunction` following UserFunction pattern:
         ```go
         func (o *CompiledFunction) Call(args ...Object) (Object, error) {
             return o.CallWithGlobals(nil, args...)
         }
         ```
       - ⏳ This ensures CompiledFunction implements the same interface as UserFunction
     - **Testing:**
       - ⏳ Add unit tests in `objects_test.go`: Test basic Call functionality
       - ⏳ Run `go test .` to verify no regressions
     - **Milestone:** ⏳ CompiledFunction has Call method. Commit: "Add Call method to CompiledFunction."

   - **Step 3.2: Add CallWithGlobals Method to CompiledFunction (Core Implementation)** - ⏳ PENDING
     - **Actions:**
       - ⏳ In `objects.go`, add `CallWithGlobals` method to `CompiledFunction`:
         ```go
         func (o *CompiledFunction) CallWithGlobals(globals []Object, args ...Object) (Object, error) {
             // Create isolated VM with custom globals
             bytecode := &Bytecode{
                 MainFunction: o,
                 Constants:    []Object{}, // Empty constants for isolated execution
             }
             vm := NewVM(bytecode, globals, 1000) // Max allocs
             
             // Set up call frame manually (similar to OpCall logic)
             vm.curFrame.fn = o
             vm.curFrame.freeVars = o.Free
             vm.curFrame.basePointer = 0
             vm.curInsts = o.Instructions
             vm.ip = -1
             
             // Set up arguments on stack
             for i, arg := range args {
                 vm.stack[i] = arg
             }
             vm.sp = len(args)
             
             err := vm.Run()
             if err != nil {
                 return nil, err
             }
             
             if vm.sp > 0 {
                 return vm.stack[vm.sp-1], nil
             }
             return UndefinedValue, nil
         }
         ```
     - **Testing:**
       - ⏳ Add tests in `objects_test.go`: Test CallWithGlobals with different global scenarios
       - ⏳ Test cases: Call with nil globals, with globals, verify global access works
       - ⏳ Run `go test .`
     - **Milestone:** ⏳ CallWithGlobals implemented. Commit: "Add CallWithGlobals to CompiledFunction."

   - **Step 3.3: Handle Global Modifications and Propagation (Optional Enhancement)** - ⏳ PENDING
     - **Actions:**
       - ⏳ If closures modify globals (via OP_SET_GLOBAL), add a way to retrieve updated globals post-call (e.g., return them alongside the result: `CallWithGlobals(...) (Object, []Object, error)`).
       - ⏳ Update method to return updated globals if modified.
     - **Testing:**
       - ⏳ Extend previous tests: Include bytecode that sets a global, assert returned globals reflect changes.
       - ⏳ Edge cases: No modifications, out-of-bounds access (should error as in original).
     - **Milestone:** ⏳ Support for mutable globals. Commit: "Support returning updated globals from CallWithGlobals."

   - **Step 3.4: Integration with Script/Compiled (User-Facing Glue)** - ⏳ PENDING
     - **Actions:**
       - ⏳ In `script.go` or a new helper, add a way to extract globals from a *Compiled after script run (e.g., `Compiled.Globals() []Object` exporting the internal slice).
       - ⏳ Update docs/examples in README or `examples/` to show usage: Compile script, get globals, pass to closure call.
     - **Testing:**
       - ⏳ Add end-to-end tests in a new file like `integration_test.go`: Compile a script defining globals and a closure, run it, extract globals, invoke closure with them, assert correct behavior.
       - ⏳ Reproduce your original issue in a test, confirm fix.
       - ⏳ Run full suite: `go test ./...`.
     - **Milestone:** ⏳ Changes usable in a DSL context. Commit: "Add globals extraction and integration tests."

#### 4. **Validation and Refinement (Testing Phase)**
   - **Actions:**
     - Benchmark: Add benchmarks in relevant _test.go files (e.g., BenchmarkCallWithGlobals vs original Call) to check for perf regressions.
     - Code review: Use `go fmt`, `go vet`, `golint`; fix any issues.
     - Handle errors robustly (e.g., nil checks, type assertions).
   - **Testing:**
     - Full regression: Ensure all original tests pass.
     - Manual test: Build a sample Go program using your branched Tengo, reproduce the inline vs API closure issue, verify fix.
     - Coverage: Aim for >80% on new code via `go test -cover`.
   - **Milestone:** Stable, tested changes. Commit: "Refinements and benchmarks."

#### 5. **Deployment and Maintenance (Wrap-Up Phase)**
   - **Actions:**
     - Push branch to your fork: `git push origin feat/closure-with-globals`.
     - For personal use: Update your DSL project's go.mod to point to your fork/branch (e.g., `replace github.com/d5/tengo/v2 => ../path/to/local/tengo` or GitHub URL).
     - Optionally: Open a PR to upstream with your changes, including tests/docs.
     - Tag a version if needed: `git tag v2.12.1-custom`.
   - **Testing:** Integrate into your Go DSL code, test the full workflow.
   - **Milestone:** Ready for use/PR.

#### Risks and Tips
- **Backward Compatibility:** New method avoids breaking existing code.
- **Scope Creep:** Stick to globals; if modules or other contexts are needed, add as future steps.
- **Debugging:** Use `fmt.Printf` in VM for interim debugging, remove before commit.
- **If Stuck:** Revert to previous commit; each step is atomic.
- **Time Estimates:** Setup (30min), Analysis (1hr), Each Impl Step (1-2hr incl. tests), Validation (1hr).

This plan ensures incremental progress with quick feedback loops via tests, minimizing bugs. If you provide more details (e.g., exact Tengo version or sample code), I can refine it further.