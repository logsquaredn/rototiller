package datastore

import (
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/logsquaredn/geocloud"
)

var (
	//go:embed psql/execs/create_storage.sql
	createStorageSQL string

	//go:embed psql/execs/delete_storage.sql
	deleteStorageSQL string

	//go:embed psql/queries/get_storage_by_id.sql
	getStorageByIDSQL string

	//go:embed psql/execs/update_storage.sql
	updateStorageSQL string

	//go:embed psql/queries/get_storage_by_customer_id.sql
	getStorgageByCustomerIDSQL string

	//go:embed psql/queries/get_storage_before.sql
	getStorageBeforeSQL string

	//go:embed psql/queries/get_output_storage_by_job_id.sql
	getOutputStorageByJobIDSQL string

	//go:embed psql/queries/get_input_storage_by_job_id.sql
	getInputStorageByJobIDSQL string
)

func (p *Postgres) UpdateStorage(s *geocloud.Storage) (*geocloud.Storage, error) {
	var (
		storageStatus string
		err           error
	)
	if err := p.stmt.updateStorage.QueryRow(
		s.ID, s.Status, time.Now(),
	).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = geocloud.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) CreateStorage(s *geocloud.Storage) (*geocloud.Storage, error) {
	var (
		id            = uuid.NewString()
		storageStatus string
		err           error
	)

	if s.Status == "" {
		s.Status = geocloud.StorageStatusUnknown
	}

	if err = p.stmt.createStorage.QueryRow(
		id, s.Status, s.CustomerID, s.Name,
	).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = geocloud.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) GetStorage(m geocloud.Message) (*geocloud.Storage, error) {
	var (
		s             = &geocloud.Storage{}
		storageStatus string
		err           error
	)

	if err := p.stmt.getStorage.QueryRow(m.GetID()).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = geocloud.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) DeleteStorage(m geocloud.Message) error {
	_, err := p.stmt.deleteStorage.Exec(m.GetID())
	return err
}

func (p *Postgres) GetCustomerStorage(m geocloud.Message) ([]*geocloud.Storage, error) {
	rows, err := p.stmt.getStorageByCustomerID.Query(m.GetID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var storage []*geocloud.Storage

	for rows.Next() {
		var (
			s             = &geocloud.Storage{}
			storageStatus string
		)

		if err = rows.Scan(
			&s.ID, &storageStatus, &s.CustomerID,
			&s.Name, &s.LastUsed, &s.CreateTime,
		); err != nil {
			return nil, err
		}

		if s.Status, err = geocloud.ParseStorageStatus(storageStatus); err != nil {
			return nil, err
		}

		storage = append(storage, s)
	}

	return storage, nil
}

func (p *Postgres) GetJobInputStorage(m geocloud.Message) (*geocloud.Storage, error) {
	var (
		s             = &geocloud.Storage{}
		storageStatus string
		err           error
	)

	if err := p.stmt.getInputStorageByJobID.QueryRow(m.GetID()).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = geocloud.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) GetJobOutputStorage(m geocloud.Message) (*geocloud.Storage, error) {
	var (
		s             = &geocloud.Storage{}
		storageStatus string
		err           error
	)

	if err := p.stmt.getOutputStorageByJobID.QueryRow(m.GetID()).Scan(
		&s.ID, &storageStatus, &s.CustomerID,
		&s.Name, &s.LastUsed, &s.CreateTime,
	); err != nil {
		return nil, err
	}

	s.Status, err = geocloud.ParseStorageStatus(storageStatus)
	return s, err
}

func (p *Postgres) GetStorageBefore(d time.Duration) ([]*geocloud.Storage, error) {
	beforeTimestamp := time.Now().Add(-d)
	rows, err := p.stmt.getStorageBefore.Query(beforeTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var storages []*geocloud.Storage

	for rows.Next() {
		var (
			s             = &geocloud.Storage{}
			storageStatus string
		)

		if err = rows.Scan(
			&s.ID, &storageStatus, &s.CustomerID,
			&s.Name, &s.LastUsed, &s.CreateTime,
		); err != nil {
			return nil, err
		}

		if s.Status, err = geocloud.ParseStorageStatus(storageStatus); err != nil {
			return nil, err
		}

		storages = append(storages, s)
	}

	return storages, nil
}
