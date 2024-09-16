package api

import (
	"encoding/xml"
	"fmt"
)

type SessionID string

type SessionInfo struct {
	Sid       SessionID     `xml:"SID"`
	Challenge string        `xml:"Challenge"`
	BlockTime string        `xml:"BlockTime"`
	Rights    SessionAccess `xml:"Rights"`
	Users     []SessionUser `xml:"Users>User"`
}

type SessionUser struct {
	Last int    `xml:"last,attr,omitempty"`
	Name string `xml:",innerxml"`
}

type SessionAccess map[string]int

type sessionAccessMap struct {
	Name   []string `xml:"Name"`
	Access []int    `xml:"Access"`
}

func (m *SessionAccess) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	keyElement := xml.StartElement{Name: xml.Name{Local: "Name"}}
	valueElement := xml.StartElement{Name: xml.Name{Local: "Access"}}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range *m {
		if err := e.EncodeElement(key, keyElement); err != nil {
			return err
		}
		if err := e.EncodeElement(value, valueElement); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

func (m *SessionAccess) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	data := sessionAccessMap{}
	if err := d.DecodeElement(&data, &start); err != nil {
		return err
	}
	if len(data.Name) != len(data.Access) {
		return fmt.Errorf("unbalanced map entries")
	}
	*m = make(map[string]int)
	for i := 0; i < len(data.Name); i++ {
		(*m)[data.Name[i]] = data.Access[i]
	}
	return nil
}
