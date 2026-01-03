# Changelog

All notable changes to the Go version of ChatGPT-Telegram-Workers will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **User Setting Permission Control**: New `ENABLE_USER_SETTING` environment variable to control whether regular users can modify their own configurations
  - When set to `true` (default): All users can modify their own settings
  - When set to `false`: Only administrators can modify configurations
- **Admin Authentication**: New `CHAT_ADMIN_KEY` environment variable to specify Telegram User IDs of administrators
  - Administrators can always modify configurations regardless of `ENABLE_USER_SETTING`
  - Falls back to Telegram group admin check if user is not in `CHAT_ADMIN_KEY`
- **Multi-Database Support**: Added support for MySQL and PostgreSQL in addition to SQLite
  - New `DSN` environment variable for database connection strings
  - Automatic database type detection based on DSN prefix (`mysql://`, `postgres://`, `postgresql://`, `sqlite://`)
  - GORM-based storage layer for consistent database operations across all database types
- **Permission System**: New `PermissionChecker` interface for extensible permission checking
  - `DefaultPermissionChecker` implementation with admin verification
  - Command visibility control based on user permissions
- **Dynamic Command Registration**: Commands are now registered dynamically based on configuration
  - Configuration commands are hidden from regular users when `ENABLE_USER_SETTING=false`

### Changed
- **Storage Layer**: Migrated from direct SQLite implementation to GORM-based abstraction
  - All database operations now use GORM models
  - Consistent API across SQLite, MySQL, and PostgreSQL
  - Automatic schema migration on startup
- **Configuration Loading**: Enhanced configuration system with new permission-related fields
  - `EnableUserSetting` defaults to `true` for backward compatibility
  - `ChatAdminKey` for admin user ID list
  - `DSN` for database connection (takes precedence over `DB_PATH`)
- **Command Registry**: Updated to support permission-based command filtering
  - Commands can now specify permission requirements
  - Command list is filtered based on user permissions

### Removed
- **Migration Tools**: Removed all migration-related code and documentation
  - Deleted `cmd/migrate` directory and all migration utilities
  - Removed `doc/MIGRATION.md` documentation
  - Removed `scripts/export_kv.sh` migration script
  - Removed migration-related build targets from Makefile
  - Removed migration instructions from README.md and deployment documentation
- **Version Command**: Removed `/version` command and version comparison functionality
  - Removed `BuildTimestamp` and `BuildVersion` fields from Config
  - Removed version display and update check logic
  - Removed version command registration and handler

### Fixed
- Improved error handling for database connection failures with clear error messages
- Enhanced DSN validation with detailed error reporting

### Documentation
- Updated `README.md` to remove migration references and add new configuration options
- Updated `doc/CONFIG.md` with detailed explanations of new environment variables
- Updated `doc/DEPLOY.md` with multi-database deployment examples
- Updated `.env.example` with new configuration options and examples

### Testing
- Added comprehensive unit tests for permission system
- Added unit tests for configuration permission control
- Added unit tests for GORM storage layer
- Added integration tests for database operations
- All tests passing with good coverage (>60% overall, 100% for i18n)

### Backward Compatibility
- `ENABLE_USER_SETTING` defaults to `true` to maintain existing behavior
- `DB_PATH` continues to work when `DSN` is not set
- Existing SQLite databases can be used without migration
- All existing functionality preserved except for removed features

## Notes

This release transforms the Go version into a standalone, production-ready application with:
- Enhanced security through granular permission control
- Scalability through multi-database support
- Simplified deployment by removing migration dependencies
- Cleaner codebase with removed legacy version comparison features
