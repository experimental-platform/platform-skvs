package main

import (
	"encoding/json"
	"errors"
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
	StatusCode  int      `json:"-"`
	Key         string   `json:"key"`
	IsNamespace bool     `json:"namespace"`
	Value       string   `json:"value,omitempty"`
	Keys        []string `json:"keys,omitempty"` // need better decision here
}

func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.URL.Path[1:]
	key_path := expandPath(key)
	value := r.PostForm.Get("value")
	var keys []string
	var err error
	var responseData ResponseData

	switch r.Method {
	case "GET":
		var values []string
		values, err = readKey(key_path)
		if len(values) == 1 {
			value = values[0]
		} else {
			keys = values
		}
	case "DELETE":
		err = deleteKey(key_path)
	case "PUT", "POST":
		err = putKey(key_path, value)
	}

	if err == nil {
		responseData = ResponseData{StatusCode: http.StatusOK, Key: key, Value: value, Keys: keys}
		if keys == nil {
			responseData.IsNamespace = false
		} else {
			responseData.IsNamespace = true
		}
	} else {
		responseData = ResponseData{StatusCode: http.StatusNotFound, Key: key}
	}

	content, err := json.Marshal(responseData)
	fmt.Println(string(content))
	if err == nil {
		w.WriteHeader(responseData.StatusCode)
		w.Write(append(content, '\n'))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		fmt.Println(err)
	}
}

func readKey(path string) ([]string, error) {
	var result []string
	var err error

	if err = isDirectory(path); err == nil {
		if files, err := ioutil.ReadDir(path); err == nil {
			for _, f := range files {
				result = append(result, f.Name())
			}
		}
	} else if err = fileExists(path); err == nil {
		if content, err := ioutil.ReadFile(path); err == nil {
			result = append(result, string(content))
		}
	}
	return result, err
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

// Return nil if File exists, else non-nil value
func fileExists(filename string) error {
	_, err := os.Stat(filename)
	return err
}

// Return nil if filename is a directory, else non-nil value
func isDirectory(filename string) error {
	stat, err := os.Stat(filename)
	if err == nil && !stat.IsDir() {
		return errors.New("'" + filename + "' is not a directory!")
	}
	return nil
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
