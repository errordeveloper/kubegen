package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/d4l3k/go-highlight"
	"github.com/docker/docker/pkg/term"
	"github.com/spf13/cobra"

	"github.com/errordeveloper/kubegen/appmaker"
)

var (
	flavor   string
	image    string
	manifest string
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
	rootCmd.AddCommand(appStack)

	//component.Flags().StringVarP(&image, "image", "", "", "Image to use")
	//component.Flags().StringVarP(&flavor, "flavor", "", "default", "Flavor of generator to use")
	//rootCmd.AddCommand(component)

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

func encodeAndOutput(app *appmaker.App) error {
	data, err := app.EncodeListToYAML()
	if err != nil {
		return err
	}

	data = append([]byte("---\n"), data...)
	var output string
	if term.IsTerminal(0) {
		pretty, err := highlight.Term("yaml", data)
		if err != nil {
			panic(err)
		}
		output = string(pretty)
	} else {
		output = string(data)
	}
	fmt.Println(output)

	return nil
}
