package cli

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/libs/db"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
)

const defaultMigrationsDir = "./internal/infra/db/migrations"
const defaultSeedsDir = "./internal/infra/db/seeds"

type migrationPair struct {
	Version  string
	Name     string
	UpPath   string
	DownPath string
}

// Db handles database migration related commands
func Db() error {
	fs := flag.NewFlagSet("db", flag.ExitOnError)
	migrationsDir := fs.String("dir", defaultMigrationsDir, "Directory for migration files")
	initFlag := fs.Bool("init", false, "Initialize migrations (dir and schema_migrations table)")
	newName := fs.String("new", "", "Create a new migration with the given snake_case name")
	up := fs.Bool("up", false, "Apply all pending migrations")
	down := fs.Bool("down", false, "Revert the last N applied migrations (use --steps)")
	steps := fs.Int("steps", 1, "Number of steps for --down")
	status := fs.Bool("status", false, "Show migration status")
	seedDir := fs.String("seedDir", defaultSeedsDir, "Directory for seed files")
	seedInit := fs.Bool("initSeeds", false, "Initialize seeds (dir and schema_seeds table)")
	seedNew := fs.String("newSeed", "", "Create a new seed with the given snake_case name")
	seedApply := fs.Bool("seed", false, "Apply all pending seeds")
	seedStatus := fs.Bool("seedStatus", false, "Show seed status")
	help := fs.Bool("help", false, "Show help for db command")
	fs.Parse(os.Args[2:])

	if *help {
		HelpDb()
	}

	// Determine action
	actions := 0
	if *initFlag {
		actions++
	}
	if *newName != "" {
		actions++
	}
	if *up {
		actions++
	}
	if *down {
		actions++
	}
	if *status {
		actions++
	}
	if *seedInit {
		actions++
	}
	if *seedNew != "" {
		actions++
	}
	if *seedApply {
		actions++
	}
	if *seedStatus {
		actions++
	}
	if actions == 0 {
		return errorhandler.New(errorhandler.KindInvalidArgument, "no action provided (use one of --init, --new, --up, --down, --status)", errorhandler.WithOp("db"))
	}
	if actions > 1 {
		return errorhandler.New(errorhandler.KindInvalidArgument, "multiple actions provided; please use only one of --init, --new, --up, --down, --status", errorhandler.WithOp("db"))
	}

	// Ensure directory exists for actions that need it
	if *initFlag {
		return initMigrationsAndSeeds(*migrationsDir, *seedDir)
	}

	if *newName != "" {
		if err := ensureDir(*migrationsDir); err != nil {
			return err
		}
		return createNewMigration(*migrationsDir, *newName)
	}

	if *seedInit {
		return initSeeds(*seedDir)
	}
	if *seedNew != "" {
		if err := ensureDir(*seedDir); err != nil {
			return err
		}
		return createNewSeed(*seedDir, *seedNew)
	}

	// For DB-connected actions, open connection first
	p, err := db.DefaultPostgres()
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to connect database", err, errorhandler.WithOp("db"))
	}
	defer p.Close()

	if *up {
		return migrateUp(p, *migrationsDir)
	}
	if *down {
		return migrateDown(p, *migrationsDir, *steps)
	}
	if *status {
		return migrationsStatus(p, *migrationsDir)
	}
	if *seedApply {
		return applySeeds(p, *seedDir)
	}
	if *seedStatus {
		return seedsStatus(p, *seedDir)
	}

	return nil
}

func HelpDb() {
	fmt.Println(`db - Database migration helper

Usage:
  guh db [flags]

Flags:
  --dir          Directory for migration files (default: ./internal/infra/db/migrations)
  --init         Initialize migrations directory and schema_migrations table
  --new          Create a new migration (pairs .up.sql and .down.sql)
  --up           Apply all pending migrations
  --down         Revert the last N migrations (use --steps)
  --steps        Number of steps for --down (default: 1)
  --status       Show migration status
  --seedDir      Directory for seed files (default: ./internal/infra/db/seeds)
  --initSeeds    Initialize seeds directory and schema_seeds table
  --newSeed      Create a new seed file
  --seed         Apply all pending seeds
  --seedStatus   Show seed status
  --help         Show help

Examples:
  guh db --init
  guh db --new=create_users_table
  guh db --up
  guh db --down --steps=1
  guh db --status
  guh db --initSeeds
  guh db --newSeed=seed_users
  guh db --seed
  guh db --seedStatus

For more information, visit: https://github.com/Arthur-Conti/guh`)
	os.Exit(0)
}

func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error creating dir %v: %v", Vals: []any{dir, err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to create migrations directory", err, errorhandler.WithOp("cli.db.ensureDir"))
	}
	return nil
}

func initMigrationsAndSeeds(migrationsDir, seedsDir string) error {
	if err := ensureDir(migrationsDir); err != nil {
		return err
	}
	p, err := db.DefaultPostgres()
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to connect database", err, errorhandler.WithOp("cli.db.initMigrations"))
	}
	defer p.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(64) PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TIMESTAMP NOT NULL
	)`
	if _, err := p.Conn.Exec(createTableSQL); err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error creating schema_migrations: %v", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to ensure schema_migrations table", err, errorhandler.WithOp("cli.db.initMigrations"))
	}
	if err := initSeeds(seedsDir); err != nil {
		return err
	}
	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Migrations initialized at %s", Vals: []any{migrationsDir}})
	return nil
}

func initMigrations(dir string) error { // kept for backward internal calls if any
	return initMigrationsAndSeeds(dir, defaultSeedsDir)
}

func createNewMigration(dir, name string) error {
	if strings.TrimSpace(name) == "" {
		return errorhandler.New(errorhandler.KindInvalidArgument, "migration name cannot be empty", errorhandler.WithOp("cli.db.createNewMigration"))
	}
	safe := sanitizeName(name)
	timestamp := time.Now().UTC().Format("20060102150405")
	upPath := filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", timestamp, safe))
	downPath := filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", timestamp, safe))

	upTpl := "-- up migration for %s\n\n"
	downTpl := "-- down migration for %s\n\n"

	if err := os.WriteFile(upPath, []byte(fmt.Sprintf(upTpl, safe)), 0644); err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to write up migration", err, errorhandler.WithOp("cli.db.createNewMigration"))
	}
	if err := os.WriteFile(downPath, []byte(fmt.Sprintf(downTpl, safe)), 0644); err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to write down migration", err, errorhandler.WithOp("cli.db.createNewMigration"))
	}
	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Created migrations: %s, %s", Vals: []any{upPath, downPath}})
	return nil
}

func sanitizeName(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	// remove any character not alnum or underscore
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Seeds
func initSeeds(dir string) error {
	if err := ensureDir(dir); err != nil {
		return err
	}
	p, err := db.DefaultPostgres()
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to connect database", err, errorhandler.WithOp("cli.db.initSeeds"))
	}
	defer p.Close()
	createTableSQL := `CREATE TABLE IF NOT EXISTS schema_seeds (
        name TEXT PRIMARY KEY,
        applied_at TIMESTAMP NOT NULL
    )`
	if _, err := p.Conn.Exec(createTableSQL); err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to ensure schema_seeds table", err, errorhandler.WithOp("cli.db.initSeeds"))
	}
	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Seeds initialized at %s", Vals: []any{dir}})
	return nil
}

func createNewSeed(dir, name string) error {
	if strings.TrimSpace(name) == "" {
		return errorhandler.New(errorhandler.KindInvalidArgument, "seed name cannot be empty", errorhandler.WithOp("cli.db.createNewSeed"))
	}
	safe := sanitizeName(name)
	filePath := filepath.Join(dir, fmt.Sprintf("%s.sql", safe))
	tpl := "-- seed for %s\n\n"
	if err := os.WriteFile(filePath, []byte(fmt.Sprintf(tpl, safe)), 0644); err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to write seed file", err, errorhandler.WithOp("cli.db.createNewSeed"))
	}
	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Created seed: %s", Vals: []any{filePath}})
	return nil
}

func applySeeds(p *db.Postgres, dir string) error {
	if err := ensureDir(dir); err != nil {
		return err
	}
	// read all .sql files sorted by name
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to read seeds directory", err, errorhandler.WithOp("cli.db.applySeeds"))
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	applied := map[string]struct{}{}
	rows, err := p.Conn.Query("SELECT name FROM schema_seeds")
	if err != nil {
		if !isUndefinedTableErr(err) {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to query schema_seeds", err, errorhandler.WithOp("cli.db.applySeeds"))
		}
	} else {
		defer rows.Close()
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err == nil {
				applied[n] = struct{}{}
			}
		}
	}
	appliedCount := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if _, ok := applied[e.Name()]; ok {
			continue
		}
		path := filepath.Join(dir, e.Name())
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to read seed", err, errorhandler.WithOp("cli.db.applySeeds"))
		}
		if _, err := p.Conn.Exec(string(sqlBytes)); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to apply seed", err, errorhandler.WithOp("cli.db.applySeeds"))
		}
		if _, err := p.Conn.Exec("INSERT INTO schema_seeds(name, applied_at) VALUES ($1, NOW())", e.Name()); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to record seed", err, errorhandler.WithOp("cli.db.applySeeds"))
		}
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Seed applied: %s", Vals: []any{e.Name()}})
		appliedCount++
	}
	if appliedCount == 0 {
		config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "No pending seeds"})
	}
	return nil
}

func seedsStatus(p *db.Postgres, dir string) error {
	if err := ensureDir(dir); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to read seeds directory", err, errorhandler.WithOp("cli.db.seedsStatus"))
	}
	names := []string{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	applied := map[string]struct{}{}
	rows, err := p.Conn.Query("SELECT name FROM schema_seeds")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err == nil {
				applied[n] = struct{}{}
			}
		}
	}
	for _, n := range names {
		state := "pending"
		if _, ok := applied[n]; ok {
			state = "applied"
		}
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "%s - %s", Vals: []any{n, state}})
	}
	return nil
}

func migrateUp(p *db.Postgres, dir string) error {
	if err := ensureDir(dir); err != nil {
		return err
	}
	pairs, err := readMigrationPairs(dir)
	if err != nil {
		return err
	}
	applied, err := loadAppliedVersions(p)
	if err != nil {
		return err
	}
	count := 0
	for _, m := range pairs {
		if _, ok := applied[m.Version]; ok {
			continue
		}
		if m.UpPath == "" {
			return errorhandler.New(errorhandler.KindInternal, fmt.Sprintf("missing up migration for version %s", m.Version), errorhandler.WithOp("cli.db.migrateUp"))
		}
		sqlBytes, err := os.ReadFile(m.UpPath)
		if err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to read up migration", err, errorhandler.WithOp("cli.db.migrateUp"))
		}
		if _, err := p.Conn.Exec(string(sqlBytes)); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, fmt.Sprintf("failed to apply migration %s", m.Version), err, errorhandler.WithOp("cli.db.migrateUp"))
		}
		if _, err := p.Conn.Exec("INSERT INTO schema_migrations(version, name, applied_at) VALUES ($1, $2, NOW())", m.Version, m.Name); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to record migration", err, errorhandler.WithOp("cli.db.migrateUp"))
		}
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Applied %s %s", Vals: []any{m.Version, m.Name}})
		count++
	}
	if count == 0 {
		config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "No pending migrations"})
	}
	return nil
}

func migrateDown(p *db.Postgres, dir string, steps int) error {
	if steps <= 0 {
		return errorhandler.New(errorhandler.KindInvalidArgument, "--steps must be >= 1", errorhandler.WithOp("cli.db.migrateDown"))
	}
	if err := ensureDir(dir); err != nil {
		return err
	}
	pairs, err := readMigrationPairs(dir)
	if err != nil {
		return err
	}
	appliedList, err := loadAppliedList(p)
	if err != nil {
		return err
	}
	if len(appliedList) == 0 {
		config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "No applied migrations to revert"})
		return nil
	}
	reverted := 0
	for i := len(appliedList) - 1; i >= 0 && reverted < steps; i-- {
		ver := appliedList[i].Version
		pair, ok := pairsByVersion(pairs)[ver]
		if !ok || pair.DownPath == "" {
			return errorhandler.New(errorhandler.KindInternal, fmt.Sprintf("missing down migration for version %s", ver), errorhandler.WithOp("cli.db.migrateDown"))
		}
		sqlBytes, err := os.ReadFile(pair.DownPath)
		if err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to read down migration", err, errorhandler.WithOp("cli.db.migrateDown"))
		}
		if _, err := p.Conn.Exec(string(sqlBytes)); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, fmt.Sprintf("failed to revert migration %s", ver), err, errorhandler.WithOp("cli.db.migrateDown"))
		}
		if _, err := p.Conn.Exec("DELETE FROM schema_migrations WHERE version = $1", ver); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to delete migration record", err, errorhandler.WithOp("cli.db.migrateDown"))
		}
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Reverted %s", Vals: []any{ver}})
		reverted++
	}
	return nil
}

func migrationsStatus(p *db.Postgres, dir string) error {
	if err := ensureDir(dir); err != nil {
		return err
	}
	pairs, err := readMigrationPairs(dir)
	if err != nil {
		return err
	}
	applied, err := loadAppliedVersions(p)
	if err != nil {
		return err
	}
	for _, m := range pairs {
		state := "pending"
		if _, ok := applied[m.Version]; ok {
			state = "applied"
		}
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "%s %s - %s", Vals: []any{m.Version, m.Name, state}})
	}
	return nil
}

func readMigrationPairs(dir string) ([]migrationPair, error) {
	entries := map[string]*migrationPair{}
	// Walk directory
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := filepath.Base(path)
		if !strings.HasSuffix(name, ".sql") {
			return nil
		}
		parts := strings.SplitN(name, "_", 2)
		if len(parts) != 2 {
			return nil
		}
		version := parts[0]
		rest := parts[1]
		var direction string
		if strings.HasSuffix(rest, ".up.sql") {
			direction = "up"
			rest = strings.TrimSuffix(rest, ".up.sql")
		} else if strings.HasSuffix(rest, ".down.sql") {
			direction = "down"
			rest = strings.TrimSuffix(rest, ".down.sql")
		} else {
			return nil
		}
		pair, ok := entries[version]
		if !ok {
			pair = &migrationPair{Version: version, Name: rest}
			entries[version] = pair
		}
		if direction == "up" {
			pair.UpPath = path
		} else {
			pair.DownPath = path
		}
		return nil
	}
	if err := filepath.WalkDir(dir, walkFn); err != nil {
		return nil, errorhandler.Wrap(errorhandler.KindInternal, "failed to read migrations directory", err, errorhandler.WithOp("cli.db.readMigrationPairs"))
	}
	list := make([]migrationPair, 0, len(entries))
	for _, p := range entries {
		list = append(list, *p)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Version < list[j].Version })
	return list, nil
}

func pairsByVersion(pairs []migrationPair) map[string]migrationPair {
	m := make(map[string]migrationPair, len(pairs))
	for _, p := range pairs {
		m[p.Version] = p
	}
	return m
}

type appliedRow struct {
	Version   string
	Name      string
	AppliedAt time.Time
}

func loadAppliedVersions(p *db.Postgres) (map[string]appliedRow, error) {
	rows, err := p.Conn.Query("SELECT version, name, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		// If the table doesn't exist, guide the user to run --init
		if isUndefinedTableErr(err) {
			return map[string]appliedRow{}, nil
		}
		return nil, errorhandler.Wrap(errorhandler.KindInternal, "failed to query schema_migrations", err, errorhandler.WithOp("cli.db.loadAppliedVersions"))
	}
	defer rows.Close()
	res := map[string]appliedRow{}
	for rows.Next() {
		var r appliedRow
		if err := rows.Scan(&r.Version, &r.Name, &r.AppliedAt); err != nil {
			return nil, errorhandler.Wrap(errorhandler.KindInternal, "failed to scan schema_migrations", err, errorhandler.WithOp("cli.db.loadAppliedVersions"))
		}
		res[r.Version] = r
	}
	return res, nil
}

func loadAppliedList(p *db.Postgres) ([]appliedRow, error) {
	rows, err := p.Conn.Query("SELECT version, name, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		if isUndefinedTableErr(err) {
			return []appliedRow{}, nil
		}
		return nil, errorhandler.Wrap(errorhandler.KindInternal, "failed to query schema_migrations", err, errorhandler.WithOp("cli.db.loadAppliedList"))
	}
	defer rows.Close()
	var list []appliedRow
	for rows.Next() {
		var r appliedRow
		if err := rows.Scan(&r.Version, &r.Name, &r.AppliedAt); err != nil {
			return nil, errorhandler.Wrap(errorhandler.KindInternal, "failed to scan schema_migrations", err, errorhandler.WithOp("cli.db.loadAppliedList"))
		}
		list = append(list, r)
	}
	return list, nil
}

func isUndefinedTableErr(err error) bool {
	if err == nil {
		return false
	}
	// A simple heuristic for lib/pq error text; avoid importing pq just for code check
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "undefined table") || strings.Contains(msg, "does not exist")
}
