package unexported

// serveHtpp handles websocket traffic.
func serveHTTP() {} // want `doc comment starts with 'serveHtpp' but symbol is 'serveHTTP' \(possible typo or old name\)`

type telemetryFilteredHook struct{}

// newTelemetryFilterdHook creates the filtered hook.
func newTelemetryFilteredHook() telemetryFilteredHook { // want `doc comment starts with 'newTelemetryFilterdHook' but symbol is 'newTelemetryFilteredHook' \(possible typo or old name\)`
	return telemetryFilteredHook{}
}

type fooServer struct{}

// fooServer.serveHTTP handles HTTP requests.
func (fooServer) serveHTTPv1() {} // want `doc comment starts with 'serveHTTP' but symbol is 'serveHTTPv1' \(possible typo or old name\)`

// decodePage updates cache entries.
func encodePage() {} // want `doc comment starts with 'decodePage' but symbol is 'encodePage' \(possible typo or old name\)`

// Read reads everything but intentionally starts with a verb and should be treated as narrative.
func readAll() {}

// wsStreamHandler handles websocket streams.
func wsStreamHandlerV1() {} // want `doc comment starts with 'wsStreamHandler' but symbol is 'wsStreamHandlerV1' \(possible typo or old name\)`

// findDBPathsById locates DB paths.
func findDBPathsByID() {} // want `doc comment starts with 'findDBPathsById' but symbol is 'findDBPathsByID' \(possible typo or old name\)`

// generates numAccounts keys for reproducible fixtures. (narrative, no diagnostic expected)
func generateKeys() {}

// note: helper for tests (label should be skipped)
func notify() {}

// ServeHTTP handles requests but the identifier is unexported so the analyzer should still flag it.
func serveHHTP() {} // want `doc comment starts with 'ServeHTTP' but symbol is 'serveHHTP' \(possible typo or old name\)`

/**
 * ServeHTTPBlock handles requests but block comments currently confuse the tokenizer.
 */
func serveHHTPBlock() {} // want `doc comment starts with 'ServeHTTPBlock' but symbol is 'serveHHTPBlock' \(possible typo or old name\)`

type handler interface {
	// ServeHTTP handles requests in interface declarations as well.
	serveHHTP()
}
