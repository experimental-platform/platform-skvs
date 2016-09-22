package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/experimental-platform/platform-skvs/server"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	DataPath    string   `short:"d" long:"data-path" default:"./data" description:"Directory where files will be stored."`
	Port        int      `short:"p" long:"port" default:"8080" description:"Port where server is listening for requests."`
	WebHookUrls []string `short:"w" long:"webhook-url" description:"WebHook-Urls."`
	CacheExempt []string `short:"e" long:"exempt-from-cache" description:"Paths which shall not use cache."`
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

	fmt.Println("PATHS EXEMPT FROM CACHE:")
	for _, p := range opts.CacheExempt {
		fmt.Printf(" - %s\n", p)
	}

	fmt.Printf("HOOKS: %+v\n", opts.WebHookUrls)

	handler := server.NewServerHandler(opts.DataPath, opts.CacheExempt, opts.WebHookUrls)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+strconv.Itoa(opts.Port), nil)
}
