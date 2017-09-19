package models

import (
	"fmt"

	telegraf "github.com/influxdata/tgconfig"
)

// ParserConfig contains the global Parser configuration.
type ParserConfig struct {
	DataFormat string `toml:"data_format"`
}

func (p *ParserConfig) String() string {
	return fmt.Sprintf("  parser:data_format:%s", p.DataFormat)
}

func NewParser() (telegraf.Parser, error) {
	// use reflection with parsers.Parsers to construct
	return nil, fmt.Errorf("not implemented")
}
