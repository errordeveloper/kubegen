package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/errordeveloper/kubegen/pkg/apps"
	"github.com/errordeveloper/kubegen/pkg/util"
)

var (
	image        string
	port         int32
	replicas     int32
	env          []string
	flavor       string
	outputFormat string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "kubegen",
		RunE: command,
	}

	rootCmd.Flags().StringVarP(&image, "image", "i", "", "Container image to use")
	rootCmd.Flags().Int32VarP(&replicas, "replicas", "r", apps.DEFAULT_REPLICAS, "Number of pods")
	rootCmd.Flags().Int32VarP(&port, "port", "p", apps.DEFAULT_PORT, "Container and service port to use")
	rootCmd.Flags().StringSliceVar(&env, "env", []string{}, "Environment variables to set")
	rootCmd.Flags().StringVarP(&flavor, "flavor", "F", apps.DefaultFlavor, "Flavor of generator to use")
	rootCmd.Flags().StringVarP(&outputFormat, "output-format", "o", "yaml-stdout", "Output format [\"yaml-stdout\", \"json-stdout\", \"yaml-files\"]")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func command(cmd *cobra.Command, args []string) error {
	if image == "" {
		return fmt.Errorf("Image must be specified")
	}

	component := apps.AppComponent{
		Image:    image,
		Port:     port,
		Replicas: &replicas,
		Flavor:   flavor,
	}

	app := &apps.App{Components: []apps.AppComponent{component}}

	if len(env) > 0 {
		component.Env = make(map[string]string)
		for _, e := range env {
			parts := strings.Split(e, "=")
			if len(parts) != 2 {
				return fmt.Errorf("Invalid environment variable: %q", e)
			}
			component.Env[parts[0]] = parts[1]
		}
	}

	var (
		data        []byte
		err         error
		wroteFiles  []string
		contentType string
	)
	switch outputFormat {
	case "yaml-stdout":
		contentType = "yaml"
		if data, err = app.EncodeListToYAML(); err != nil {
			return err
		}
	case "json-stdout":
		contentType = "json"
		if data, err = app.EncodeListToPrettyJSON(); err != nil {
			return err
		}
	case "yaml-files":
		contentType = "yaml"
		if wroteFiles, err = app.DumpListToFilesAsYAML(); err != nil {
			return err
		}
	}

	if len(wroteFiles) > 0 {
		fmt.Printf("Wrote these files in the current working directory:\n %s", wroteFiles)
	} else {
		if err := util.Dump(contentType, data); err != nil {
			return err
		}
	}

	return nil
}
