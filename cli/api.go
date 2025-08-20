package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/Arthur-Conti/guh/config"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
	projectconfig "github.com/Arthur-Conti/guh/libs/project_config"
)

func Api() error {
	fs := flag.NewFlagSet("api", flag.ExitOnError)
	serve := fs.Bool("serve", false, "Serve your application")
	background := fs.Bool("bg", false, "If true runs the server in backgroud")
	kill := fs.Bool("kill", false, "Kill your application")
	newRoute := fs.String("newRoute", "", "New route to add to the application")
	get := fs.String("get", "", "Run a http request to the application")
	help := fs.Bool("help", false, "Help with serve command")
	fs.Parse(os.Args[2:])

	if *serve && *kill {
		return errorhandler.New(errorhandler.KindInvalidArgument, "You cant kill and serve at the same time", errorhandler.WithOp("api"))
	}

	if *help {
		HelpApi()
	}

	if *serve {
		return serveApi(*background)
	}

	if *kill {
		return killAPI()
	}

	if *newRoute != "" {
		return addRoute(*newRoute)
	}

	if *get != "" {
		return getRequest(*get)
	}

	return nil
}

func serveApi(background bool) error {
	if background {
		cmd := exec.Command("docker", "compose", "up", "--build", "-d")
		if _, err := cmd.CombinedOutput(); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "api", Message: "Error serving your application: %v\n", Vals: []any{err}})
			return errorhandler.Wrap(errorhandler.KindInternal, "Error serving your application", err, errorhandler.WithOp("api.serveApi"))
		}
		return nil
	}

	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "api", Message: "Starting docker compose (streaming logs)..."})
	cmd := exec.Command("docker", "compose", "up", "--build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "api", Message: "Error serving your application: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error serving your application", err, errorhandler.WithOp("api.serveApi"))
	}
	return nil
}

func killAPI() error {
	cmd := exec.Command("docker", "compose", "down")
	_, err := cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "api", Message: "Error serving your application: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error serving your application", err, errorhandler.WithOp("api.killAPI"))
	}
	return nil
}

func getRequest(request string) error {
	pc, err := projectconfig.Load()
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Error loading project config", err, errorhandler.WithOp("api.getRequest"))
	}
	cmd := exec.Command("curl", "-s", pc.BaseUrl+request)
	output, err := cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "api", Message: "Error sending request to the application: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error sending request to the application", err, errorhandler.WithOp("api.getRequest"))
	}
	fmt.Println(string(output))
	return nil
}

func addRoute(route string) error {
	// Derive identifier (as typed, cleaned) and path slug (snake_case)
	raw := strings.TrimSpace(route)
	raw = strings.TrimPrefix(raw, "/")
	if raw == "" {
		return errorhandler.New(errorhandler.KindInvalidArgument, "invalid route name", errorhandler.WithOp("api.addRoute"))
	}

	// Keep user casing for identifier; strip invalid chars and ensure valid Go ident
	toIdent := func(s string) string {
		var out []rune
		for i, r := range s {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				if i == 0 && unicode.IsDigit(r) {
					out = append(out, '_', r)
				} else {
					out = append(out, r)
				}
			}
		}
		if len(out) == 0 {
			return ""
		}
		return string(out)
	}
	// Convert camelCase/PascalCase or mixed to snake_case for the URL path
	toSnake := func(s string) string {
		if s == "" {
			return s
		}
		var out []rune
		var prevLowerOrDigit bool
		for _, r := range s {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				isUpper := unicode.IsUpper(r)
				if isUpper && prevLowerOrDigit {
					out = append(out, '_')
				}
				out = append(out, unicode.ToLower(r))
				prevLowerOrDigit = !isUpper || unicode.IsDigit(r)
			} else {
				if len(out) > 0 && out[len(out)-1] != '_' {
					out = append(out, '_')
				}
				prevLowerOrDigit = false
			}
		}
		// trim trailing underscore
		if len(out) > 0 && out[len(out)-1] == '_' {
			out = out[:len(out)-1]
		}
		return string(out)
	}

	ident := toIdent(raw)
	if ident == "" {
		return errorhandler.New(errorhandler.KindInvalidArgument, "invalid route name", errorhandler.WithOp("api.addRoute"))
	}
	slug := toSnake(raw)

	// 1) Ensure per-route file exists with a basic stub (use snake_case filename)
	routeFilePath := fmt.Sprintf("./internal/infra/http/routes/%s.go", slug)
	if _, err := os.Stat(routeFilePath); os.IsNotExist(err) {
		routesFile, err := os.Create(routeFilePath)
		if err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "Error creating routes file", err, errorhandler.WithOp("api.addRoute"))
		}
		defer routesFile.Close()
		if _, err := routesFile.WriteString(fmt.Sprintf(`package routes

import (
	"github.com/gin-gonic/gin"
)

func %sRoutes(group *gin.RouterGroup) {

}
`, ident)); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "Error writing routes file", err, errorhandler.WithOp("api.addRoute"))
		}
	}

	// 2) Complement main routes.go by registering the new route group
	mainRoutesPath := "./internal/infra/http/routes/routes.go"
	contentBytes, err := os.ReadFile(mainRoutesPath)
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Error reading routes.go", err, errorhandler.WithOp("api.addRoute"))
	}
	content := string(contentBytes)

	// If already registered, do nothing
	if strings.Contains(content, fmt.Sprintf("%sRoutes(", ident)) {
		config.Config.Logger.Warningf(logger.LogMessage{ApplicationPackage: "api", Message: "Route already registered: %v", Vals: []any{ident}})
		return nil
	}

	insertBlock := fmt.Sprintf("\n\t%[1]sRouter := server.Group(groupName(\"/%[2]s\"))\n\t%[1]sRoutes(%[1]sRouter)\n", ident, slug)

	// Find closing brace of RouterRegister and insert before it
	funcStart := strings.Index(content, "func RouterRegister(")
	if funcStart == -1 {
		return errorhandler.New(errorhandler.KindInternal, "Could not find RouterRegister in routes.go", errorhandler.WithOp("api.addRoute"))
	}
	braceOpen := strings.Index(content[funcStart:], "{")
	if braceOpen == -1 {
		return errorhandler.New(errorhandler.KindInternal, "Malformed RouterRegister: missing '{'", errorhandler.WithOp("api.addRoute"))
	}
	openIdx := funcStart + braceOpen
	depth := 0
	closeIdx := -1
scanLoop:
	for i := openIdx; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				closeIdx = i
				break scanLoop
			}
		}
	}
	if closeIdx == -1 {
		return errorhandler.New(errorhandler.KindInternal, "Malformed RouterRegister: missing '}'", errorhandler.WithOp("api.addRoute"))
	}

	newContent := content[:closeIdx] + insertBlock + content[closeIdx:]
	if err := os.WriteFile(mainRoutesPath, []byte(newContent), 0644); err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Error updating routes.go", err, errorhandler.WithOp("api.addRoute"))
	}

	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "api", Message: "Added route: %v (%v)", Vals: []any{ident, slug}})
	return nil
}

func HelpApi() {
	fmt.Println(`api - The api command help you handle your application, serving it and more 

Usage:
  guh mod [flags]

Flags:
  --serve     Serve your service base on your docker compose file
  --bg        Make your server run in background
  --kill      Kill your service 
  --newRoute  Add a new route to the application
  --get       Run http request to the application and output the response as json

Examples:
  guh api --serve
  guh api --kill
  guh api --newRoute=/order

For more information, visit: https://github.com/Arthur-Conti/guh`)
	os.Exit(0)
}
