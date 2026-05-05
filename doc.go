// Package boba provides a platform-polymorphic entry point for running
// BubbleTea programs that target both native terminals and web browsers.
//
// The same main.go works on both targets: when compiled for a native
// platform, [Run] runs the model with a standard tea.Program; when
// compiled with GOOS=js GOARCH=wasm, [Run] installs the JavaScript
// bridge that boba's web terminal uses.
//
//	package main
//
//	import (
//	    "log"
//
//	    boba "github.com/btwiuse/boba"
//	)
//
//	func main() {
//	    if err := boba.Run(initialModel()); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// For finer control — or to match the original Bubble Tea API during
// porting — use [NewProgram]:
//
//	bp := boba.NewProgram(initialModel())
//	if _, err := bp.Run(); err != nil {
//	    log.Fatal(err)
//	}
//
// For more granular control, subpackages are available:
//   - [github.com/btwiuse/boba/wasm] exposes the browser bridge directly.
//   - [github.com/btwiuse/boba/serve] runs a BubbleTea program as an
//     HTTP/WebSocket/WebTransport backend for browser clients.
package boba
