package main

import (
	"flag"
	"fmt"
)

var DATA_PATH string
var PORT int

func main() {
	flag.StringVar(&DATA_PATH, "data-path", "./data", "Directory where files will be stored.")
	flag.IntVar(&PORT, "port", 8080, "Port where server is listening for requests")
	flag.Parse()

	fmt.Println("DATA_PATH:", DATA_PATH)

	fmt.Println("PORT:", PORT)
}
