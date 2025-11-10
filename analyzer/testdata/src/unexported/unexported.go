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

// read reads everything.
func readAll() {} // want `doc comment starts with 'read' but symbol is 'readAll' \(possible typo or old name\)`

// wsStreamHandler handles websocket streams.
func wsStreamHandlerV1() {} // want `doc comment starts with 'wsStreamHandler' but symbol is 'wsStreamHandlerV1' \(possible typo or old name\)`

// findDBPathsById locates DB paths.
func findDBPathsByID() {} // want `doc comment starts with 'findDBPathsById' but symbol is 'findDBPathsByID' \(possible typo or old name\)`

// generates numAccounts keys for reproducible fixtures. (narrative, no diagnostic expected)
func generateKeys() {}

// note: helper for tests (label should be skipped)
func notify() {}
