package cli

import (
	"flag"
	"os"
	"os/exec"

	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/packages/db"
	"github.com/Arthur-Conti/guh/packages/log/logger"
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
	fs.Parse(os.Args[2:])
	
	switch *dbName {
	case "Postgres":
		opts := db.PostgresOpts{
			User: *user,
			Password: *pass,
			IP: *ip,
			Port: *port,
			Database: *database,
		}
		if err := db.PostgresDockerCompose(opts); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error creating docker compose: %v\n", Vals: []any{err}})
		}
	}

	if *run {
		if err := RunCompose(); err != nil {
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
		return err
	}
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Docker Compose started successfully."})
	return nil
}