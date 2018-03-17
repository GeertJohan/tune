package main

import (
	"fmt"
	"os"
	"time"
)

// TODO: this whole panic to file was usefull for development, but should be removed and just print to stdout.
func panicToFile(output string) {
	panicFile, err := os.Create(fmt.Sprintf("tune-panic-%d", time.Now().Unix()))
	if err != nil {
		panic(output + `\n` + err.Error()) // seriously!? :p
	}
	defer panicFile.Close()
	panicFile.WriteString(output)
	fmt.Println("PANIC TO FILE!")
}
