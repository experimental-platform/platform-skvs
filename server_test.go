package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

	entry, err := readKey(testPath)
	if err != nil {
		t.Error(err)
	}
	if entry.isNamespace {
		t.Error("Is namespaced value, but should not be.")
	}
	if len(entry.data) != 1 || entry.data[0] != testContent {
		t.Errorf("Too many results given (%+v) or first result has not expected content (%s).", entry.data, testContent)
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

	entry, err := readKey(testPathDirectory)
	if err != nil {
		t.Error(err)
	}
	if !entry.isNamespace {
		t.Error("Is not namespaced value, but should be.")
	}
	if len(entry.data) != 2 {
		t.Errorf("Too many/few results given (%+v).", entry.data)
	}
}

func TestHTTPGetKey(t *testing.T) {
	cleanData()
	// Key does not exists
	req, err := http.NewRequest("GET", "http://localhost/foobar", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected Code %i, got %i", http.StatusNotFound, w.Code)
	}

	testFilePath := expandPath("foobar")
	testContent := "foobar"
	if err := os.MkdirAll(filepath.Dir(testFilePath), os.ModePerm); err != nil {
		t.Errorf("Could not create directory '%s'\n", testFilePath)
	}
	if err := ioutil.WriteFile(testFilePath, []byte(testContent), os.ModePerm); err != nil {
		t.Errorf("Could not write file '%s' with content '%s' (%v)\n", testFilePath, testContent, err)
	}

	req, err = http.NewRequest("GET", "http://localhost/foobar", nil)
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected Code %i, got %i", http.StatusNotFound, w.Code)
	}
	expectedBody := "{\"key\":\"foobar\",\"namespace\":false,\"value\":\"foobar\"}\n"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected Body '%s', got '%s'", expectedBody, w.Body.String())
	}
}

func TestCacheUpdatedOnWrite(t *testing.T) {
	cleanData()
	testPath := expandPath("foo/bar/zero")
	testContent1 := "oldContent"
	testContent2 := "newContent"

	err := putKey(testPath, testContent1)
	if err != nil {
		t.Error(err)
	}

	entry, err := readKey(testPath)
	if err != nil {
		t.Error(err)
	}
	if entry.isNamespace {
		t.Error("Is namespaced value, but should not be.")
	}
	if len(entry.data) != 1 || entry.data[0] != testContent1 {
		t.Errorf("Too many results given (%+v) or first result has not expected content (%s).", entry.data, testContent1)
	}

	err = putKey(testPath, testContent2)
	if err != nil {
		t.Error(err)
	}

	entry, err = readKey(testPath)
	if err != nil {
		t.Error(err)
	}
	if entry.isNamespace {
		t.Error("Is namespaced value, but should not be.")
	}
	if len(entry.data) != 1 || entry.data[0] != testContent2 {
		t.Errorf("Too many results given (%+v) or first result has not expected content (%s).", entry.data, testContent2)
	}
}

func TestParentCacheUpdated(t *testing.T) {
	cleanData()
	testPath1 := expandPath("foo/bar/zero")
	testPath2 := expandPath("foo/bar/one")
	testPathParent := expandPath("foo/bar")
	testContent := "foobar"

	err := putKey(testPath1, testContent)
	if err != nil {
		t.Error(err)
	}

	entry, err := readKey(testPathParent)
	if err != nil {
		t.Error(err)
	}
	if !entry.isNamespace {
		t.Error("Is not a namespaced value, but should be.")
	}
	if len(entry.data) != 1 {
		t.Errorf("Should have 1 result, got %v.", len(entry.data))
	}

	err = putKey(testPath2, testContent)
	if err != nil {
		t.Error(err)
	}

	entry, err = readKey(testPathParent)
	if err != nil {
		t.Error(err)
	}
	if !entry.isNamespace {
		t.Error("Is not a namespaced value, but should be.")
	}
	if len(entry.data) != 2 {
		t.Errorf("Should have 2 results, got %v.", len(entry.data))
	}
}

func TestRootParentCacheUpdated(t *testing.T) {
	cleanData()
	testPath1 := expandPath("zero")
	testPath2 := expandPath("one")
	testPathParent := expandPath("/")
	testContent := "foobar"

	err := putKey(testPath1, testContent)
	if err != nil {
		t.Error(err)
	}

	entry, err := readKey(testPathParent)
	if err != nil {
		t.Error(err)
	}
	if !entry.isNamespace {
		t.Error("Is not a namespaced value, but should be.")
	}
	if len(entry.data) != 1 {
		t.Errorf("Should have 1 result, got %v.", len(entry.data))
	}

	err = putKey(testPath2, testContent)
	if err != nil {
		t.Error(err)
	}

	entry, err = readKey(testPathParent)
	if err != nil {
		t.Error(err)
	}
	if !entry.isNamespace {
		t.Error("Is not a namespaced value, but should be.")
	}
	if len(entry.data) != 2 {
		t.Errorf("Should have 2 results, got %v.", len(entry.data))
	}
}

func TestCacheUpdatedOnDelete(t *testing.T) {
	cleanData()
	testPath := expandPath("foo/bar/zero")
	testContent := "testContent"

	err := putKey(testPath, testContent)
	if err != nil {
		t.Error(err)
	}

	entry, err := readKey(testPath)
	if err != nil {
		t.Error(err)
	}
	if entry.isNamespace {
		t.Error("Is namespaced value, but should not be.")
	}
	if len(entry.data) != 1 || entry.data[0] != testContent {
		t.Errorf("Too many results given (%+v) or first result has not expected content (%s).", entry.data, testContent)
	}

	err = deleteKey(testPath)
	if err != nil {
		t.Error("Failed to remove key.")
	}

	entry, err = readKey(testPath)
	if err == nil {
		t.Errorf("Entry '%+v' at '%v' should not exist, but does.", entry, testPath)
	}
}

func cleanData() {
	os.RemoveAll(opts.DataPath)
	skvsCache = make(map[string]Entry)
}

func TestMain(m *testing.M) {
	opts.DataPath, _ = filepath.Abs("./data-test")
	exit := m.Run()
	cleanData()
	os.Exit(exit)
}
