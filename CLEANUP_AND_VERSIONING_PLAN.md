# Repository Cleanup and Versioning Plan

This document outlines the actions to clean up metadata and module references in our Tengo fork, update the module naming, and create a new semantic version tag v2.17.1. All references to "the plan" in conversations most likely refer to this document.

## Current Observations (As of Plan Completion)

- The `go.mod` now correctly declares:  
  `module github.com/tiagoj/tengo/v2`  
  which points to our fork's module path.

- Version tags now exist: v2.17.0 and v2.17.1 have been created.

- Default branch is `master`, and only one branch exists.

- Documentation and code now reference the correct fork module path.

## Cleanup and Versioning Goals ✅ COMPLETE

1. ✅ Update `go.mod` module path to reflect our fork repository path (`github.com/tiagoj/tengo/v2`).
2. ✅ Update any import statements in the code base that import the upstream module path to the new fork module path.
3. ✅ Confirm all README links and documentation reflect new repository paths (including go get instructions).
4. ✅ Create a new annotated Git tag `v2.17.1` reflecting the latest stable changes that include the closure enhancements.
5. ✅ Push the tag to the remote GitHub repository.
6. ✅ Verify module usage by building and running tests after module path updates.
7. Optional: Create a GitHub release for `v2.17.1` capturing the release notes.
8. Optional: Add changelog if not present to document current changes compared to `v2.17.0`.

---

## Detailed Steps

### Step 1: Change Module Path in `go.mod` ✅ COMPLETE

- ✅ Edit `go.mod` file:
  ```
  module github.com/tiagoj/tengo/v2
  go 1.21
  ```

- ✅ Run `go mod tidy` to update module dependencies accordingly.

### Step 2: Update Import Paths in Code ✅ COMPLETE

- ✅ Recursively search files for upstream module path and replace with fork module path.
- ✅ This includes internal imports, tests, examples, and documentation code blocks.

### Step 3: Update README and Docs ✅ COMPLETE

- ✅ Update `README.md` lines referencing `go get` to use new fork path.
- ✅ Confirm links to examples and API docs reflect new paths if relative links break.

### Step 4: Commit All Changes ✅ COMPLETE

- ✅ Commit with message:
  ```
  Update module path and references for fork usage
  ```

### Step 5: Tag and Release ✅ COMPLETE

- ✅ Create annotated tag:
  ```
  git tag -a v2.17.1 -m "Release v2.17.1 with closure enhancements and clean module path"
  git push origin v2.17.1
  ```

### Step 6: Verification ✅ COMPLETE

- ✅ Run `go build ./...` and `go test ./...` to ensure no import errors.
- ✅ Try using the module in a test module outside the repo to verify correct module path resolution.
- ✅ Core functionality verified - module works successfully from external projects.

### Step 7: GitHub Release (Optional) ✅ COMPLETE

**Prerequisites:**
- ✅ GitHub CLI (`gh`) is now available
- ✅ Authentication completed: `gh auth login`
- ✅ Default repository set: `gh repo set-default tiagoj/tengo`

**Steps completed:**
1. ✅ Authenticated with GitHub CLI: `gh auth login`
2. ✅ Set default repository: `gh repo set-default tiagoj/tengo`
3. ✅ Created the release with command:
   ```bash
   gh release create v2.17.1 --title "v2.17.1 - Enhanced Fork with Closure Improvements" --notes-file CHANGELOG.md
   ```

**Result:** ✅ GitHub release v2.17.1 successfully created at:
https://github.com/tiagoj/tengo/releases/tag/v2.17.1

### Step 8: CHANGELOG.md Creation (Optional) ✅ COMPLETE

- ✅ Create a CHANGELOG.md file documenting version history
- ✅ Include changes from v2.17.0 to v2.17.1:
  - Enhanced closure functionality with globals access
  - Comprehensive testing suite additions
  - Performance benchmarking
  - Documentation updates
  - Module path corrections
- ✅ Follow standard changelog format (Keep a Changelog)
- ✅ Include migration guide and compatibility notes

### Step 9: Plan Document Corrections (Optional) ✅ COMPLETE

- ✅ Fix errors in "Current Observations" section:
  - ✅ Updated to show actual current module path
  - ✅ Removed duplicate reference to same module path
  - ✅ Fixed search/replace of same path error
- ✅ Added completion markers throughout the document
- ✅ Updated status summary to reflect current state

---

## Status Summary

**🎉 ENTIRE PLAN: ✅ COMPLETE**
- All essential steps (1-6) have been completed successfully
- All optional enhancements (7-9) have been completed successfully
- Module is fully functional and can be used from external projects
- Tag v2.17.1 created and pushed to remote
- GitHub release v2.17.1 published with comprehensive release notes
- Complete documentation and changelog in place

**FINAL DELIVERABLES:**
- ✅ Updated module path: `github.com/tiagoj/tengo/v2`
- ✅ Git tag v2.17.1 with all enhancements
- ✅ GitHub release v2.17.1 with detailed release notes
- ✅ Comprehensive CHANGELOG.md documentation
- ✅ All imports and documentation updated
- ✅ Module verified and tested

---

## Additional Notes

- Coordinate with repository maintainers or document governance if applicable.
- Update any CI/CD workflows that use module paths.
- Consider creating release branches or tags for long-term maintenance.
- GitHub CLI (`gh`) is now available - requires authentication for release creation

---
