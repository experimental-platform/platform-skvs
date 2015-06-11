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

func TestDeleteKey(t *testing.T) {
	path := expandPath("foobar")
	content := "foobar"
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		t.Errorf("Could not create key '%s'\n", path)
	}
	if err := ioutil.WriteFile(path, []byte(content), os.ModePerm); err != nil {
		t.Errorf("Could not write file '%s' with content '%s'\n", path, content)
	}
	deleteKey(path)
	if _, err := os.Stat(path); err == nil || !os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestMain(m *testing.M) {
	opts.DataPath, _ = filepath.Abs("./data-test")
	exit := m.Run()
	os.RemoveAll(opts.DataPath)
	os.Exit(exit)
}
