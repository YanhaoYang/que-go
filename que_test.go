package que

import (
	"os"
	"testing"

	"github.com/jackc/pgx"
)

func connConfig() pgx.ConnConfig {
	config := pgx.ConnConfig{
		Host:     "localhost",
		Database: "que-go-test",
	}

	dbURL := os.Getenv("DB_URL")
	var err error
	if dbURL != "" {
		config, err = pgx.ParseURI(dbURL)
		if err != nil {
			panic("Cannot parse db URL")
		}
	}
	return config
}

var testConnConfig = connConfig()

func openTestClientMaxConns(t testing.TB, maxConnections int) *Client {
	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig:     testConnConfig,
		MaxConnections: maxConnections,
		AfterConnect:   PrepareStatements,
	}
	pool, err := pgx.NewConnPool(connPoolConfig)
	if err != nil {
		t.Fatal(err)
	}
	return NewClient(pool)
}

func openTestClient(t testing.TB) *Client {
	return openTestClientMaxConns(t, 5)
}

func truncateAndClose(pool *pgx.ConnPool) {
	if _, err := pool.Exec("TRUNCATE TABLE que_jobs"); err != nil {
		panic(err)
	}
	pool.Close()
}

func findOneJob(q queryable) (*Job, error) {
	findSQL := `
	SELECT priority, run_at, job_id, job_class, args, error_count, last_error, queue
	FROM que_jobs LIMIT 1`

	j := &Job{}
	err := q.QueryRow(findSQL).Scan(
		&j.Priority,
		&j.RunAt,
		&j.ID,
		&j.Type,
		&j.Args,
		&j.ErrorCount,
		&j.LastError,
		&j.Queue,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return j, nil
}
