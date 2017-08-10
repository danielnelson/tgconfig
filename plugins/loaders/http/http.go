package http

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/BurntSushi/toml"

	telegraf "github.com/influxdata/tgconfig"
	tomlplugin "github.com/influxdata/tgconfig/plugins/loaders/toml"
)

const (
	Name = "http"
)

type Config struct {
	Origin string
}

type HTTP struct {
	origin *url.URL
	client *http.Client
}

func New(config *Config) (telegraf.Loader, error) {
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
	return Name
}

type telegrafConfig struct {
	Inputs map[string][]toml.Primitive
}

func (c *HTTP) Load(ctx context.Context, plugins *telegraf.Plugins) (*telegraf.Config, error) {
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

	var md toml.MetaData
	conf := telegrafConfig{}
	if md, err = toml.DecodeReader(resp.Body, &conf); err != nil {
		return nil, err
	}

	ri, err := tomlplugin.LoadInputs(md, plugins, conf.Inputs)
	if err != nil {
		return nil, err
	}

	return &telegraf.Config{Inputs: ri}, nil
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
func (c *HTTP) MonitorC(ctx context.Context) <-chan error {
	out := make(chan error, 1)

	url := c.URLWithPath("/config/poll")
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		out <- err
		close(out)
		return out
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

	return out
}

// Debugging
func (c *HTTP) String() string {
	return "Config: http"
}
