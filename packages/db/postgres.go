package db

import (
	"context"
	"fmt"
	"os"

	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/logger/logger"
	"github.com/jackc/pgx/v5"
)

type PostgresOpts struct {
	User                  string
	Password              string
	Database              string
	IP                    string
	Port                  string
}

func DefaultPostgres() (*pgx.Conn, error) {
	opts := PostgresOpts{
		User: "user_test",
		Password: "pass_test",
		IP: "localhost",
		Port: "5432",
		Database: "default",
	}
	return postgresInit(postgresUri(opts))
} 

func Postgres(opts PostgresOpts) (*pgx.Conn, error) {
	return postgresInit(postgresUri(opts))
}

func postgresUri(opts PostgresOpts) string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v", opts.User, opts.Password, opts.IP, opts.Port, opts.Database)
}

func postgresInit(uri string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), uri)
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "DB", Message: "Unable to connect to database: %v\n", Vals: []any{err}})
		return nil, err
	}
	defer conn.Close(context.Background())
	err = conn.Ping(context.Background())
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "DB", Message: "Ping failed: %v\n", Vals: []any{err}})
		return nil, err
	}
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "DB", Message: "Connected to PostgreSQL successfully!"})
	return conn, nil
}

func PostgresDockerCompose(opts PostgresOpts) error {
	fileName := "docker-compose.yml"

	if _, err := os.Stat(fileName); err == nil {
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "DB", Message: "File '%s' already exists. Skipping creation.", Vals: []any{fileName}})
		return nil
	} else if !os.IsNotExist(err) {
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "DB", Message: "Error checking file: %v\n", Vals: []any{err}})	
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

	err := os.WriteFile("./" + fileName, []byte(content), 0644)
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "DB", Message: "Error writing file: %v\n", Vals: []any{err}})		
		return err
	}

	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "DB", Message: "docker-compose.yml generated successfully."})
	return nil
}
