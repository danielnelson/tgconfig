package telegraf

// PluginType is an enum of the different plugin types.
type PluginType int

const (
	InputType PluginType = iota
	OutputType
	LoaderType
	ParserType
)

// AgentConfig contains the Agent configuration
type AgentConfig struct {
	Interval int `toml:"interval"`
}

// FilterConfig contains the standard filtering configuration.  We may need
// one of these for each of inputs, processors, aggregators, outputs.
type FilterConfig struct {
	NameOverride string `toml:"name_override"`
}

// ParserConfig is the shared configuration for Parsers.
type ParserConfig struct {
	DataFormat string `toml:"data_format"`
}

// CommonInputConfig is the configuration options that can be set on any Input.
type CommonInputConfig struct {
	FilterConfig
	ParserConfig
}

// CommonOutputConfig is the configuration options that can be set on any Output.
type CommonOutputConfig struct {
	FilterConfig
}

// CommonLoaderConfig is the configuration options that can be set on any Loader.
type CommonLoaderConfig struct {
}

// PluginConfig is a config struct for plugin.
type PluginConfig = interface{}

// PluginFactory function for creating plugins: func (*PluginConfig) (Plugin, error)
type PluginFactory = interface{}

// InputConfig is all configuration needed to create the Inputs.
type InputConfig struct {
	Config       *CommonInputConfig
	PluginConfig PluginConfig
	ParserConfig PluginConfig
}

// OutputConfig is all configuration needed to create the Outputs.
type OutputConfig struct {
	Config       *CommonOutputConfig
	PluginConfig PluginConfig
}

// LoaderConfig is all configuration needed to create the Loaders.
type LoaderConfig struct {
	Config       *CommonLoaderConfig
	PluginConfig PluginConfig
}

// Config is the full set of loadable configuration.
type Config struct {
	Agent   AgentConfig
	Inputs  map[string][]*InputConfig
	Outputs map[string][]*OutputConfig
	Loaders map[string][]*LoaderConfig
}

// Registry is an interface for creating known plugins.
type Registry interface {
	CreateInputs(name string, c PluginConfig) ([]Input, error)
	CreateOutputs(name string, c PluginConfig) ([]Output, error)
	CreateLoaders(name string, c PluginConfig) ([]Loader, error)

	CreateParser(name string, c PluginConfig) (Parser, error)

	GetConfigRegistry() ConfigRegistry
}

// ConfigRegistry is an interface that can create empty config structs.
type ConfigRegistry interface {
	GetPluginConfig(pluginType PluginType, name string) (PluginConfig, bool)
}
