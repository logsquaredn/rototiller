package worker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	ds      *datastore.Postgres
	os      *objectstore.S3
	workdir string
}

func New(opts *Opts) (*Worker, error) {
	return &Worker{
		ds:      opts.Datastore,
		os:      opts.Objectstore,
		workdir: opts.WorkDir,
	}, nil
}

func (o *Worker) Send(m geocloud.Message) error {
	k, v := "id", m.GetID()
	log.Info().Str(k, v).Msg("processing message")

	log.Trace().Str(k, v).Msg("getting job from datastore")
	j, err := o.ds.GetJob(m)
	if err != nil {
		return err
	}

	switch j.Status {
	case geocloud.JobStatusComplete, geocloud.JobStatusInProgress:
		return nil
	}

	var (
		stderr   = new(bytes.Buffer)
		outputID = ""
	)
	defer func() {
		j.EndTime = time.Now()
		jobErr := stderr.Bytes()
		switch {
		case len(jobErr) > 0:
			j.Error = fmt.Sprint(jobErr)
			j.Status = geocloud.JobStatusError
		case err != nil:
			j.Error = err.Error()
			j.Status = geocloud.JobStatusError
		default:
			j.Status = geocloud.JobStatusComplete
			j.OutputID = outputID
		}
		log.Err(fmt.Errorf(j.Error)).Str(k, v).Msgf("job finished with status %s", j.Status.Status())
		o.ds.UpdateJob(j)
	}()

	go func() {
		log.Debug().Str(k, v).Msg("getting input storage")
		ist, _ := o.ds.GetStorage(geocloud.NewMessage(j.InputID))
		log.Debug().Str(k, v).Msg("updating input storage")
		o.ds.UpdateStorage(ist)
	}()

	j.Status = geocloud.JobStatusInProgress
	log.Trace().Str(k, v).Msgf("setting job to %s", j.Status.Status())
	j, err = o.ds.UpdateJob(j)
	if err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("getting task for job from datastore")
	t, err := o.ds.GetTaskByJobID(m)
	if err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("creating input volume")
	invol, err := o.involume(m)
	if err != nil {
		return err
	}
	defer os.RemoveAll(o.jobdir(m))

	log.Trace().Str(k, v).Msg("getting input")
	input, err := o.os.GetObject(geocloud.NewMessage(j.InputID))
	if err != nil {
		return err
	}

	log.Debug().Str(k, v).Msg("downloading input")
	if err = input.Download(o.involumePath(j)); err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("creating output volume")
	outvol, err := o.outvolume(m)
	if err != nil {
		return err
	}

	var filename string
	invol.Walk(func(_ string, f geocloud.File, e error) error {
		if e != nil {
			return e
		}
		filename = f.Name()
		return fmt.Errorf("found") // we only expect 1 input, so use the first one we find
	})

	if filename == "" {
		return fmt.Errorf("no input found")
	}

	args := append(
		[]string{
			filepath.Join(o.involumePath(j), filename),
			o.outvolumePath(j),
		},
		j.Args...,
	)
	task := exec.Command(t.Type.Name(), args...)
	task.Env = []string{}
	task.Stdin = os.Stdin
	task.Stdout = os.Stdout
	task.Stderr = stderr

	log.Info().Str(k, v).Msgf("running task %s", task.Path)
	if err := task.Run(); err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("creating output storage")
	ost, err := o.ds.CreateStorage(&geocloud.Storage{
		CustomerID: j.CustomerID,
	})
	if err != nil {
		return err
	}

	outputID = ost.ID
	j.OutputID = ost.ID
	log.Trace().Str(k, v).Msgf("updating job output")
	j, err = o.ds.UpdateJob(j)
	if err != nil {
		return err
	}

	log.Debug().Str(k, v).Msg("uploading output")
	if err = o.os.PutObject(geocloud.NewMessage(j.OutputID), outvol); err != nil {
		return err
	}

	return nil
}

func (o *Worker) involumePath(m geocloud.Message) string {
	return filepath.Join(o.jobdir(m), "input")
}

func (o *Worker) outvolumePath(m geocloud.Message) string {
	return filepath.Join(o.jobdir(m), "output")
}

func (o *Worker) involume(m geocloud.Message) (geocloud.Volume, error) {
	return geocloud.NewDirVolume(o.involumePath(m))
}

func (o *Worker) outvolume(m geocloud.Message) (geocloud.Volume, error) {
	return geocloud.NewDirVolume(o.outvolumePath(m))
}

func (o *Worker) jobdir(m geocloud.Message) string {
	return filepath.Join(o.jobsdir(), m.GetID())
}

func (o *Worker) jobsdir() string {
	return filepath.Join(o.workdir, "jobs")
}