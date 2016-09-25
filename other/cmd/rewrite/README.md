Rewrite Tool
============

This `rewrite` tool rewrites source code to keep track of the current `context.Context` in goroutine-local storage. Whenever a function returns a `context.Context` object, it is set as the calling goroutine's new local `context.Context`, and when a new goroutine is spawned, if its first argument is a `context.Context` object, that is set as the new goroutine's initial local `context.Context`.
