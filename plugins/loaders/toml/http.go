package toml

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	telegraf "github.com/influxdata/tgconfig"
)

const (
	HTTPName = "http"
)

type HTTPConfig struct {
	Origin string
}

type HTTP struct {
	origin *url.URL
	client *http.Client
}

func NewHTTP(config *HTTPConfig) (telegraf.Loader, error) {
	origin, err := url.Parse(config.Origin)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	http := &HTTP{
		origin: origin,
		client: client,
	}
	return http, nil
}

func (c *HTTP) Name() string {
	return HTTPName
}

func (c *HTTP) Load(ctx context.Context, registry *telegraf.PluginRegistry) (*telegraf.Config, error) {
	url := *c.origin
	url.Path = "/config"
	req, err := http.NewRequest("GET", url.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parser := NewParser(registry)
	return parser.Parse(resp.Body)
}

func (c *HTTP) URLWithPath(path string) *url.URL {
	url := *c.origin
	url.Path = path
	return &url
}

func (c *HTTP) Monitor(ctx context.Context) error {
	url := c.URLWithPath("/config/poll")
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return telegraf.ReloadConfig
}

// MonitorC is an example of using a long poll http request for monitoring
func (c *HTTP) MonitorC(ctx context.Context) (<-chan error, error) {
	out := make(chan error, 1)

	url := c.URLWithPath("/config/poll")
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	go func() {
		resp, err := c.client.Do(req.WithContext(ctx))
		if err != nil {
			out <- err
			close(out)
			return
		}
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		out <- telegraf.ReloadConfig
		close(out)
	}()

	return out, nil
}

// StartWatch establishes the watch
func (h *HTTP) StartWatch(ctx context.Context) error {
	// In order to avoid missing events after this function ends, the http
	// client would need to ask for events after an event number in the
	// WaitWatch, or use another custom method depending on the backend.
	return nil
}

// WaitWatch blocks until the Loader should be reloaded
func (h *HTTP) WaitWatch(ctx context.Context) error {
	url := h.URLWithPath("/config/poll")
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return nil
}

// Debugging
func (c *HTTP) String() string {
	return "Config: http"
}
