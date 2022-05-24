package main

import (
	"github.com/logsquaredn/geocloud"
	"github.com/spf13/cobra"
)

var (
	getStorageCmd = &cobra.Command{
		Use:     "storage",
		Aliases: []string{"storages", "s"},
		RunE:    runGetStorage,
		Args:    cobra.RangeArgs(0, 1),
	}
	createStorageCmd = &cobra.Command{
		Use:     "storage",
		Aliases: []string{"s"},
		RunE:    runCreateStorage,
		Args:    cobra.RangeArgs(0, 1),
	}
)

func init() {
	flags := getStorageCmd.PersistentFlags()
	flags.String("input-of", "", "Job ID to get storage input of")
	flags.String("output-of", "", "Job ID to get storage output of")
}

func init() {
	flags := createStorageCmd.PersistentFlags()
	flags.StringP("name", "n", "", "Name of storage")
}

func runGetStorage(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	var (
		s any
		i = cmd.Flag("input-of").Value.String()
		o = cmd.Flag("output-of").Value.String()
	)

	switch {
	case len(args) > 0:
		s, err = client.GetStorage(args[0])
	case o != "":
		s, err = client.GetJobOutput(o)
	case i != "":
		s, err = client.GetJobInput(i)
	default:
		s, err = client.GetStorages()
	}
	if err != nil {
		return err
	}

	return write(s)
}

func runCreateStorage(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	i, err := getInput(cmd)
	if err != nil {
		return err
	}

	s, err := client.CreateStorage(geocloud.NewStorageWithName(i, cmd.Flag("name").Value.String()))
	if err != nil {
		return err
	}

	return write(s)
}