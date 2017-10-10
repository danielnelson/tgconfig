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

func (c *HTTP) Load(ctx context.Context, registry *telegraf.ConfigRegistry) (*telegraf.Config, error) {
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

func (c *HTTP) Watch(ctx context.Context) (telegraf.Waiter, error) {
	url := c.URLWithPath("/config/poll")
	return NewHTTPWaiter(ctx, c.client, url.String())
}

type HTTPWaiter struct {
	ctx  context.Context
	resp *http.Response
}

func NewHTTPWaiter(ctx context.Context, client *http.Client, url string) (*HTTPWaiter, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	w := &HTTPWaiter{ctx: ctx}
	w.resp, err = client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (w *HTTPWaiter) Wait() error {
	defer w.resp.Body.Close()

	_, err := ioutil.ReadAll(w.resp.Body)
	if err != nil {
		return err
	}
	return w.ctx.Err()
}

// Debugging
func (c *HTTP) String() string {
	return "Config: http"
}
