package datastore

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/logsquaredn/geocloud"
)

var (
	//go:embed psql/execs/create_job.sql
	createJobSQL string

	//go:embed psql/execs/delete_job.sql
	deleteJobSQL string

	//go:embed psql/execs/update_job.sql
	updateJobSQL string

	//go:embed psql/queries/get_jobs_before.sql
	getJobsBeforeSQL string

	//go:embed psql/queries/get_job_by_id.sql
	getJobByIDSQL string

	//go:embed psql/queries/get_job_by_customer_id.sql
	getJobsByCustomerIDSQL string
)

func (p *Postgres) CreateJob(j *geocloud.Job) (*geocloud.Job, error) {
	var (
		id        = uuid.New().String()
		jobErr    sql.NullString
		jobStatus string
		endTime   sql.NullTime
		taskType  string
	)

	err := p.stmt.createJob.QueryRow(
		id, j.CustomerID,
		j.InputID, j.OutputID,
		j.TaskType.String(),
		pq.Array(j.Args),
	).Scan(
		&j.ID, &j.CustomerID,
		&j.InputID, &j.OutputID,
		&taskType,
		&jobStatus, &jobErr,
		&j.StartTime, &endTime,
		pq.Array(&j.Args),
	)
	if err != nil {
		return j, err
	}

	if jobErr.String != "" {
		j.Err = fmt.Errorf(jobErr.String)
	}
	j.EndTime = endTime.Time

	j.TaskType, err = geocloud.TaskTypeFrom(taskType)
	if err != nil {
		return j, err
	}

	j.Status, err = geocloud.JobStatusFrom(jobStatus)
	if err != nil {
		return j, err
	}

	return j, nil
}

func (p *Postgres) UpdateJob(j *geocloud.Job) (*geocloud.Job, error) {
	var (
		jobErr      sql.NullString
		jobStatus   string
		endTime     sql.NullTime
		taskType    string
		jobErrError = ""
	)

	// avoid nil pointer dereference on j.Err.Error()
	if j.Err != nil {
		jobErrError = j.Err.Error()
	}

	err := p.stmt.updateJob.QueryRow(
		j.GetID(),
		j.Status.String(), jobErrError,
		j.StartTime, j.EndTime,
	).Scan(
		&j.ID, &j.CustomerID,
		&j.InputID, &j.OutputID,
		&taskType,
		&jobStatus, &jobErr,
		&j.StartTime, &endTime,
		pq.Array(&j.Args),
	)
	if err != nil {
		return j, err
	}

	if jobErr.String != "" {
		j.Err = fmt.Errorf(jobErr.String)
	}
	j.EndTime = endTime.Time

	j.TaskType, err = geocloud.TaskTypeFrom(taskType)
	if err != nil {
		return j, err
	}

	j.Status, err = geocloud.JobStatusFrom(jobStatus)
	if err != nil {
		return j, err
	}

	return j, nil
}

func (p *Postgres) GetJob(m geocloud.Message) (*geocloud.Job, error) {
	var (
		j         = &geocloud.Job{}
		jobErr    sql.NullString
		jobStatus string
		endTime   sql.NullTime
		taskType  string
	)

	err := p.stmt.getJobByID.QueryRow(m.GetID()).Scan(
		&j.ID, &j.CustomerID,
		&j.InputID, &j.OutputID,
		&taskType,
		&jobStatus, &jobErr,
		&j.StartTime, &endTime,
		pq.Array(&j.Args),
	)
	if err != nil {
		return j, err
	}

	if jobErr.String != "" {
		j.Err = fmt.Errorf(jobErr.String)
	}
	j.EndTime = endTime.Time

	j.TaskType, err = geocloud.TaskTypeFrom(taskType)
	if err != nil {
		return j, err
	}

	j.Status, err = geocloud.JobStatusFrom(jobStatus)
	if err != nil {
		return j, err
	}

	return j, nil
}

func (p *Postgres) GetJobsBefore(d time.Duration) ([]*geocloud.Job, error) {
	beforeTimestamp := time.Now().Add(-d)
	rows, err := p.stmt.getJobsBefore.Query(beforeTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*geocloud.Job

	for rows.Next() {
		var (
			j         = &geocloud.Job{}
			jobErr    sql.NullString
			jobStatus string
			endTime   sql.NullTime
			taskType  string
		)

		err = rows.Scan(
			&j.ID, &j.CustomerID,
			&j.InputID, &j.OutputID,
			&taskType,
			&jobStatus, &jobErr,
			&j.StartTime, &endTime,
			pq.Array(&j.Args),
		)
		if err != nil {
			return nil, err
		}

		if jobErr.String != "" {
			j.Err = fmt.Errorf(jobErr.String)
		}
		j.EndTime = endTime.Time

		j.TaskType, err = geocloud.TaskTypeFrom(taskType)
		if err != nil {
			return nil, err
		}

		j.Status, err = geocloud.JobStatusFrom(jobStatus)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}

func (p *Postgres) DeleteJob(m geocloud.Message) error {
	_, err := p.stmt.deleteJob.Exec(m.GetID())
	return err
}

func (p *Postgres) GetCustomerJobs(m geocloud.Message) ([]*geocloud.Job, error) {
	rows, err := p.stmt.getJobsByCustomerID.Query(m.GetID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*geocloud.Job

	for rows.Next() {
		var (
			j         = &geocloud.Job{}
			jobErr    sql.NullString
			jobStatus string
			endTime   sql.NullTime
			taskType  string
		)

		err = rows.Scan(
			&j.ID, &j.CustomerID,
			&j.InputID, &j.OutputID,
			&taskType,
			&jobStatus, &jobErr,
			&j.StartTime, &endTime,
			pq.Array(&j.Args),
		)
		if err != nil {
			return nil, err
		}

		if jobErr.String != "" {
			j.Err = fmt.Errorf(jobErr.String)
		}
		j.EndTime = endTime.Time

		j.TaskType, err = geocloud.TaskTypeFrom(taskType)
		if err != nil {
			return nil, err
		}

		j.Status, err = geocloud.JobStatusFrom(jobStatus)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}