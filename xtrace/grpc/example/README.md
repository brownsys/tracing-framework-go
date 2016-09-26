Example Service
===============

This is an example service built using the gRPC shim. It performs simple mathematical operations by having one service per component of an operation. For example, in order to compute (x * 5) + 3, one service would be run to multiply its argument by 5, and another service would be run to add 3 to its argument. The output of the first service would be fed to the second, and clients could make calls to the first service. Using the gRPC shim, X-Trace IDs are propagated alongside the RPC calls, and used when sending log messages to the X-Trace server.
