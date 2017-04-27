package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c, err := NewClient("XXX")
	if err != nil {
		t.Error(err)
	}

	if c.Host == "" {
		t.Error("missing host")
	}

	if c.APIKey != "XXX" {
		t.Error("set wrong API key")
	}
}

func TestFetch(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{"foo":"bar"}`)
	}))

	defer s.Close()

	c, err := NewClient("XXX")
	if err != nil {
		t.Error(err)
	}

	c.Host = s.URL

	data, err := c.fetch()
	if err != nil {
		t.Error(err)
	}

	foo, ok := data["foo"].(string)
	if !ok {
		t.Error("did not return map")
	}
	if foo != "bar" {
		t.Errorf("expecting 'bar' got '%v'", foo)
	}
}

func TestFetch400Error(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	defer s.Close()

	c, err := NewClient("XXX")

	if err != nil {
		t.Error(err)
	}

	c.Host = s.URL

	_, err = c.fetch()
	if err == nil {
		t.Error("did not error")
	}
}

func TestFetchEmtpyResponse(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	defer s.Close()

	c, err := NewClient("XXX")
	if err != nil {
		t.Error(err)
	}

	c.Host = s.URL

	_, err = c.fetch()
	if err == nil {
		t.Error("did not error")
	}
}

func TestStart(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{"foo":"bar"}`)
	}))

	defer s.Close()

	c, err := NewClient("XXX")
	if err != nil {
		t.Error(err)
	}

	c.Host = s.URL
	c.Timeout = time.Microsecond

	err = c.Start()
	if err != nil {
		t.Error(err)
	}

	if c.ready != true {
		t.Error("not ready")
	}

	time.Sleep(time.Microsecond * 10)

	c.Stop()
}

func TestLoopSuccess(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{"foo":"bar"}`)
	}))

	defer s.Close()

	c, err := NewClient("XXX")
	if err != nil {
		t.Error(err)
	}

	c.Host = s.URL
	c.cache = map[string]interface{}{
		"qux": "mux",
	}

	c.loop()

	foo, _ := c.cache["foo"].(string)
	if foo != "bar" {
		t.Error("did not update values")
	}
}

func TestLoopError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	defer s.Close()

	c, err := NewClient("XXX")
	if err != nil {
		t.Error(err)
	}

	invokedErrorHandler := false

	c.Host = s.URL
	c.ErrorHandler = func(err error) {
		invokedErrorHandler = true
	}
	c.loop()

	if !invokedErrorHandler {
		t.Error("did not invoke error handler")
	}
}

func TestGetNotPrepared(t *testing.T) {
	c, err := NewClient("XXX")
	if err != nil {
		t.Error(err)
	}

	defer func() {
		recover()
	}()
	c.get("foo")
}

func TestGetBoolean(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{ "str":"hello", "true":true, "false":false }`)
	}))

	defer s.Close()

	c, err := NewClient("XXX")
	if err != nil {
		t.Error(err)
	}

	c.Host = s.URL
	if err := c.Start(); err != nil {
		t.Error(err)
	}

	if _, ok := c.GetBoolean("not real"); ok != false {
		t.Error("returned ok non-existing key")
	}

	if val, _ := c.GetBoolean("true"); val != true {
		t.Errorf("expected `true` got `%v`", val)
	}

	if val, _ := c.GetBoolean("false"); val != false {
		t.Errorf("expected `false` got `%v`", val)
	}

	if _, ok := c.GetBoolean("str"); ok != false {
		t.Error("returned string as bool")
	}

	c.Stop()
}

func TestGetString(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{ "str":"hello", "true":true, "false":false }`)
	}))

	defer s.Close()

	c, err := NewClient("XXX")
	if err != nil {
		t.Error(err)
	}

	c.Host = s.URL
	if err := c.Start(); err != nil {
		t.Error(err)
	}

	if _, ok := c.GetString("not real"); ok != false {
		t.Error("returned ok non-existing key")
	}

	if val, _ := c.GetString("str"); val != "hello" {
		t.Errorf("expected `hello` got `%v`", val)
	}

	c.Stop()
}
