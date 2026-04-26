# Changelog

All notable changes to the Tuseme Go SDK will be documented in this file.

## [1.0.0] - 2026-04-25

### Added
- Initial release of the Tuseme Go SDK.
- `NewClient()` with functional options pattern.
- `Messages.Send()` — send SMS to one or more recipients.
- `Messages.Get()` — check delivery status of a message.
- `Messages.List()` — list sent messages with pagination.
- Thread-safe token management with `sync.Mutex`.
- Built-in retry logic with exponential backoff.
- Zero external dependencies — uses only Go standard library.
