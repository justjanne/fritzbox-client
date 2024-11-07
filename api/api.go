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

type UpdateResult struct {
	Data DataResult `json:"data"`
}

type DataResult struct {
	Apply    string         `json:"apply"`
	Redirect RedirectResult `json:"redirect,omitempty"`
	ValError ValErrorResult `json:"valerror,omitempty"`
}

type RedirectResult struct {
	Back bool `json:"back"`
}

type ValErrorResult struct {
	Ok     bool     `json:"ok"`
	ToMark []string `json:"tomark"`
	Result string   `json:"result"`
	Alert  string   `json:"alert"`
}

type PhoneNumber struct {
	Number           string          `json:"number"`
	OutboundProxy    string          `json:"outboundproxy"`
	Active           bool            `json:"active"`
	ProviderName     string          `json:"providername"`
	CountTrunk       int             `json:"count_trunk"`
	Deletable        bool            `json:"deletable"`
	MsnNumber        string          `json:"msnnum"`
	AreaCode         string          `json:"number1"`
	LocalNumber      string          `json:"number2"`
	Sip              SipData         `json:"sip"`
	Type             string          `json:"type"`
	Mode             string          `json:"mode"`
	Id               string          `json:"id"`
	WebUiTrunkId     string          `json:"webui_trunk:id"`
	Registrar        string          `json:"registrar"`
	TelConfig        TelephoneConfig `json:"telcfg"`
	Uid              string          `json:"uid"`
	TelConfigId      string          `json:"telcfg_id"`
	ParentProviderId string          `json:"parentprovider_id"`
	ProviderId       string          `json:"provider_id"`
	Registered       bool            `json:"registered"`
	GuiReadonly      string          `json:"gui_readonly"`
	Name             string          `json:"name"`
}

type SipData struct {
	OutboundProxyWithoutRouteHeader string `json:"outboundproxy_without_route_header"`
	ProviderName                    string `json:"providername"`
	MwiSupported                    string `json:"mwi_supported"`
	ProtocolPrefer                  string `json:"protocolprefer"`
	Username                        string `json:"username"`
	Trunk                           string `json:"Trunk"`
	UseInternatCallingNumber        string `json:"use_internat_calling_numb"`
	DoNotRegister                   string `json:"do_not_register"`
	Reception                       string `json:"Reception"`
	ExtensionLength                 string `json:"ExtensionLength"`
	TransportType                   string `json:"transport_type"`
	Registrar                       string `json:"registrar"`
	ClirType                        string `json:"clirtype"`
	G726ViaRfc3551                  string `json:"g726_via_rfc3551_"`
	ShowProtocolPrefer              bool   `json:"showprotocolprefer"`
	CallDeflection                  string `json:"call_deflection"`
	OutboundProxy                   string `json:"outboundproxy"`
	VoipProviderListId              string `json:"voip_providerlist_id"`
	DisplayName                     string `json:"displayname"`
	EncryptionEnabled               string `json:"encryption_enabled"`
	CryptoAvpMode                   string `json:"crypto_avp_mode"`
	SrtpSupported                   string `json:"srtp_supported"`
	TxPacketSizeInMs                string `json:"tx_packetsize_in_ms"`
	Node                            string `json:"_node"`
	NoRegisterFetch                 string `json:"no_register_fetch"`
	CcbsSupported                   string `json:"ccbs_supported"`
	ReadPAssertedIdentityHeader     string `json:"read_p_asserted_identity_header"`
	ID                              string `json:"ID"`
	DTMFConfig                      string `json:"dtmfcfg"`
	OriginStunServer                string `json:"origin_stunserver"`
	StunServer                      string `json:"stunserver"`
	RouteAlwaysOverInternet         string `json:"route_always_over_internet"`
	OriginOutboundProxy             string `json:"origin_outboundproxy"`
	OriginRegistrar                 string `json:"origin_registrar"`
	OriginUsername                  string `json:"origin_username"`
	Mode                            string `json:"mode"`
	WebUiTrunkId                    string `json:"webui_trunk_id"`
	ClipNsType                      string `json:"clipnstype"`
	DdiType                         string `json:"dditype"`
	Password                        string `json:"password"`
	VoipOverMobile                  string `json:"voip_over_mobile"`
	SippingInterval                 string `json:"sipping_interval"`
	Registered                      string `json:"registered"`
	AuthnameNeeded                  string `json:"authname_needed"`
	Authname                        string `json:"authname"`
	GuiReadonly                     string `json:"gui_readonly"`
	Activated                       string `json:"activated"`
}

type TelephoneConfig struct {
	RegistryType    string `json:"RegistryType"`
	AKN             string `json:"AKN"`
	EmergencyRule   string `json:"EmergencyRule"`
	KeepLKZPrefix   string `json:"KeepLKZPrefix"`
	KeepOKZPrefix   string `json:"KeepOKZPrefix"`
	Suffix          string `json:"Suffix"`
	ClipNoScreening string `json:"ClipNoScreening"`
	AlternatePrefix string `json:"AlternatePrefix"`
	UseOKZ          string `json:"UseOKZ"`
	UseLKZ          string `json:"UseLKZ"`
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
