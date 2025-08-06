package cli

import (
	"fmt"
	"os"
)

func Help() {
	fmt.Println(`guh - Go Universal Helper was made to help you with any anoying repetitive task in go

Usage:
  guh <command> [flags]

Available Commands:
  help               Show this help message
  compose            Create a docker-compose.yml file with the services you might need in your project (Databases, service, etc...)
  config             Create the config files for your project (logger, databases, init, etc...)
  mod                Help you start your project go mod, connect to github, create your go mod file and download GUH packages
  structure          Create the core structure to start your project

Flags:
  --help         Show help for any command

Examples:
  guh help
  guh structure --showFirst
  guh compose --debug
  guh mod --gin

For more information, visit: https://github.com/Arthur-Conti/guh`)
	os.Exit(0)
}