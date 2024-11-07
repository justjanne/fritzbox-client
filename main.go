package main

import (
	"errors"
	"github.com/alexflint/go-arg"
	"log"
	"os"
	"strings"
)

type args struct {
	Hostname string       `arg:"--host,required" placeholder:"host"`
	Username string       `arg:"--user,required" placeholder:"user"`
	Password string       `arg:"--pass,required" placeholder:"pass"`
	Sip      *sipCommand  `arg:"subcommand:sip"`
	Cert     *certCommand `arg:"subcommand:cert"`
}

func main() {
	var args args
	p, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		log.Fatalf("there was an error in the definition of the Go struct: %v", err)
	}

	err = p.Parse(os.Args[1:])
	switch {
	case errors.Is(err, arg.ErrHelp):
		_ = p.WriteHelpForSubcommand(os.Stdout, p.SubcommandNames()...)
		os.Exit(0)
	case err != nil:
		_ = p.WriteUsageForSubcommand(os.Stdout, p.SubcommandNames()...)
		os.Exit(64)
	}

	if args.Sip != nil && (strings.EqualFold(args.Sip.Task, "reconnect") || strings.EqualFold(args.Sip.Task, "disconnect") || strings.EqualFold(args.Sip.Task, "connect")) {
		if err := commandSip(args); err != nil {
			os.Exit(1)
		}
	} else if args.Cert != nil {
		if err := commandCert(args); err != nil {
			os.Exit(1)
		}
	} else {
		_ = p.WriteHelpForSubcommand(os.Stdout, p.SubcommandNames()...)
		os.Exit(64)
	}
}
