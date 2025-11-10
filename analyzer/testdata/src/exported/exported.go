package exported

import "testing"

// ServerHTTP handles HTTP traffic over websockets.
func ServeHTTP() {} // want `doc comment starts with 'ServerHTTP' but symbol is 'ServeHTTP' \(possible typo or old name\)`

// newFilteredTelemetryHook creates a hook filter for ensuring telemetry events are always included.
func NewTelemetryFilteredHook() {} // want `doc comment starts with 'newFilteredTelemetryHook' but symbol is 'NewTelemetryFilteredHook' \(possible typo or old name\)`

// TestCatchpoint_FastUpdates exercises the fast path.
func TestCatchpointFastUpdates() {} // want `doc comment starts with 'TestCatchpoint_FastUpdates' but symbol is 'TestCatchpointFastUpdates' \(possible typo or old name\)`

// TestCachePageLoading ensures reload path.
func TestCachePageReloading() {} // want `doc comment starts with 'TestCachePageLoading' but symbol is 'TestCachePageReloading' \(possible typo or old name\)`

// BenchMarkVerify benchmarks signature verification.
func BenchmarkVerify(b *testing.B) {} // want `doc comment starts with 'BenchMarkVerify' but symbol is 'BenchmarkVerify' \(possible typo or old name\)`

// TelemetryHistoryState stores prior hook state.
type TelemetryHistory struct{} // want `doc comment starts with 'TelemetryHistoryState' but symbol is 'TelemetryHistory' \(possible typo or old name\)`

type FooServer struct{}

// FooServer.ServeHTTP handles websocket traffic.
func (FooServer) ServeHTTPv1() {} // want `doc comment starts with 'ServeHTTP' but symbol is 'ServeHTTPv1' \(possible typo or old name\)`
