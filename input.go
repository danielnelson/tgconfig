package telegraf

// Input is an input plugin.
type Input interface {
	Gather() error
}
