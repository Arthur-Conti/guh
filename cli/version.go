package cli

import "fmt"

var Version = "dev"

func PrintVersion() {
	fmt.Println("GUH version:", Version)
}