package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	envhandler "github.com/Arthur-Conti/guh/libs/env_handler"
	envlocations "github.com/Arthur-Conti/guh/libs/env_handler/env_locations"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	_ "github.com/lib/pq"
)

type PostgresOpts struct {
	Ctx      context.Context
	User     string
	Password string
	Database string
	IP       string
	Port     string
}

type Postgres struct {
	Opts PostgresOpts
	Conn *sql.DB
}

func GetDefaultPostgresOpts() *PostgresOpts {
	env := envhandler.NewEnvs(envlocations.NewLocalEnvs("./.env"))
	if err := env.EnvLocation.LoadDotEnv(); err != nil {
		return nil
	}
	return &PostgresOpts{
		Ctx:      context.Background(),
		User:     env.EnvLocation.Get("DB_USER"),
		Password: env.EnvLocation.Get("DB_PASS"),
		IP:       env.EnvLocation.Get("DB_IP"),
		Port:     env.EnvLocation.Get("DB_PORT"),
		Database: env.EnvLocation.Get("DB_DATABASE"),
	}
}

func DefaultPostgres() (*Postgres, error) {
	p := NewPostgres(*GetDefaultPostgresOpts())
	if err := p.Connect(); err != nil {
		return nil, err
	}
	return p, nil
}

func NewPostgres(opts PostgresOpts) *Postgres {
	return &Postgres{
		Opts: opts,
	}
}

func (p *Postgres) Connect() error {
	return p.init()
}

func (p *Postgres) Close() error {
	if err := p.Conn.Close(); err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Error closing db", err, errorhandler.WithOp("db.Close"))
	}
	return nil
}

func (p *Postgres) uri() string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", p.Opts.User, p.Opts.Password, p.Opts.IP, p.Opts.Port, p.Opts.Database)
}

func (p *Postgres) init() error {
	conn, err := sql.Open("postgres", p.uri())
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "unable to connect to database", err, errorhandler.WithOp("db.init"))
	}
	err = conn.Ping()
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindUnavailable, "ping failed", err, errorhandler.WithOp("db.init"))
	}
	p.Conn = conn
	return nil
}

func (p *Postgres) CreateTable(sql string) error {
	_, err := p.Conn.Query(sql)
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Error creating table", err, errorhandler.WithOp("db.CreateTable"), errorhandler.WithFields(map[string]any{"sql": sql}))
	}
	return nil
}

func (p *Postgres) QueryRow(dest any, query string, args ...any) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errorhandler.New(errorhandler.KindInvalidArgument, "dest must be a pointer to a struct", errorhandler.WithOp("db.QueryRow"))
	}

	rows, err := p.Conn.Query(query, args...)
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Query failed", err, errorhandler.WithOp("db.QueryRow"), errorhandler.WithFields(map[string]any{"query": query, "args": args}))
	}
	defer rows.Close()

	if !rows.Next() {
		return errorhandler.New(errorhandler.KindNotFound, "error no rows", errorhandler.WithOp("db.QueryRow"))
	}

	columns, err := rows.Columns()
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Failed to get columns", err, errorhandler.WithOp("db.QueryRow"), errorhandler.WithFields(map[string]any{"query": query, "args": args}))
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Failed to scan row", err, errorhandler.WithOp("db.QueryRow"), errorhandler.WithFields(map[string]any{"query": query, "args": args}))
	}

	structValue := v.Elem()
	structType := structValue.Type()

	for i, colName := range columns {
		for j := 0; j < structType.NumField(); j++ {
			field := structType.Field(j)
			tag := field.Tag.Get("db")
			if tag == "" {
				tag = strings.ToLower(field.Name)
			}

			if strings.EqualFold(tag, colName) {
				fieldValue := structValue.Field(j)
				if fieldValue.CanSet() {
					err := setFieldValue(fieldValue, values[i])
					if err != nil {
						return errorhandler.Wrap(errorhandler.KindInternal, "Error setting field", err, errorhandler.WithOp("db.QueryRow"), errorhandler.WithFields(map[string]any{"query": query, "args": args}))
					}
				}
				break
			}
		}
	}

	return nil
}

func (p *Postgres) Query(dest any, query string, args ...any) error {
	ptrVal := reflect.ValueOf(dest)
	if ptrVal.Kind() != reflect.Ptr {
		return errorhandler.New(errorhandler.KindInvalidArgument, "dest must be a pointer to a slice", errorhandler.WithOp("db.Query"))
	}
	sliceVal := ptrVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errorhandler.New(errorhandler.KindInvalidArgument, "dest must point to a slice", errorhandler.WithOp("db.Query"))
	}

	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return errorhandler.New(errorhandler.KindInvalidArgument, "Slice element type must be struct", errorhandler.WithOp("db.Query"))
	}

	rows, err := p.Conn.Query(query, args...)
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Query failed", err, errorhandler.WithOp("db.Query"), errorhandler.WithFields(map[string]any{"query": query, "args": args}))
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Failed to get columns", err, errorhandler.WithOp("db.Query"), errorhandler.WithFields(map[string]any{"query": query, "args": args}))
	}

	for rows.Next() {
		newElemPtr := reflect.New(elemType)
		newElem := newElemPtr.Elem()

		values := make([]any, len(columns))
		for i := range values {
			var v any
			values[i] = &v
		}

		if err := rows.Scan(values...); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "Scan failed", err, errorhandler.WithOp("db.Query"), errorhandler.WithFields(map[string]any{"query": query, "args": args}))
		}

		for i, colName := range columns {
			rawValPtr := values[i].(*any)
			val := *rawValPtr

			for j := 0; j < elemType.NumField(); j++ {
				field := elemType.Field(j)
				tag := field.Tag.Get("db")
				if tag == "" {
					tag = strings.ToLower(field.Name)
				}
				if strings.EqualFold(tag, colName) {
					fieldValue := newElem.Field(j)
					if fieldValue.CanSet() {
						if err := setFieldValue(fieldValue, val); err != nil {
							return errorhandler.Wrap(errorhandler.KindInternal, "Failed to set field "+field.Name, err, errorhandler.WithOp("db.Query"), errorhandler.WithFields(map[string]any{"query": query, "args": args}))
						}
					}
					break
				}
			}
		}

		sliceVal.Set(reflect.Append(sliceVal, newElem))
	}

	return nil
}

func setFieldValue(fieldValue reflect.Value, val any) error {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case []byte:
		strVal := string(v)
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(strVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			iv, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				return err
			}
			fieldValue.SetInt(iv)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uv, err := strconv.ParseUint(strVal, 10, 64)
			if err != nil {
				return err
			}
			fieldValue.SetUint(uv)
		case reflect.Float32, reflect.Float64:
			fv, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				return err
			}
			fieldValue.SetFloat(fv)
		case reflect.Bool:
			bv, err := strconv.ParseBool(strVal)
			if err != nil {
				return err
			}
			fieldValue.SetBool(bv)
		default:
			// Unsupported type; do nothing or return error
		}
	case int64:
		if fieldValue.Kind() >= reflect.Int && fieldValue.Kind() <= reflect.Int64 {
			fieldValue.SetInt(v)
		}
	case float64:
		if fieldValue.Kind() == reflect.Float32 || fieldValue.Kind() == reflect.Float64 {
			fieldValue.SetFloat(v)
		}
	case bool:
		if fieldValue.Kind() == reflect.Bool {
			fieldValue.SetBool(v)
		}
	case string:
		if fieldValue.Kind() == reflect.String {
			fieldValue.SetString(v)
		}
	default:
		rv := reflect.ValueOf(val)
		if rv.Type().AssignableTo(fieldValue.Type()) {
			fieldValue.Set(rv)
		}
	}
	return nil
}
