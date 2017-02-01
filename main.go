package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/d4l3k/go-highlight"
	"github.com/docker/docker/pkg/term"
	"github.com/spf13/cobra"

	//"github.com/davecgh/go-spew/spew"

	"github.com/errordeveloper/kubegen/appmaker"
	"github.com/errordeveloper/kubegen/resources"
)

var (
	flavor       string
	image        string
	manifest     string
	outputFormat string
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "kubegen",
	}

	//var component = &cobra.Command{
	//	Use:  "kubegen ",
	//	RunE: generateComponent,
	//}

	var appStack = &cobra.Command{
		Use:  "stack",
		RunE: generateAppStack,
	}

	appStack.Flags().StringVarP(&manifest, "manifest", "", "Kubefile", "App manifest to use")
	appStack.Flags().StringVarP(&outputFormat, "output-format", "", "yaml", "App manifest to use")
	rootCmd.AddCommand(appStack)

	//component.Flags().StringVarP(&image, "image", "", "", "Image to use")
	//component.Flags().StringVarP(&flavor, "flavor", "", "default", "Flavor of generator to use")
	//rootCmd.AddCommand(component)

	var module = &cobra.Command{
		Use:  "module",
		RunE: moduleTest,
	}
	rootCmd.AddCommand(module)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func generateAppStack(cmd *cobra.Command, args []string) error {
	data, err := ioutil.ReadFile(manifest)
	if err != nil {
		return err
	}

	app, err := appmaker.NewFromHCL(data)
	if err != nil {
		return err
	}

	return encodeAndOutput(app)
}

/*
	if image == "" {
		return fmt.Errorf("Image must be specified")
	}

	app := &appmaker.App{
		GroupName: "test",
		Components: []appmaker.AppComponent{
			{
				Image:  image,
				Flavor: flavor,
			},
		},
	}

	return encodeAndOutput(app)
*/

func moduleTest(cmd *cobra.Command, args []string) error {
	var (
		data   []byte
		output string
		err    error
	)

	module, err := resources.NewResourceGroupFromPath(args[0])
	if err != nil {
		panic(err)
	}

	//spew.Dump(module)

	if data, err = module.EncodeListToPrettyJSON(); err != nil {
		return err
	}

	if term.IsTerminal(0) {
		veryPretty, err := highlight.Term(outputFormat, data)
		if err != nil {
			panic(err)
		}
		output = string(veryPretty)
	} else {
		output = string(data)
	}

	fmt.Println(output)

	return nil
}

func encodeAndOutput(app *appmaker.App) error {
	var (
		data   []byte
		output string
		err    error
	)

	switch outputFormat {
	case "yaml":
		if data, err = app.EncodeListToYAML(); err != nil {
			return err
		}
		data = append([]byte("---\n"), data...)
	case "json":
		if data, err = app.EncodeListToPrettyJSON(); err != nil {
			return err
		}
	}

	if term.IsTerminal(0) {
		veryPretty, err := highlight.Term(outputFormat, data)
		if err != nil {
			panic(err)
		}
		output = string(veryPretty)
	} else {
		output = string(data)
	}

	fmt.Println(output)

	return nil
}
