package main

import (
	"bitbucket.org/GeertJohan/audioaddict/aaapi"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"os"
)

func main() {
	account, err := aaapi.NetworkDI.Authenticate("gjr19912@gmail.com", "retentiet3st")
	// account, err := aaapi.NetworkDI.Authenticate("gjr19912+devtest@gmail.com", "testtesttest")
	if err != nil {
		fmt.Printf("error authenticating: %v\n", err)
		os.Exit(1)
	}
	spew.Dump(account)
}
