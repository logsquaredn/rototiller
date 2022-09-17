package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/frantjc/go-js"
	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api"
)

const (
	qInput    = "input"
	qInputOf  = "input-of"
	qOutputOf = "output-of"
)

func (a *Handler) createJobForOwner(ctx *gin.Context, taskType api.TaskType, ownerID string) (*api.Job, error) {
	task, err := a.getTaskType(taskType)
	if err != nil {
		return nil, err
	}

	var (
		input    = ctx.Query(qInput)
		inputOf  = ctx.Query(qInputOf)
		outputOf = ctx.Query(qOutputOf)
		inputIDs = js.Filter(
			[]string{input, inputOf, outputOf},
			func(s string, _ int, _ []string) bool {
				return s != ""
			},
		)
		storage *api.Storage
	)
	switch {
	case len(inputIDs) > 1:
		return nil, api.NewErr(fmt.Errorf("cannot specify more than one of queries '%s', '%s' and '%s'", qInput, qInputOf, qOutputOf), http.StatusBadRequest)
	case input != "":
		storage, err = a.getStorageForOwner(input, ownerID)
		if err != nil {
			return nil, err
		}
	case inputOf != "":
		storage, err = a.getJobInputStorageForOwner(ctx, inputOf, ownerID)
		if err != nil {
			return nil, err
		}
	case outputOf != "":
		storage, err = a.getJobOutputStorageForOwner(ctx, outputOf, ownerID)
		if err != nil {
			return nil, err
		}
	default:
		storage, err = a.putRequestVolumeForOwner(ctx, ctx.GetHeader("Content-Type"), ctx.Query("name"), ctx.Request.Body, ownerID)
		if err != nil {
			return nil, err
		}

		defer ctx.Request.Body.Close()
	}

	switch api.StorageStatus(storage.Status) {
	case api.StorageStatusFinal:
		return nil, api.NewErr(fmt.Errorf("cannot create job, storage id %s is final", storage.Id), http.StatusBadRequest)
	case api.StorageStatusUnusable:
		return nil, api.NewErr(fmt.Errorf("cannot create job, storage id %s is unsusable", storage.Id), http.StatusBadRequest)
	}

	job, err := a.Datastore.CreateJob(&api.Job{
		TaskType: task.Type,
		Args:     buildJobArgs(ctx, task.Params),
		OwnerId:  ownerID,
		InputId:  storage.Id,
	})
	if err != nil {
		return nil, err
	}

	if err = a.EventStreamProducer.Emit(ctx, &api.Event{
		Type: api.EventTypeJobCreated.String(),
		Metadata: map[string]string{
			"id": job.Id,
		},
	}); err != nil {
		return nil, err
	}

	return job, nil
}

func (a *Handler) createJob(ctx *gin.Context, taskType api.TaskType) (*api.Job, error) {
	ownerID, err := a.getOwnerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.createJobForOwner(ctx, taskType, ownerID)
}

func (a *Handler) getJob(ctx *gin.Context, id string) (*api.Job, error) {
	ownerID, err := a.getOwnerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.getJobForOwner(ctx, id, ownerID)
}

func (a *Handler) getJobForOwner(ctx *gin.Context, id string, ownerID string) (*api.Job, error) {
	job, err := a.Datastore.GetJob(id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, api.NewErr(fmt.Errorf("job '%s' not found", id), http.StatusNotFound)
	case err != nil:
		return nil, err
	}

	return a.checkJobOwnership(job, ownerID)
}

func (a *Handler) checkJobOwnership(job *api.Job, ownerID string) (*api.Job, error) {
	if job.OwnerId != ownerID {
		return nil, api.NewErr(fmt.Errorf("user does not own job '%s'", job.Id), http.StatusForbidden)
	}

	return job, nil
}

func buildJobArgs(ctx *gin.Context, taskParams []string) []string {
	jobArgs := make([]string, len(taskParams))
	for i, param := range taskParams {
		jobArgs[i] = ctx.Query(param)
	}
	return jobArgs
}
