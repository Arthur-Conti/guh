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
)

func Compose() error {
	fs := flag.NewFlagSet("compose", flag.ExitOnError)
	dbName := fs.String("dbName", "Postgres", "DB to create compose")
	user := fs.String("user", "user_test", "DB user")
	pass := fs.String("pass", "pass_test", "DB password")
	ip := fs.String("ip", "localhost", "DB ip")
	port := fs.String("port", "5432", "DB port")
	database := fs.String("database", "default", "DB database name")
	run := fs.Bool("run", false, "Run compose")
	help := fs.Bool("help", false, "Help with compose command")
	fs.Parse(os.Args[2:])

	if *help {
		HelpCompose()
	}

	switch *dbName {
	case "Postgres":
		opts := db.PostgresOpts{
			User:     *user,
			Password: *pass,
			IP:       *ip,
			Port:     *port,
			Database: *database,
		}
		if err := db.PostgresDockerCompose(opts); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error creating docker compose: %v\n", Vals: []any{err}})
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

func RunCompose() error {
	cmd := exec.Command("docker", "compose", "up", "--build", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Running docker-compose up -d..."})
	if err := cmd.Run(); err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running docker-compose: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.InternalServerError, "Error running docker-compose", err)
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
  --user           Define the database user (Defaults to the default database credentials)
  --pass           Define the database password (Defaults to the default database credentials)
  --ip             Define the database ip (Defaults to the default database credentials)
  --port           Define the database port (Defaults to the default database credentials)
  --database       Define the database database (Defaults to the default database credentials)
  --run            Run the docker compose in background after creating it

Examples:
  guh compose --dbName=Postgres
  guh compose --user=test_user
  guh compose --run

For more information, visit: https://github.com/Arthur-Conti/guh`)
	os.Exit(0)
}