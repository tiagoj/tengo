# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.17.1] - 2025-01-19

### Added
- Enhanced closure functionality with access to global variables
- Comprehensive closure testing suite with advanced edge cases
- Performance benchmarking for closure execution
- Call method to CompiledFunction for direct function invocation
- Advanced closure access pattern tests
- Plan B documentation and rollback instructions for VM changes

### Changed
- Updated module path from upstream to fork (`github.com/tiagoj/tengo/v2`)
- Enhanced VM frame handling for direct function calls
- Improved README to highlight enhanced fork features
- Updated all import statements and documentation to use fork module path
- Code formatting with `go fmt`

### Fixed
- VM frame handling issues for direct function calls
- Module path references throughout codebase
- Import statements in tests, examples, and documentation

### Documentation
- Added comprehensive TENGO_ENHANCEMENT_PLAN with progress tracking
- Updated README with closure-with-globals feature highlights
- Added performance benchmarking documentation
- Created rollback and fallback documentation

## [v2.17.0] - 2024-XX-XX (Upstream)

### Fixed
- Fixed regex alternation (#460)

---

## Release Notes

### v2.17.1 - Enhanced Fork with Closure Improvements

This release represents a significant enhancement of the Tengo scripting language, focusing on advanced closure functionality and comprehensive testing. The fork now provides enhanced closure capabilities that allow closures to access global variables, making it more powerful for complex scripting scenarios.

**Key Highlights:**
- **Enhanced Closure System**: Closures can now access global variables, providing more flexible scripting capabilities
- **Comprehensive Testing**: Added extensive test suites covering edge cases and performance scenarios  
- **Performance Optimized**: Includes benchmarking and performance analysis for closure operations
- **Clean Module Path**: Updated to use proper fork module path for easy integration
- **Battle Tested**: Includes comprehensive edge case testing and validation

**Migration from Upstream:**
If migrating from the upstream d5/tengo repository, simply update your import statements:
```go
// Old
import "github.com/d5/tengo/v2"

// New  
import "github.com/tiagoj/tengo/v2"
```

**Compatibility:**
- Fully backward compatible with existing Tengo v2.17.0 code
- Enhanced closure features are opt-in and don't break existing functionality
- All existing APIs remain unchanged

For detailed technical documentation and examples, see the README.md file.
