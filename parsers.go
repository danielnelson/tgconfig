package telegraf

// Existing: plugins/parsers/registry.Parser
type Parser interface {
	Parse(buf []byte) ([]Metric, error)
}

// ParserInput is an Input that allows setting of a Parser.
//
// Existing: plugins/parsers/registry.ParserInput
type ParserInput interface {
	SetParser(parser Parser)
}
