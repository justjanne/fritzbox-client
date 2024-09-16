package main

import (
	"fmt"
	"fritzbox-client/api"
	"github.com/alexflint/go-arg"
	"io"
	"os"
)

type CertificateUpdateOptions struct {
	Hostname        string `arg:"--host,required"`
	Username        string `arg:"--user,required"`
	Password        string `arg:"--pass,required"`
	KeyPath         string `arg:"--key,required"`
	CertificatePath string `arg:"--cert,required"`
}

func updateCertificate(options CertificateUpdateOptions) error {
	var err error

	var client api.FritzboxClient
	if client, err = api.NewClient(options.Hostname); err != nil {
		return err
	}

	fmt.Printf("Logging in to %s as %s\n", options.Hostname, options.Username)
	var sessionInfo api.SessionInfo
	if sessionInfo, err = client.Login(options.Username, options.Password); err != nil {
		return err
	}

	fmt.Printf("Loading certificate from %s\n", options.CertificatePath)
	var certificate io.ReadCloser
	if certificate, err = os.Open(options.CertificatePath); err != nil {
		return err
	}

	fmt.Printf("Loading key from %s\n", options.KeyPath)
	var key io.ReadCloser
	if key, err = os.Open(options.KeyPath); err != nil {
		return err
	}

	fmt.Printf("Updating TLS certificate\n")
	var message string
	if message, err = client.UpdateTLSCertificate(sessionInfo.Sid, "", []io.ReadCloser{certificate, key}); err != nil {
		return err
	}
	fmt.Printf("Finished updating TLS certificate: %s\n", message)

	return nil
}

func main() {
	var args CertificateUpdateOptions
	arg.MustParse(&args)
	if err := updateCertificate(args); err != nil {
		panic(err)
	}
}
