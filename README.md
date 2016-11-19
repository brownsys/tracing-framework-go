tracing-framework-go
====================

## Setup
Add the `bin` directory to your `PATH` before whatever directory contains your system's `go` binary. `bin` contains `bin/go`, a script which uses the appropriate Go installation for your architecture (see the `go` directory). You can now use the `go` command as normal.


## Usage
The two packages which should be used by normal consumers are the `xtrace/client` and `xtrace/grpc` packages. Both have their primary documentation in doc comments in the code; view using the standard `go doc` tools or `godoc.org`.

`xtrace/client` provides an X-Trace client that provides a simple logging interface. `xtrace/grpc` provides wrappers around standard `grpc` functions that propagate X-Trace state. Use these in place of the normal `grpc` functions.

## Code Rewriting
The `local` package (which you shouldn't have to import directly, but is used by the `xtrace/client` package) requires code to be rewritten in order to work properly. Use the tool in `cmd/rewrite` to rewrite each package that you want to be capable of propagating X-Trace state when new goroutines are spawned. Note that some standard library or third party packages could spawn goroutines which call callbacks which, if defined in your code, could contain logging statements or gRPC calls that need to consume or propagate X-Trace state; you may want to rewrite these packages in addition to your own packages. Rewriting standard library packages has not been thoroughly tested, but it should in theory be completely safe.
