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
	cleanData()
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
	cleanData()
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
	cleanData()
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

func TestReadKeyWithoutNamespace(t *testing.T) {
	cleanData()
	testPath := expandPath("foobar")
	testContent := "foobar"
	if err := os.MkdirAll(filepath.Dir(testPath), os.ModePerm); err != nil {
		t.Errorf("Could not create key '%s'\n", testPath)
	}
	if err := ioutil.WriteFile(testPath, []byte(testContent), os.ModePerm); err != nil {
		t.Errorf("Could not write file '%s' with content '%s'\n", testPath, testContent)
	}

	results, isNamespace, err := readKey(testPath)
	if err != nil {
		t.Error(err)
	}
	if isNamespace {
		t.Error("Is namespaced value, but should not be.")
	}
	if len(results) != 1 || results[0] != testContent {
		t.Errorf("Too many results given (%+v) or first result has not expected content (%s).", results, testContent)
	}
}

func TestReadKeyWithNamespace(t *testing.T) {
	cleanData()
	testPathDirectory := expandPath("foo/")
	testPathFile1 := expandPath("foo/bar1")
	testPathFile2 := expandPath("foo/bar2")
	testContent := "foobar"
	if err := os.MkdirAll(testPathDirectory, os.ModePerm); err != nil {
		t.Errorf("Could not create directory '%s'\n", testPathDirectory)
	}
	if err := ioutil.WriteFile(testPathFile1, []byte(testContent), os.ModePerm); err != nil {
		t.Errorf("Could not write file '%s' with content '%s'\n", testPathFile1, testContent)
	}
	if err := ioutil.WriteFile(testPathFile2, []byte(testContent), os.ModePerm); err != nil {
		t.Errorf("Could not write file '%s' with content '%s'\n", testPathFile2, testContent)
	}

	results, isNamespace, err := readKey(testPathDirectory)
	if err != nil {
		t.Error(err)
	}
	if !isNamespace {
		t.Error("Is not namespaced value, but should be.")
	}
	if len(results) != 2 {
		t.Errorf("Too many/few results given (%+v).", results)
	}
}

func cleanData() {
	os.RemoveAll(opts.DataPath)
}

func TestMain(m *testing.M) {
	opts.DataPath, _ = filepath.Abs("./data-test")
	exit := m.Run()
	cleanData()
	os.Exit(exit)
}
