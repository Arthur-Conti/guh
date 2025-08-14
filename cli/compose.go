package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/libs/db"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
	projectconfig "github.com/Arthur-Conti/guh/libs/project_config"
	"gopkg.in/yaml.v3"
)

func Compose() error {
	fs := flag.NewFlagSet("compose", flag.ExitOnError)
	dbName := fs.String("dbName", "Postgres", "DB to create compose")
	addService := fs.Bool("addService", false, "Add the service at the docker compose")
	run := fs.Bool("run", false, "Run compose")
	help := fs.Bool("help", false, "Help with compose command")
	fs.Parse(os.Args[2:])

	if *help {
		HelpCompose()
	}

	switch *dbName {
	case "Postgres":
		if err := PostgresDockerCompose(*db.GetDefaultPostgresOpts()); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error creating docker compose: %v\n", Vals: []any{err}})
			return err
		}
	}

	if *addService {
		cfg, err := projectconfig.Load()
		if err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to load project config", err, errorhandler.WithOp("compose"))
		}
		if cfg.ServiceName == "" {
			return errorhandler.New(errorhandler.KindInvalidArgument, "To add the service to the compose you must provide a service name (run 'guh structure --serviceName=...' first)", errorhandler.WithOp("compose"))
		}
		if err := AddAppService("docker-compose.yml", cfg.ServiceName); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error adding service to docker compose: %v\n", Vals: []any{err}})
			return err
		}
	}

	if *run {
		if err := RunCompose(); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running docker compose: %v\n", Vals: []any{err}})
			return err
		}
	}

	return nil
}

func PostgresDockerCompose(opts db.PostgresOpts) error {
	fileName := "docker-compose.yml"

	if _, err := os.Stat(fileName); err == nil {
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "db", Message: "File '%s' already exists. Skipping creation.", Vals: []any{fileName}})
		return nil
	} else if !os.IsNotExist(err) {
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "db", Message: "Error checking file: %v\n", Vals: []any{err}})
		return nil
	}

	content := fmt.Sprintf(`version: '3.8'

services:
  postgres:
    image: postgres:15
    container_name: postgres_container
    restart: always
    environment:
      POSTGRES_USER: %[1]v
      POSTGRES_PASSWORD: %[2]v
      POSTGRES_DB: %[3]v
    ports:
      - "%[4]v:%[4]v"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
`, opts.User, opts.Password, opts.Database, opts.Port)

	err := os.WriteFile("./"+fileName, []byte(content), 0644)
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "db", Message: "Error writing file: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "error writing file", err, errorhandler.WithOp("compose.PostgresDockerCompose"))
	}

	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "db", Message: "docker-compose.yml generated successfully."})
	return nil
}

func AddAppService(composePath, serviceName string) error {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose map[string]interface{}
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return fmt.Errorf("failed to parse compose file: %w", err)
	}

	services, ok := compose["services"].(map[string]interface{})
	if !ok {
		services = make(map[string]interface{})
		compose["services"] = services
	}

	if _, exists := services[serviceName]; exists {
		config.Config.Logger.Warningf(logger.LogMessage{ApplicationPackage: "cli", Message: "Service already exists: %v", Vals: []any{serviceName}})
		return nil
	}

	services[serviceName] = map[string]interface{}{
		"build": map[string]interface{}{
			"context": ".",
		},
		"container_name": serviceName,
		"ports":          []string{"8080:8080"},
		"depends_on":     []string{"postgres"},
	}

	newData, err := yaml.Marshal(compose)
	if err != nil {
		return fmt.Errorf("failed to marshal updated compose file: %w", err)
	}

	if err := os.WriteFile(composePath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write updated compose file: %w", err)
	}

	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Added service: %v", Vals: []any{serviceName}})
	return nil
}

func RunCompose() error {
	cmd := exec.Command("docker", "compose", "up", "--build", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Running docker-compose up -d..."})
	if err := cmd.Run(); err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running docker-compose: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error running docker-compose", err, errorhandler.WithOp("compose.RunCompose"))
	}
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Docker Compose started successfully."})
	return nil
}

func HelpCompose() {
	fmt.Println(`compose - The compose command help you creating your docker compose files 
	
Usage:
  guh compose [flags]
	
Flags:
  --dbName         Define the kind of database you want into your docker compose (Defaults to Postgres)
  --run            Run the docker compose in background after creating it
  --addService     Add your service to the docker compose

Examples:
  guh compose --dbName=Postgres
  guh compose --run
  guh compose --addService

For more information, visit: https://github.com/Arthur-Conti/guh`)
	os.Exit(0)
}
