package telegraf

// Existing: plugins/parsers/registry.Parser
type Parser interface {
	Parse(buf []byte) ([]Metric, error)
}
