package runtime

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

type OS struct {
	ds      *datastore.Postgres
	os      *objectstore.S3
	workdir string
}

func NewOS(opts *OSRuntimeOpts) (*OS, error) {
	return &OS{
		ds:      opts.Datastore,
		os:      opts.Objectstore,
		workdir: opts.WorkDir,
	}, nil
}

func (o *OS) Send(m geocloud.Message) error {
	k, v := "id", m.GetID()
	log.Info().Str(k, v).Msg("processing message")

	log.Trace().Str(k, v).Msg("getting job from datastore")
	j, err := o.ds.GetJob(m)
	if err != nil {
		return err
	}

	switch j.Status {
	case geocloud.Complete, geocloud.InProgress:
		return nil
	}

	stderr := new(bytes.Buffer)
	defer func() {
		j.EndTime = time.Now()
		jobErr := stderr.Bytes()
		if len(jobErr) > 0 {
			j.Err = fmt.Errorf("%s", jobErr)
			j.Status = geocloud.Error
		} else if err != nil {
			j.Err = err
			j.Status = geocloud.Error
		} else {
			j.Status = geocloud.Complete
		}
		log.Err(j.Err).Str(k, v).Msgf("job finished with status %s", j.Status.Status())
		o.ds.UpdateJob(j)
	}()

	go func() {
		log.Debug().Str(k, v).Msg("getting input storage")
		ist, _ := o.ds.GetStorage(geocloud.NewMessage(j.InputID))
		log.Debug().Str(k, v).Msg("updating input storage")
		o.ds.UpdateStorage(ist)
	}()

	go func() {
		log.Debug().Str(k, v).Msg("getting output storage")
		ost, _ := o.ds.GetStorage(geocloud.NewMessage(j.OutputID))
		log.Debug().Str(k, v).Msg("updating output storage")
		o.ds.UpdateStorage(ost)
	}()

	j.Status = geocloud.InProgress
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
	if err = input.Download(invol.path); err != nil {
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
			filepath.Join(invol.path, filename),
			outvol.path,
		},
		j.Args...,
	)
	task := exec.Command(t.Type.Name(), args...)
	task.Stdin = os.Stdin
	task.Stdout = os.Stdout
	task.Stderr = stderr

	log.Info().Str(k, v).Msgf("running task %s", task.Path)
	if err := task.Run(); err != nil {
		return err
	}

	log.Debug().Str(k, v).Msg("uploading output")
	if err = o.os.PutObject(geocloud.NewMessage(j.OutputID), outvol); err != nil {
		return err
	}

	return nil
}

func volume(path string) (*dirVolume, error) {
	return &dirVolume{path: path}, os.MkdirAll(path, 0755)
}

func (o *OS) involume(m geocloud.Message) (*dirVolume, error) {
	return volume(filepath.Join(o.jobdir(m), "input"))
}

func (o *OS) outvolume(m geocloud.Message) (*dirVolume, error) {
	return volume(filepath.Join(o.jobdir(m), "output"))
}

func (o *OS) jobdir(m geocloud.Message) string {
	return filepath.Join(o.jobsdir(), m.GetID())
}

func (o *OS) jobsdir() string {
	return filepath.Join(o.workdir, "jobs")
}
