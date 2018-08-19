package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Parquery/gocontracts/gocontracts"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: gocontracts [flags] [path]\n")
	flag.PrintDefaults()
}

func run() (retcode int) {
	inPlace := flag.Bool("w", false, "write result to (source) file instead of stdout")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Expected the path to the file as a single positional argument, "+
			"but got positional %d argument(s)\n", flag.NArg())
		usage()
		retcode = 1
		return
	}

	pth := flag.Arg(0)

	data, err := ioutil.ReadFile(pth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read: %s\n", err)
		return 1
	}

	text := string(data)

	updated, err := gocontracts.Process(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to process %s: %s\n", pth, err)
		return 1
	}

	if *inPlace {
		var fi os.FileInfo
		fi, err = os.Stat(pth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to stat %s: %s\n", pth, err)
			return 1
		}

		ioutil.WriteFile(pth, []byte(updated), fi.Mode())
	} else {
		fmt.Fprint(os.Stdout, updated)
		if !strings.HasSuffix(updated, "\n") {
			fmt.Fprint(os.Stdout, "\n")
		}
	}

	return 0
}

func main() {
	os.Exit(run())
}
