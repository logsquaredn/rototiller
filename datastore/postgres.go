package datastore

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	// postgres must be imported to inject the postgres driver
	// into the database/sql package
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

//go:embed psql/migrations/*.up.sql
var migrations embed.FS

type Postgres struct {
	db   *sql.DB
	stmt *struct {
		createJob               *sql.Stmt
		createCustomer          *sql.Stmt
		updateJob               *sql.Stmt
		getJobByID              *sql.Stmt
		getJobsBefore           *sql.Stmt
		deleteJob               *sql.Stmt
		getTaskByJobID          *sql.Stmt
		getTaskByType           *sql.Stmt
		getTasksByTypes         *sql.Stmt
		getCustomerByID         *sql.Stmt
		getStorage              *sql.Stmt
		createStorage           *sql.Stmt
		deleteStorage           *sql.Stmt
		updateStorage           *sql.Stmt
		getStorageByCustomerID  *sql.Stmt
		getStorageBefore        *sql.Stmt
		getJobsByCustomerID     *sql.Stmt
		getOutputStorageByJobID *sql.Stmt
		getInputStorageByJobID  *sql.Stmt
	}
}

func NewPostgres(opts *PostgresOpts) (*Postgres, error) {
	var (
		p = &Postgres{
			stmt: &struct {
				createJob               *sql.Stmt
				createCustomer          *sql.Stmt
				updateJob               *sql.Stmt
				getJobByID              *sql.Stmt
				getJobsBefore           *sql.Stmt
				deleteJob               *sql.Stmt
				getTaskByJobID          *sql.Stmt
				getTaskByType           *sql.Stmt
				getTasksByTypes         *sql.Stmt
				getCustomerByID         *sql.Stmt
				getStorage              *sql.Stmt
				createStorage           *sql.Stmt
				deleteStorage           *sql.Stmt
				updateStorage           *sql.Stmt
				getStorageByCustomerID  *sql.Stmt
				getStorageBefore        *sql.Stmt
				getJobsByCustomerID     *sql.Stmt
				getOutputStorageByJobID *sql.Stmt
				getInputStorageByJobID  *sql.Stmt
			}{},
		}
		err error
		i   int64 = 1
	)
	for p.db, err = sql.Open("postgres", opts.connectionString()); err != nil; p.db, err = sql.Open("postgres", opts.connectionString()) {
		if i >= opts.Retries && opts.Retries > 0 {
			return nil, fmt.Errorf("failed to connect to db after %d attempts: %w", i, err)
		}
		time.Sleep(opts.RetryDelay)
		i++
	}

	i = 1
	for ; err != nil; err = p.db.Ping() {
		if i >= opts.Retries && opts.Retries > 0 {
			return nil, fmt.Errorf("failed to ping db after %d attempts: %w", i, err)
		}
		time.Sleep(opts.RetryDelay)
		i++
	}

	if p.stmt.createJob, err = p.db.Prepare(createJobSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.createCustomer, err = p.db.Prepare(createCustomerSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.updateJob, err = p.db.Prepare(updateJobSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getJobByID, err = p.db.Prepare(getJobByIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getJobsBefore, err = p.db.Prepare(getJobsBeforeSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.deleteJob, err = p.db.Prepare(deleteJobSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getTaskByJobID, err = p.db.Prepare(getTaskByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getTaskByType, err = p.db.Prepare(getTaskByTypeSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getTasksByTypes, err = p.db.Prepare(getTasksByTypesSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	if p.stmt.getCustomerByID, err = p.db.Prepare(getCustomerByIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if p.stmt.createStorage, err = p.db.Prepare(createStorageSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if p.stmt.deleteStorage, err = p.db.Prepare(deleteStorageSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if p.stmt.getStorage, err = p.db.Prepare(getStorageByIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if p.stmt.updateStorage, err = p.db.Prepare(updateStorageSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if p.stmt.getJobsByCustomerID, err = p.db.Prepare(getJobsByCustomerIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if p.stmt.getStorageByCustomerID, err = p.db.Prepare(getStorgageByCustomerIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if p.stmt.getStorageBefore, err = p.db.Prepare(getStorageBeforeSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if p.stmt.getOutputStorageByJobID, err = p.db.Prepare(getOutputStorageByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	if p.stmt.getInputStorageByJobID, err = p.db.Prepare(getInputStorageByJobIDSQL); err != nil {
		return nil, fmt.Errorf("failed to prepare statement; %w", err)
	}

	return p, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}
