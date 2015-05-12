package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var DATA_PATH string
var PORT int

type ResponseData struct {
	StatusCode int    `json:"-"`
	Key        string `json:"key"`
	Value      string `json:"value"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.URL.Path[1:]
	key_path := expandPath(key)
	value := r.PostForm.Get("value")
	var err error
	var responseData ResponseData
	switch r.Method {
	case "GET":
		value, err = readKey(key_path)
	case "DELETE":
		err = deleteKey(key_path)
	case "PUT", "POST":
		err = putKey(key_path, value)
	}

	if err == nil {
		responseData = ResponseData{StatusCode: http.StatusOK, Key: key, Value: value}
	} else {
		responseData = ResponseData{StatusCode: http.StatusNotFound, Key: key}
	}

	content, err := json.Marshal(responseData)
	if err == nil {
		w.WriteHeader(responseData.StatusCode)
		w.Write(append(content, '\n'))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		fmt.Println(err)
	}
}

func readKey(path string) (string, error) {
	var content []byte
	var err error
	if err = fileExists(path); err == nil {
		content, err = ioutil.ReadFile(path)
	}
	return string(content), err
}

func putKey(path string, value string) error {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	return ioutil.WriteFile(path, []byte(value), os.ModePerm)
}

func deleteKey(path string) error {
	return os.Remove(path)
}

func expandPath(key string) string {
	return filepath.Join(DATA_PATH, key)
}

// Return nil if File exists, else non-nil
func fileExists(filename string) error {
	_, err := os.Stat(filename)
	return err
}

func main() {
	flag.StringVar(&DATA_PATH, "data-path", "./data", "Directory where files will be stored.")
	flag.IntVar(&PORT, "port", 8080, "Port where server is listening for requests")
	flag.Parse()
	DATA_PATH, _ = filepath.Abs(DATA_PATH)
	fmt.Println("DATA_PATH:", DATA_PATH)
	fmt.Println("PORT:", PORT)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+strconv.Itoa(PORT), nil)
}
