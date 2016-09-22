package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type ResponseData struct {
	StatusCode  int      `json:"-"`
	Key         string   `json:"key"`
	IsNamespace bool     `json:"namespace"`
	Value       string   `json:"value"`
	Keys        []string `json:"keys,omitempty"`  // need better decision here
	Error       string   `json:"error,omitempty"` // need better decision here
}

type Entry struct {
	data        []string
	isNamespace bool
}

var skvsCache map[string]Entry = make(map[string]Entry)
var skvsCacheMutex sync.Mutex

var validKey = regexp.MustCompile(`^[a-zA-Z0-9_\-/:]+$`)

func NewServerHandler(dataPath string, cacheExempionList []string, webHookURLs []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		var responseData ResponseData
		key := r.URL.Path[1:]
		if validKey.MatchString(key) {
			key_path := filepath.Join(dataPath, key)
			exempt := isExemptFromCache(key, cacheExempionList)
			value := r.PostForm.Get("value")
			var keys []string
			var err error

			switch r.Method {
			case "GET":
				var entry Entry
				entry, err = readKey(key_path, exempt)
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
				err = putKey(key_path, exempt, value)
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
			callHooks(key, r.Method, nil)
			w.WriteHeader(responseData.StatusCode)
			w.Header().Set("Content-Type", "application/json")
			w.Write(append(content, '\n'))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			fmt.Println(err)
		}
	}
}

func readKey(path string, exemptFromCache bool) (Entry, error) {
	// return from cache if available
	if cached, ok := skvsCache[path]; ok {
		return cached, nil
	}

	// otherwise read from FS
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

	entry := Entry{data: result, isNamespace: isNamespace}
	if err == nil && !exemptFromCache {
		// store in cache for future reads
		skvsCacheMutex.Lock()
		defer skvsCacheMutex.Unlock()
		skvsCache[path] = entry
	}

	return entry, err
}

func putKey(path string, exemptFromCache bool, value string) error {
	// if cache already contains identical data, then do nothing
	if v, ok := skvsCache[path]; ok && !exemptFromCache && len(v.data) == 1 && v.data[0] == value {
		return nil
	}

	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	err := ioutil.WriteFile(path, []byte(value), os.ModePerm)
	if err != nil {
		return err
	}

	if !exemptFromCache {
		// update cache
		skvsCacheMutex.Lock()
		defer skvsCacheMutex.Unlock()
		invalidateCache(path)

		skvsCache[path] = Entry{data: []string{value}, isNamespace: false}
	}

	return nil
}

func deleteKey(path string) error {
	skvsCacheMutex.Lock()
	defer skvsCacheMutex.Unlock()
	invalidateCache(path)
	return os.RemoveAll(path)
}

// Return nil if File exists, else non-nil value
func fileExists(filename string) error {
	_, err := os.Stat(filename)
	return err
}

// removes cache entries for the given path and all its parents and/or children
func invalidateCache(path string) {
	key, _ := readKey(path, true)
	if key.isNamespace {
		for _, child := range key.data {
			invalidateCache(path + "/" + child)
		}
	}

	for {
		delete(skvsCache, path)
		if path == "/" {
			break
		} else {
			path = filepath.Dir(path)
		}
	}
}

// Return nil if filename is a directory, else non-nil value
func isDirectory(filename string) error {
	stat, err := os.Stat(filename)
	if err == nil && !stat.IsDir() {
		return errors.New("'" + filename + "' is not a directory!")
	}
	return nil
}

func callHooks(key, action string, webHookURLs []string) {
	keyparts := strings.Split(key, "/")
	tmpkey := keyparts[0]
	callHook(tmpkey, action, webHookURLs)
	for _, keypart := range keyparts[1:] {
		tmpkey = tmpkey + "/" + keypart
		callHook(tmpkey, action, webHookURLs)
	}
}

// ignore errors, just print them and continue
func callHook(key, action string, webHookURLs []string) {
	if webHookURLs == nil {
		return
	}

	for _, hookUrl := range webHookURLs {
		go func(hookUrl string, hookData url.Values) {
			if _, err := http.PostForm(hookUrl, hookData); err != nil {
				fmt.Printf("WebHook Post failed: %s\n", err)
			} else {
				fmt.Printf("Called '%s' with Payload '%+v'.\n", hookUrl, hookData)
			}
		}(hookUrl, url.Values{"key": {key}, "action": {action}})
	}
}

func isExemptFromCache(path string, exemptionList []string) bool {
	if exemptionList == nil {
		return false
	}

	for _, exempt := range exemptionList {
		if exempt == path {
			return true
		}
	}

	return false
}
