package client

import (
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/experimental-platform/platform-skvs/server"
	"github.com/stretchr/testify/assert"
)

func TestGetNonexistent(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	defer os.RemoveAll(tmpdir)
	srv := httptest.NewServer(server.NewServerHandler(tmpdir, nil, nil))
	c := NewFromURL(srv.URL)

	_, err = c.Get("foobar")

	assert.NotNil(t, err)
}

func TestSetGet(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	defer os.RemoveAll(tmpdir)
	srv := httptest.NewServer(server.NewServerHandler(tmpdir, nil, nil))
	c := NewFromURL(srv.URL)

	testContent := `This
	is a test
	data string
	for SKVS`

	err = c.Set("foobar", testContent)
	assert.Nil(t, err)

	rcvdContent, err := c.Get("foobar")
	assert.Nil(t, err)
	assert.Equal(t, rcvdContent, testContent)

	_, err = c.Get("other/foobar")
	assert.NotNil(t, err)
}

func TestSetDelete(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	defer os.RemoveAll(tmpdir)
	srv := httptest.NewServer(server.NewServerHandler(tmpdir, nil, nil))
	c := NewFromURL(srv.URL)

	testContent := `This
	is a test
	data string
	for SKVS`

	err = c.Set("foobar", testContent)
	assert.Nil(t, err)

	rcvdContent, err := c.Get("foobar")
	assert.Nil(t, err)
	assert.Equal(t, rcvdContent, testContent)

	err = c.Delete("foobar")
	assert.Nil(t, err)

	_, err = c.Get("foobar")
	assert.NotNil(t, err)
}
