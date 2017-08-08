package telegraf

import "fmt"

// ParserConfig contains global Parser configuration.
type ParserConfig struct {
	DataFormat string `toml:"data_format"`
}

func (p *ParserConfig) String() string {
	return fmt.Sprintf("  parser:data_format:%s", p.DataFormat)
}
