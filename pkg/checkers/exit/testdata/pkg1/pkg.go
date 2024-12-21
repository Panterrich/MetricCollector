package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello world")
	os.Exit(0) // want "direct calling os.Exit in main func"
}
