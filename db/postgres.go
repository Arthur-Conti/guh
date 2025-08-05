package db

import "database/sql"

type PostgresOpts struct {
	hasDB bool
	uri string
}

func Postgres(opts PostgresOpts) *sql.DB {
	
	return postgresInit()
}

func postgresInit() *sql.DB {
	conn, err := pgx.Connect(context.Background(), "postgres://username:password@localhost:5432/dbname")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatalf("Ping failed: %v\n", err)
	}

	fmt.Println("Connected to PostgreSQL successfully!")
}