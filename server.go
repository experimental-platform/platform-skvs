package main

import (
	"encoding/json"
	"errors"
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ResponseData struct {
	StatusCode  int      `json:"-"`
	Key         string   `json:"key"`
	IsNamespace bool     `json:"namespace"`
	Value       string   `json:"value"`
	Keys        []string `json:"keys,omitempty"`  // need better decision here
	Error       string   `json:"error,omitempty"` // need better decision here
}

var validKey = regexp.MustCompile(`^[a-zA-Z0-9_\-/:]+$`)

func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var responseData ResponseData
	key := r.URL.Path[1:]
	if validKey.MatchString(key) {
		key_path := expandPath(key)
		value := r.PostForm.Get("value")
		var keys []string
		var err error

		switch r.Method {
		case "GET":
			var entry Entry
			entry, err = readKey(key_path)
			if err == nil {
				if entry.isNamespace {
					keys = entry.data
				} else {
					value = entry.data[0]
				}
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
			responseData = ResponseData{StatusCode: http.StatusNotFound, Key: key, Error: err.Error()}
		}
	} else {
		responseData = ResponseData{StatusCode: http.StatusBadRequest, Key: key, Error: "Invalid key. Only " + validKey.String() + " allowed!"}
	}

	content, err := json.Marshal(responseData)
	if err == nil && responseData.StatusCode != 0 {
		callHooks(key, r.Method)
		w.WriteHeader(responseData.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.Write(append(content, '\n'))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		fmt.Println(err)
	}
}

func readKey(path string) (Entry, error) {
	var result []string
	var err error

	var fileInfo os.FileInfo
	fileInfo, err = os.Stat(path)
	isNamespace := false
	if err == nil {
		if fileInfo.IsDir() {
			isNamespace = true
			var files []os.FileInfo
			if files, err = ioutil.ReadDir(path); err == nil {
				for _, f := range files {
					result = append(result, f.Name())
				}
			}
		} else {
			var content []byte
			if content, err = ioutil.ReadFile(path); err == nil {
				result = append(result, string(content))
			}
		}
	}
	return Entry{data: result, isNamespace: isNamespace}, err
}

func putKey(path string, value string) error {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	return ioutil.WriteFile(path, []byte(value), os.ModePerm)
}

func deleteKey(path string) error {
	return os.Remove(path)
}

func expandPath(key string) string {
	return filepath.Join(opts.DataPath, key)
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

func callHooks(key, action string) {
	keyparts := strings.Split(key, "/")
	tmpkey := keyparts[0]
	callHook(tmpkey, action)
	for _, keypart := range keyparts[1:] {
		tmpkey = tmpkey + "/" + keypart
		callHook(tmpkey, action)
	}
}

// ignore errors, just print them and continue
func callHook(key, action string) {
	for _, hookUrl := range opts.WebHookUrls {
		go func(hookUrl string, hookData url.Values) {
			if _, err := http.PostForm(hookUrl, hookData); err != nil {
				fmt.Printf("WebHook Post failed: %s\n", err)
			} else {
				fmt.Printf("Called '%s' with Payload '%+v'.\n", hookUrl, hookData)
			}
		}(hookUrl, url.Values{"key": {key}, "action": {action}})
	}
}

var opts struct {
	DataPath    string   `short:"d" long:"data-path" default:"./data" description:"Directory where files will be stored."`
	Port        int      `short:"p" long:"port" default:"8080" description:"Port where server is listening for requests."`
	WebHookUrls []string `short:"w" long:"webhook-url" description:"WebHook-Urls."`
}

func main() {
	flags.Parse(&opts)
	opts.DataPath, _ = filepath.Abs(opts.DataPath)
	fmt.Println("DATA_PATH:", opts.DataPath)
	fmt.Println("PORT:", opts.Port)
	for i, hookUrl := range opts.WebHookUrls {
		if len(hookUrl) >= 4 && hookUrl[:4] != "http" {
			opts.WebHookUrls[i] = "http://" + hookUrl
		}
	}

	fmt.Printf("HOOKS: %+v\n", opts.WebHookUrls)

	deviceMux := http.NewServeMux()
	deviceMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		r.URL.Path = "/devices" + p
		handler(w, r)
	})
	go http.ListenAndServe(":82", deviceMux)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+strconv.Itoa(opts.Port), nil)
}
