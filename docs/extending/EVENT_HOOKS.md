# Event Hooks [PLANNED]

<!-- events: planned -->

An event hook system is **not implemented** in the current codebase. When (and if) events are added, document the real payloads here and keep the list above in sync. Until then, downstream tools should not rely on events.

## How to add (future)
1. Implement event types and emitters in code.
2. Add a verified list of event names to this file (replace the marker above).
3. Update `cmd/verify-docs` to include the new events and payload schemas.
