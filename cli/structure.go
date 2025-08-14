package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Arthur-Conti/guh/config"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
	projectconfig "github.com/Arthur-Conti/guh/libs/project_config"
)

var noServiceName string = "no_service_name"

func Structure() error {
	fs := flag.NewFlagSet("structure", flag.ExitOnError)
	create := fs.Bool("create", false, "Create the project structure")
	showFirst := fs.Bool("showFirst", false, "Show the structure before creating it")
	serviceName := fs.String("serviceName", noServiceName, "The name of your service")
	help := fs.Bool("help", false, "Help with structure command")
	fs.Parse(os.Args[2:])

	if *help {
		HelpStructure()
	}

	if *serviceName == noServiceName {
		return errorhandler.New(errorhandler.BadRequest, "Service name must be passed")
	}

	cfg, err := projectconfig.Load()
	if err != nil {
		return errorhandler.Wrap(errorhandler.InternalServerError, "failed to load project config", err)
	}
	cfg.ServiceName = *serviceName
	if err := projectconfig.Save(cfg); err != nil {
		return errorhandler.Wrap(errorhandler.InternalServerError, "failed to save project config", err)
	}

	if *create {
		return createStructure(*serviceName)
	}
	if *showFirst {
		showStructure()
		fmt.Print("Continue? [Y/n]: ")
		var resp string
		fmt.Scanln(&resp)
		if resp != "" && strings.ToLower(resp) != "y" {
			config.Config.Logger.Warning(logger.LogMessage{ApplicationPackage: "cli", Message: "Aborted by user"})
			return nil
		}
		return createStructure(*serviceName)
	}
	return nil
}

func createStructure(serviceName string) error {
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

import (
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	routes.RouterRegister(server)
	server.Run(":8080")
}`,
		"./.env": `DB_USER: 'user_test'
DB_PASS: 'pass_test'
DB_IP: 'localhost'
DB_PORT: '5432'
DB_DATABASE: 'default'`,
		"./Dockerfile": `FROM golang:1.24.4-alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]`,
		"./internal/infra/http/routes/routes.go": fmt.Sprintf(`package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var serviceName = "%v"

func RouterRegister(server *gin.Engine) {

	server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Server Alive"})
	})
}

func groupName(name string) string {
	return serviceName + name
}`, serviceName),
	}
	for filePath, content := range fileMap {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error creating file %v: %v\n", Vals: []any{filePath, err}})
			return errorhandler.Wrap(errorhandler.InternalServerError, "error creating file "+filePath, err)
		}
	}
	return nil
}

func createDir(dirList []string, pathToCreate string) error {
	for _, dir := range dirList {
		if err := os.MkdirAll(filepath.Join(pathToCreate, dir), 0755); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error creating dir %v: %v\n", Vals: []any{dir, err}})
			return errorhandler.Wrap(errorhandler.InternalServerError, "error creating dir "+dir, err)
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
		"│       │   	 └── routes.go",
		"│       └── repositories/",
		"├── Dockerfile",
		"└── .env",
	}

	for _, line := range tree {
		fmt.Println(line)
	}
	fmt.Println()
}

func HelpStructure() {
	fmt.Println(`structure - The structure command help you creating the initial core structure for your project  

Usage:
  guh structure [flags]

Flags:
  --create         Creates your initial core structure for your project  
  --showFirst      Shows the structure that is gonna be created before it creates it
  --serviceName    (Required) The name of your service

Examples:
  guh structure --create --serviceName=test
  guh structure --showFirst --serviceName=test

For more information, visit: https://github.com/Arthur-Conti/guh`)
	os.Exit(0)
}
