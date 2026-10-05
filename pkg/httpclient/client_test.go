package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captured struct {
	method   string
	path     string
	header   http.Header
	body     string
	user     string
	password string
}

func newServer(t *testing.T, status int, body string) (*httptest.Server, *captured) {
	t.Helper()
	got := &captured{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		got.method, got.path, got.header, got.body = r.Method, r.URL.Path, r.Header, string(raw)
		got.user, got.password, _ = r.BasicAuth()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server, got
}

func TestClient_Get(t *testing.T) {
	server, got := newServer(t, http.StatusOK, `{"name":"ana"}`)
	client := New(WithBaseURL(server.URL+"/"), WithHeader("X-Api-Key", "k1"))

	response, err := client.Get(context.Background(), "/v1/residents")

	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, got.method)
	assert.Equal(t, "/v1/residents", got.path, "trailing slash of the base URL is trimmed")
	assert.Equal(t, "k1", got.header.Get("X-Api-Key"))
	assert.True(t, response.IsSuccess())

	var decoded struct{ Name string }
	require.NoError(t, response.DecodeJSON(&decoded))
	assert.Equal(t, "ana", decoded.Name)
}

func TestClient_PostForm(t *testing.T) {
	server, got := newServer(t, http.StatusCreated, `{}`)
	client := New(WithBaseURL(server.URL), WithBasicAuth("user", "secret"))

	_, err := client.PostForm(context.Background(), "/messages", url.Values{"To": {"+5511"}})

	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, got.method)
	assert.Equal(t, "application/x-www-form-urlencoded", got.header.Get("Content-Type"))
	assert.Equal(t, "To=%2B5511", got.body)
	assert.Equal(t, "user", got.user)
	assert.Equal(t, "secret", got.password)
}

func TestClient_PostJSON(t *testing.T) {
	server, got := newServer(t, http.StatusOK, `{}`)

	_, err := New(WithBaseURL(server.URL)).PostJSON(context.Background(), "/events", map[string]string{"id": "1"})

	require.NoError(t, err)
	assert.Equal(t, "application/json", got.header.Get("Content-Type"))
	assert.JSONEq(t, `{"id":"1"}`, got.body)
}

func TestClient_ErrorStatusIsNotAnError(t *testing.T) {
	server, _ := newServer(t, http.StatusBadRequest, `{"code":21211}`)

	response, err := New(WithBaseURL(server.URL)).Get(context.Background(), "/")

	require.NoError(t, err)
	assert.False(t, response.IsSuccess())
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Equal(t, `{"code":21211}`, string(response.Body))
}

func TestClient_TransportError(t *testing.T) {
	server, _ := newServer(t, http.StatusOK, `{}`)
	server.Close()

	_, err := New(WithBaseURL(server.URL)).Get(context.Background(), "/")

	require.Error(t, err)
}

func TestClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	_, err := New(WithBaseURL(server.URL), WithTimeout(20*time.Millisecond)).Get(context.Background(), "/")

	require.Error(t, err)
}

func TestClient_RequestHeadersOverrideDefaults(t *testing.T) {
	server, got := newServer(t, http.StatusOK, `{}`)
	client := New(WithBaseURL(server.URL), WithHeader("Content-Type", "text/plain"))

	_, err := client.PostJSON(context.Background(), "/", struct{}{})

	require.NoError(t, err)
	assert.Equal(t, "application/json", got.header.Get("Content-Type"))
}

func TestResponse_DecodeJSONError(t *testing.T) {
	response := &Response{Body: []byte("not json")}

	require.Error(t, response.DecodeJSON(&struct{}{}))
}
