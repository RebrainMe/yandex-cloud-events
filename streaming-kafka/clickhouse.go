package main

import (
	// clickhouse driver
	"github.com/kshvakov/clickhouse"
	// monitoring
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	// statement for inserting data into clickhouse
	ChPrepareStmt string = "INSERT INTO events (ts, gender, age, path, browser, os) VALUES (?,?,?,?,?,?)"
)

var (
	chDB *sql.DB
)

func initClickhouse() error {
	var err error
	log.Printf("INFO: Connecting to clickhouse")

	chDB, err = sql.Open("clickhouse", *ClickHouseDSN)
	if err != nil {
		return err
	}

	if err := chDB.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			return fmt.Errorf("[%d] %s %s", exception.Code, exception.Message, exception.StackTrace)
		}
		return err
	}
	log.Printf("INFO: Successfully connected to ClickHouse")

	return nil
}

func flushEvents(events []event) error {
	start := time.Now()

	// Starting transaction
	tx, err := chDB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(ChPrepareStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Preparing all events
	for _, ev := range events {
		_, err = stmt.Exec(
			time.Unix(ev.TS, 0),
			ev.Gender,
			ev.Age,
			ev.Path,
			ev.Browser,
			ev.OS,
		)
		if err != nil {
			return err
		}
	}

	// Flushing to clickhouse
	tx.Commit()

	// Publishing some metrics
	elapsed := time.Since(start)
	ChFlushTime.Set(elapsed.Seconds())
	ChEventsFlushed.Set(float64(len(events)))

	log.Printf("INFO: Flushed data to clickhouse in %s, count: %d\n", elapsed, len(events))
	return nil
}
