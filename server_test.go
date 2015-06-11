package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	expected := opts.DataPath + "/foobar"
	actual := expandPath("foobar")
	if expected != actual {
		t.Errorf("Expected: '%s' Got: '%s'\n", expected, actual)
	}
}

func TestPutKeyWithoutNamespace(t *testing.T) {
	testPath := expandPath("foobar")
	testContent := "foobar"
	err := putKey(testPath, testContent)
	if err != nil {
		t.Fail()
	}
	if content, err := ioutil.ReadFile(testPath); err != nil || string(content) != testContent {
		t.Fail()
	}
}

func TestPutKeyWithNamespace(t *testing.T) {
	testPathDirectory := expandPath("foo/")
	testPathFile := expandPath("foo/bar")
	testContent := "foobar"
	err := putKey(testPathFile, testContent)
	if err != nil {
		t.Fail()
	}
	if fileInfo, err := os.Stat(testPathDirectory); err != nil || !fileInfo.IsDir() {
		t.Fail()
	}
	if content, err := ioutil.ReadFile(testPathFile); err != nil || string(content) != testContent {
		t.Fail()
	}
}

func TestMain(m *testing.M) {
	opts.DataPath, _ = filepath.Abs("./data-test")
	exit := m.Run()
	os.RemoveAll(opts.DataPath)
	os.Exit(exit)
}
