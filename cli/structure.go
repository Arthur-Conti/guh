package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Arthur-Conti/guh/config"
	errorhandler "github.com/Arthur-Conti/guh/packages/error_handler"
	"github.com/Arthur-Conti/guh/packages/log/logger"
)

func Structure() error {
	fs := flag.NewFlagSet("structure", flag.ExitOnError)
	create := fs.Bool("create", false, "Create the project structure")
	showFirst := fs.Bool("showFirst", false, "Show the structure before creating it")
	fs.Parse(os.Args[2:])

	if *create {
		return createStructure()
	}
	if *showFirst {
		showStructure()
		fmt.Print("Continue? [Y/n]: ")
		var resp string
		fmt.Scanln(&resp)
		if resp != "" && strings.ToLower(resp) != "y" {
			fmt.Println("❌ Aborted.")
			return nil
		}
		return createStructure()
	}
	return nil
}

func createStructure() error {
	mainDirList := []string{"cmd", "internal"}
	if err := createDir(mainDirList, "./"); err != nil {
		return err
	}
	internalDirList := []string{"domain", "config", "application", "infra"}
	if err := createDir(internalDirList, "./internal/"); err != nil {
		return err
	}
	applicationDirList := []string{"services"}
	if err := createDir(applicationDirList, "./internal/application/"); err != nil {
		return err
	}
	infraDirList := []string{"http", "repositories"}
	if err := createDir(infraDirList, "./internal/infra/"); err != nil {
		return err
	}
	httpDirList := []string{"controllers", "routes"}
	if err := createDir(httpDirList, "./internal/infra/http/"); err != nil {
		return err
	}
	fileMap := map[string]string{
		"./cmd/main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello from GUH!")
}`,
		"./.env": `'DB_USER': 'user_test'
'DB_PASS': 'pass_test'
'DB_IP': 'localhost'
'DB_PORT': '5432'
'DB_DATABASE': 'default'
		`,
	}
	for filePath, content := range fileMap {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error creating file %v: %v\n", Vals: []any{filePath, err}})
			return errorhandler.Wrap("InternalServerError", "error creating file "+filePath, err)
		}
	}
	return nil
}

func createDir(dirList []string, pathToCreate string) error {
	for _, dir := range dirList {
		if err := os.MkdirAll(filepath.Join(pathToCreate, dir), 0755); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error creating dir %v: %v\n", Vals: []any{dir, err}})
			return errorhandler.Wrap("InternalServerError", "error creating dir "+dir, err)
		}
	}

	return nil
}

func showStructure() {
	fmt.Println("Project structure to be created:")

	tree := []string{
		"├── cmd/",
		"│   └── main.go",
		"├── internal/",
		"│   ├── domain/",
		"│   ├── config/",
		"│   ├── application/",
		"│   │   └── services/",
		"│   └── infra/",
		"│       ├── http/",
		"│       │   ├── controllers/",
		"│       │   └── routes/",
		"│       └── repositories/",
		"└── .env",
	}

	for _, line := range tree {
		fmt.Println(line)
	}
	fmt.Println()
}
