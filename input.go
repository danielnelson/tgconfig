package telegraf

// Input is an input plugin.
type Input interface {
	Gather() error
}

// plugins/parsers/registry.go
type ParserInput interface {
	// SetParser sets the parser function for the interface
	SetParser(parser Parser)
}
