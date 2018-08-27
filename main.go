package main

import (
	"flag"
	"fmt"
	"github.com/Parquery/gocontracts/gocontracts"
	"os"
)

var inPlace = flag.Bool("w", false, "write result to (source) file instead of stdout")

func usage() {
	_, err := fmt.Fprintf(os.Stderr, "usage: gocontracts [flags] [path]\n")
	if err != nil {
		panic(err.Error())
	}

	flag.PrintDefaults()
}

func main() {
	os.Exit(func() (retcode int) {
		flag.Parse()

		if flag.NArg() != 1 {
			_, err := fmt.Fprintf(os.Stderr, "Expected the path to the file as a single positional argument, "+
				"but got positional %d argument(s)\n", flag.NArg())

			if err != nil {
				panic(err.Error())
			}

			usage()
			retcode = 1
			return
		}

		pth := flag.Arg(0)

		if *inPlace {
			err := gocontracts.ProcessInPlace(pth)
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, err.Error())
				if err != nil {
					panic(err.Error())
				}
				return 1
			}
		} else {
			updated, err := gocontracts.ProcessFile(pth)
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, err.Error())
				if err != nil {
					panic(err.Error())
				}
				return 1
			}

			_, err = fmt.Fprint(os.Stdout, updated)
			if err != nil {
				panic(err.Error())
			}
		}

		return 0
	}())
}
