package main

import (
	"fmt"
	"fritzbox-client/api"
	"slices"
	"strings"
)

type sipCommand struct {
	Task string   `arg:"positional,required" placeholder:"<connect|disconnect|reconnect>"`
	Ids  []string `arg:"positional" placeholder:"uid"`
}

func commandSip(options args) error {
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

	fmt.Print("Querying list of phone numbers… ")
	phoneNumbers, err := client.ListPhoneNumbers(sessionInfo.Sid)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}
	fmt.Printf("Found %d numbers.\n", len(phoneNumbers))
	for _, phoneNumber := range phoneNumbers {
		if phoneNumber.Type != "sip" {
			continue
		}
		if len(options.Sip.Ids) > 0 && !slices.Contains(options.Sip.Ids, phoneNumber.Uid) {
			continue
		}
		var data api.PhoneNumber
		fmt.Printf("Loading configuration for phone number %s… ", phoneNumber.Number)
		data, err = client.GetPhoneNumber(sessionInfo.Sid, phoneNumber.Uid)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return err
		}
		fmt.Println("Done.")
		if strings.EqualFold(options.Sip.Task, "disconnect") || strings.EqualFold(options.Sip.Task, "reconnect") {
			fmt.Printf("Disabling SIP Number %s… ", phoneNumber.Number)
			if err = client.DisableSIP(sessionInfo.Sid, phoneNumber.Uid); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return err
			}
			fmt.Println("Done.")
		}
		if strings.EqualFold(options.Sip.Task, "connect") || strings.EqualFold(options.Sip.Task, "reconnect") {
			fmt.Printf("Enabling SIP Number %s… ", phoneNumber.Number)
			if err = client.EnableSIP(sessionInfo.Sid, phoneNumber.Uid, data.ProviderId, data.AreaCode, data.LocalNumber, data.Sip.Username, data.Sip.Password); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return err
			}
			fmt.Println("Done.")
		}
	}
	return nil
}
