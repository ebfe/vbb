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

func lastBuildStatus(builder string) (string, error) {
	for i := -1; i > -3; i-- {
		u := fmt.Sprintf("https://build.voidlinux.eu/json/builders/%s/builds/%d", builder, i)
		rsp, err := http.Get(u)
		if err != nil {
			return "", err
		}
		defer rsp.Body.Close()

		if rsp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("http status: %d - %s\n", rsp.StatusCode, rsp.Status)
		}

		var doc struct {
			Text []string `json:"text"`
		}
		d := json.NewDecoder(rsp.Body)
		err = d.Decode(&doc)
		if err != nil {
			return "", err
		}
		return strings.Join(doc.Text, " "), nil
	}
	return "", nil
}

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
		Current []int  `json:"currentBuilds"`
		Cached  []int  `json:"cachedBuilds"`
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
		arch := strings.Replace(builder, "_builder", "", 1)
		fmt.Printf("%12s: %9s", arch, info.State)
		if info.Pending != 0 {
			fmt.Printf(" %3d pending", info.Pending)
		}
		last, err := lastBuildStatus(builder)
		if err != nil {
		}
		if last != "" {
			fmt.Printf(" | %s", last)
		}
		fmt.Println()
	}
}

func main() {
	flag.Parse()
	do()
}
