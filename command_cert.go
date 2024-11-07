package main

import (
	"fmt"
	"fritzbox-client/api"
	"io"
	"os"
)

type certCommand struct {
	KeyPath         string `arg:"positional,required" placeholder:"path_key"`
	CertificatePath string `arg:"positional,required" placeholder:"path_cert"`
	KeyPass         string `arg:"positional" placeholder:"pass_key"`
}

func commandCert(options args) error {
	var err error

	var client api.FritzboxClient
	if client, err = api.NewClient(options.Hostname); err != nil {
		return err
	}

	fmt.Printf("Logging in to %s as %s… ", options.Hostname, options.Username)
	var sessionInfo api.SessionInfo
	if sessionInfo, err = client.Login(options.Username, options.Password); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}
	fmt.Println("Done.")

	fmt.Printf("Loading certificate from %s… ", options.Cert.CertificatePath)
	var certificate io.ReadCloser
	if certificate, err = os.Open(options.Cert.CertificatePath); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}
	fmt.Println("Done.")

	fmt.Printf("Loading key from %s… ", options.Cert.KeyPath)
	var key io.ReadCloser
	if key, err = os.Open(options.Cert.KeyPath); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}
	fmt.Println("Done.")

	fmt.Printf("Updating TLS certificate… ")
	var message string
	if message, err = client.UpdateTLSCertificate(sessionInfo.Sid, options.Cert.KeyPass, []io.ReadCloser{certificate, key}); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}
	fmt.Printf("Done: %s\n", message)

	return nil
}
