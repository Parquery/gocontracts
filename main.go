package main

import (
	"flag"
	"fmt"
	"github.com/Parquery/gocontracts/gocontracts"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: gocontracts [flags] [path]\n")
	flag.PrintDefaults()
}

func main() {
	os.Exit(func() (retcode int) {
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

		if *inPlace {
			err := gocontracts.ProcessInPlace(pth)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				return 1
			}
		} else {
			updated, err := gocontracts.ProcessFile(pth)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				return 1
			}

			fmt.Fprint(os.Stdout, updated)
		}

		return 0
	}())
}
