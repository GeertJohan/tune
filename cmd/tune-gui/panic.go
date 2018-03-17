package main

import (
	"fmt"
	"os"
	"time"
)

func panicToFile(output string) {
	panicFile, err := os.Create(fmt.Sprintf("aagui-panic-%d", time.Now().Unix()))
	if err != nil {
		panic(err) // seriously!? Looks like we're out of luck today..
	}
	defer panicFile.Close()
	panicFile.WriteString(output)
	fmt.Println("PANIC TO FILE!")
}
