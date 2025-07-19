# Repository Cleanup and Versioning Plan

This document outlines the actions to clean up metadata and module references in our Tengo fork, update the module naming, and create a new semantic version tag v2.17.1. All references to "the plan" in conversations most likely refer to this document.

## Current Observations

- The `go.mod` currently declares:  
  `module github.com/tiagoj/tengo/v2`  
  which points to the original upstream module path, not our fork.

- No version tags exist in the fork for releases (e.g., v2.17.0 of upstream not present).

- Default branch is `master`, and only one branch exists.

- Documentation and code reference the upstream module path.

## Cleanup and Versioning Goals

1. Update `go.mod` module path to reflect our new repository path (e.g., `github.com/tiagoj/tengo/v2` or other).
2. Update any import statements in the code base that import the old module path (`github.com/tiagoj/tengo/v2`) to the new module path.
3. Confirm all README links and documentation reflect new repository paths (including go get instructions).
4. Create a new annotated Git tag `v2.17.1` reflecting the latest stable changes that include the closure enhancements.
5. Push the tag to the remote GitHub repository.
6. Optionally create a GitHub release for `v2.17.1` capturing the release notes.
7. Verify module usage by building and running tests after module path updates.
8. Optionally add changelog if not present to document current changes compared to `v2.17.0`.

---

## Detailed Steps

### Step 1: Change Module Path in `go.mod`

- Edit `go.mod` file:
  ```
  module github.com/tiagoj/tengo/v2
  go 1.21  # update Go version if desired
  ```

- Run `go mod tidy` to update module dependencies accordingly.

### Step 2: Update Import Paths in Code

- Recursively search files for `github.com/tiagoj/tengo/v2` and replace with `github.com/tiagoj/tengo/v2`.
- This includes internal imports, tests, examples, and documentation code blocks.

### Step 3: Update README and Docs

- Update `README.md` lines referencing `go get github.com/tiagoj/tengo/v2` to use new path.
- Confirm links to examples and API docs reflect new paths if relative links break.

### Step 4: Commit All Changes

- Commit with message:
  ```
  Update module path and references for fork usage
  ```

### Step 5: Tag and Release

- Create annotated tag:
  ```
  git tag -a v2.17.1 -m "Release v2.17.1 with closure enhancements and clean module path"
  git push origin v2.17.1
  ```

- (Optional) Create GitHub release matching the tag with release notes.

### Step 6: Verification

- Run `go build ./...` and `go test ./...` to ensure no import errors.
- Try using the module in a test module outside the repo to verify correct module path resolution.

---

## Additional Notes

- Coordinate with repository maintainers or document governance if applicable.
- Update any CI/CD workflows that use module paths.
- Consider creating release branches or tags for long-term maintenance.

---

If you approve, I will begin implementation by updating `go.mod` and import paths first.
