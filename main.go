package main

import (
	"fmt"
	"os"
	"flag"
	"strings"
)

var (
	owner = flag.String("owner","","provide name of doc owner")
	paths = flag.String("paths","","provide comma separated args of paths --paths=/path1,/path2,/path3")
)

func main() {
	flag.Parse()
	if *paths == ""{
		fmt.Printf("--paths flag must be set and contain  comma-separated paths to api objects")
		os.Exit(1)
	}
	if *owner == ""{
		fmt.Printf("--owner must be set")
		os.Exit(1)
	}
	apiPaths := strings.Split(*paths,",")
	printAPIDocs(apiPaths,*owner)
}
