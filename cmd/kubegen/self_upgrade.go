package main

import (
	"fmt"

	"github.com/equinox-io/equinox"
	"github.com/spf13/cobra"
)

const equinoxAppID = "app_gAaxG6Siijv"

var equinoxPublicKey = []byte(`
-----BEGIN ECDSA PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE5tTFS5qFvYXh30syZDNwu4ldneAVIHzq
bd2Ua3/c7IRUyPpqm/bn4PIIXEI4/VbrYKcxNbGj75xPEA1eEzPoSstN+0V/2SSX
9GqA9J64sEsQOPmbllQ8tNsXeGH2z1ro
-----END ECDSA PUBLIC KEY-----
`)

var doUpgrade bool

var selfUpgradeCmd = &cobra.Command{
	Use:   "self-upgrade [--yes]",
	Short: "Upgrade kubegen to latest version",
	RunE:  selfUpgradeFn,
}

func init() {
	selfUpgradeCmd.Flags().BoolVarP(&doUpgrade, "yes", "y", false, "Confirm upgrade")
	selfUpgradeCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Printf("%s\n\nUsage:\n  kubegen %s\n", cmd.Short, cmd.Use)
		return nil
	})
	selfUpgradeCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		cmd.Usage()
	})
}

func selfUpgradeFn(cmd *cobra.Command, args []string) error {
	opts := equinox.Options{Channel: "latest"}
	fmt.Printf("Checking for updates on the %s release channel...\n", opts.Channel)
	if err := opts.SetPublicKeyPEM(equinoxPublicKey); err != nil {
		return err
	}

	resp, err := equinox.Check(equinoxAppID, opts)
	switch {
	case err == equinox.NotAvailableErr:
		return fmt.Errorf("No update available, already at the latest version!")
	case err != nil:
		return err
	}

	if doUpgrade {
		err = resp.Apply()
		if err != nil {
			return err
		}

		fmt.Printf("Upgraded to new version: %s!\n", resp.ReleaseVersion)
	} else {
		fmt.Println("New version available, please provide --yes to confirm upgrade.")
	}
	return nil
}
