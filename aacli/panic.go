package main

import (
	"fmt"
	"os"
	"time"
)

func panicToFile(output string) {
	panicFile, err := os.Create(fmt.Sprintf("aacli-panic-%d", time.Now().Unix()))
	if err != nil {
		panic(err) // seriously!? :p
	}
	defer panicFile.Close()
	panicFile.WriteString(output)
	fmt.Println("PANIC TO FILE!")
}
