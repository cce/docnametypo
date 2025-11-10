package fixes

// ServeHTTP handles websocket requests but the function stayed unexported.
func serveHHTP() {} // want `doc comment starts with 'ServeHTTP' but symbol is 'serveHHTP' \(possible typo or old name\)`

/**
 * ServeHTTPBlock handles block comment cases.
 */
func serveHHTPBlock() {} // want `doc comment starts with 'ServeHTTPBlock' but symbol is 'serveHHTPBlock' \(possible typo or old name\)`
