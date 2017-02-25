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

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update kubegen to latest version",
	RunE:  updateFn,
}

func updateFn(cmd *cobra.Command, args []string) error {
	opts := equinox.Options{Channel: "latest"}
	fmt.Printf("Checking for updates on the %s release channel...\n", opts.Channel)
	if err := opts.SetPublicKeyPEM(equinoxPublicKey); err != nil {
		return err
	}

	resp, err := equinox.Check(equinoxAppID, opts)
	switch {
	case err == equinox.NotAvailableErr:
		fmt.Println("No update available, already at the latest version!")
		return nil
	case err != nil:
		return err
	}

	// TODO can ask user for confirmation and print binary path etc
	err = resp.Apply()
	if err != nil {
		return err
	}

	fmt.Printf("Updated to new version: %s!\n", resp.ReleaseVersion)
	return nil
}
