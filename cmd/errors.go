package commands

import (
	"fmt"
	"os"

	"errors"
)

var verbose = false

func SetVerbose() {
	verbose = true
}

func Winddown(message string, params ...interface{}) {
	MaybeWindDown(errors.New(fmt.Sprintf(message, params...)))
}

func MaybeWindDown(err error) {
	if err != nil {
		fmt.Printf("Error occured, aborting. %v\n\n", err)
		if verbose {
			panic(err)
		}
		os.Exit(1)
	}
}

func Assert(cond bool, message string, params ...interface{}) {
	if !cond {
		Winddown(message, params...)
	}
}

func Break(s interface{}) {

}
