package telegraf

// Output is an output plugin
type Output interface {
	Connect() error
}
