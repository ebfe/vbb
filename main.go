package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

var debug = flag.Bool("d", false, "debug")

func do() {
	rsp, err := http.Get("https://build.voidlinux.eu/json/builders")
	if err != nil {
		fmt.Fprintf(os.Stderr, "vbb: get: %s\n", err)
		return
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "vbb: get: %d - %s\n", rsp.StatusCode, rsp.Status)
		return
	}

	var r io.Reader
	if *debug {
		r = io.TeeReader(rsp.Body, os.Stderr)
	} else {
		r = rsp.Body
	}

	var doc map[string]struct {
		State   string `json:"state"`
		Pending int    `json:"pendingBuilds"`
	}

	d := json.NewDecoder(r)
	if err := d.Decode(&doc); err != nil {
		fmt.Fprintf(os.Stderr, "vbb: json: %s\n", err)
		return
	}

	var builders = make([]string, 0, len(doc))
	for builder := range doc {
		builders = append(builders, builder)
	}
	sort.Strings(builders)

	for _, builder := range builders {
		info := doc[builder]
		builder = strings.Replace(builder, "_builder", "", 1)
		fmt.Printf("%12s: %9s", builder, info.State)
		if info.Pending != 0 {
			fmt.Printf(" %3d pending", info.Pending)
		}
		fmt.Println()
	}
}

func main() {
	flag.Parse()
	do()
}
