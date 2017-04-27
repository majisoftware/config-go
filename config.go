package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// Client is a thingy
type Client struct {
	Host         string
	APIKey       string
	ErrorC       chan error
	Timeout      time.Duration
	ErrorHandler func(error)
	ticker       *time.Ticker
	cache        map[string]interface{}
	lock         sync.RWMutex
	ready        bool
}

func defaultErrorHandler(err error) {
	fmt.Println("config error:", err)
}

// NewClient creates a Client
func NewClient(apikey string) (*Client, error) {
	c := &Client{
		Host:         "https://api.config.maji.cloud",
		APIKey:       apikey,
		ready:        false,
		Timeout:      5 * time.Second,
		ErrorHandler: defaultErrorHandler,
	}

	return c, nil
}

// Start fetches the initial data and starts polling
func (c *Client) Start() error {
	data, err := c.fetch()
	if err != nil {
		return err
	}

	c.cache = data
	c.ready = true
	c.ticker = time.NewTicker(c.Timeout)

	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.loop()
			}
		}
	}()

	return nil
}

// Stop quits polling
func (c *Client) Stop() {
	c.ticker.Stop()
}

func (c *Client) loop() {
	data, err := c.fetch()
	if err != nil {
		c.ErrorHandler(err)
	} else {
		c.lock.Lock()
		c.cache = data
		c.lock.Unlock()
	}
}

func (c *Client) fetch() (map[string]interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", c.Host+"/getConfig", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("authorization", "bearer "+c.APIKey)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("config: %d response", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// get returns a raw value and a boolean indicating whether or not the requested key was found
func (c *Client) get(key string) (interface{}, bool) {
	if !c.ready {
		panic("config: not prepared")
	}

	c.lock.RLock()
	raw, ok := c.cache[key]
	c.lock.RUnlock()

	return raw, ok
}

// GetBoolean returns the boolean value of `key` as well as an `ok` boolean indicating whether the requested key was found
func (c *Client) GetBoolean(key string) (bool, bool) {
	raw, ok := c.get(key)
	if !ok {
		return false, false
	}

	val, ok := raw.(bool)
	return val, ok
}

// GetString returns the string value of `key` as well as an `ok` boolean indicating whether the requested key was found
func (c *Client) GetString(key string) (string, bool) {
	raw, ok := c.get(key)
	if !ok {
		return "", false
	}

	val, ok := raw.(string)
	return val, ok
}
