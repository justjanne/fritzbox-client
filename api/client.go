package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type FritzboxClient struct {
	baseUrl    *url.URL
	httpClient *http.Client
}

func NewClient(baseUrl string) (FritzboxClient, error) {
	parsedUrl, err := url.Parse(baseUrl)
	if err != nil {
		return FritzboxClient{}, err
	}
	return FritzboxClient{
		baseUrl:    parsedUrl,
		httpClient: http.DefaultClient,
	}, nil
}

func (c *FritzboxClient) getSessionInfo() (SessionInfo, error) {
	resp, err := c.httpClient.Get(c.baseUrl.JoinPath("/login_sid.lua").String())
	if err != nil {
		return SessionInfo{}, err
	}
	var result SessionInfo
	err = xml.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (c *FritzboxClient) challengeResponseLogin(sid SessionID, username string, response string) (SessionInfo, error) {
	requestUrl := c.baseUrl.JoinPath("/login_sid.lua")
	query := requestUrl.Query()
	query.Set("sid", string(sid))
	query.Set("username", username)
	query.Set("response", response)
	requestUrl.RawQuery = query.Encode()
	resp, err := c.httpClient.Get(requestUrl.String())
	if err != nil {
		return SessionInfo{}, err
	}
	var result SessionInfo
	err = xml.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return SessionInfo{}, err
	}
	if result.Sid == "0000000000000000" {
		return SessionInfo{}, errors.New("login failed")
	}
	return result, nil
}

func (c *FritzboxClient) Login(username string, password string) (SessionInfo, error) {
	sessionInfo, err := c.getSessionInfo()
	if err != nil {
		return SessionInfo{}, err
	}
	response := fmt.Sprintf("%s-%s", sessionInfo.Challenge, challengeResponse(sessionInfo.Challenge, password))
	return c.challengeResponseLogin(sessionInfo.Sid, username, response)
}

func generateCertificateBody(multipartWriter *multipart.Writer, id SessionID, password string, files []io.ReadCloser) error {
	var err error
	var fieldWriter io.Writer
	if fieldWriter, err = multipartWriter.CreateFormField("sid"); err != nil {
		return err
	}
	if _, err = io.WriteString(fieldWriter, string(id)); err != nil {
		return err
	}
	if fieldWriter, err = multipartWriter.CreateFormField("BoxCertPassword"); err != nil {
		return err
	}
	if _, err = io.WriteString(fieldWriter, password); err != nil {
		return err
	}
	if fieldWriter, err = multipartWriter.CreateFormFile("BoxCertImportFile", "BoxCert.pem"); err != nil {
		return err
	}
	for _, file := range files {
		if _, err = io.Copy(fieldWriter, file); err != nil {
			return err
		}
	}
	if err = multipartWriter.Close(); err != nil {
		return err
	}
	return nil
}

func (c *FritzboxClient) UpdateTLSCertificate(id SessionID, password string, files []io.ReadCloser) (string, error) {
	buffer := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(buffer)

	var err error
	if err = generateCertificateBody(multipartWriter, id, password, files); err != nil {
		return "", err
	}

	requestUrl := c.baseUrl.JoinPath("/cgi-bin/firmwarecfg").String()
	var resp *http.Response
	if resp, err = c.httpClient.Post(requestUrl, multipartWriter.FormDataContentType(), buffer); err != nil {
		return "", err
	}

	var updateMessage string
	if updateMessage, err = parseUpdateResponse(resp); err != nil {
		return "", err
	}

	return updateMessage, nil
}

func decodeEmbeddedJson(reader io.Reader, v interface{}, prefix string, suffix string) error {
	scanner := bufio.NewScanner(reader)
	const maxCapacity int = 256 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) && strings.HasSuffix(line, suffix) {
			line = line[len(prefix) : len(line)-len(suffix)]
			if err := json.Unmarshal([]byte(line), &v); err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("could not find embedded json")
}

func (c *FritzboxClient) ListPhoneNumbers(id SessionID) ([]PhoneNumber, error) {
	requestUrl := c.baseUrl.JoinPath("/fon_num/fon_num_list.lua").String()

	values := url.Values{
		"xhr": {"1"},
		"sid": {string(id)},
	}

	var data []PhoneNumber
	var err error
	var resp *http.Response
	if resp, err = c.httpClient.PostForm(requestUrl, values); err != nil {
		return data, err
	}
	if err = decodeEmbeddedJson(resp.Body, &data, "var gFonNums = ", ";"); err != nil {
		return data, err
	}
	return data, nil
}

func (c *FritzboxClient) GetPhoneNumber(id SessionID, phoneNumberId string) (PhoneNumber, error) {
	requestUrl := c.baseUrl.JoinPath("/data.lua").String()

	params := fmt.Sprintf(
		"xhr=1&uid=%s&sid=%s&page=sip_edit",
		phoneNumberId,
		string(id),
	)
	var data PhoneNumber
	var err error
	var resp *http.Response
	if resp, err = c.httpClient.Post(requestUrl, "application/x-www-form-urlencoded", strings.NewReader(params)); err != nil {
		return data, err
	}

	if err = decodeEmbeddedJson(resp.Body, &data, "const g_fondata = [", "];"); err != nil {
		return data, err
	}
	return data, nil
}

func (c *FritzboxClient) DisableSIP(id SessionID, sipID string) error {
	requestUrl := c.baseUrl.JoinPath("/data.lua").String()

	values := url.Values{
		"xhr":   {"1"},
		"isnew": {"0"},
		"uid":   {sipID},
		"sid":   {string(id)},
		"page":  {"sip_edit"},
		"apply": {""},
	}

	var err error
	var resp *http.Response
	if resp, err = c.httpClient.PostForm(requestUrl, values); err != nil {
		return err
	}

	var updateResult UpdateResult
	if err = json.NewDecoder(resp.Body).Decode(&updateResult); err != nil {
		return err
	}
	if updateResult.Data.Apply == "valerror" {
		return errors.New(updateResult.Data.ValError.Alert)
	}
	if updateResult.Data.Apply != "ok" {
		return errors.New("unknown error while processing response")
	}
	return nil
}

func (c *FritzboxClient) EnableSIP(id SessionID, sipID string, provider string, areaCode string, localNumber string, username string, password string) error {
	requestUrl := c.baseUrl.JoinPath("/data.lua").String()

	values := url.Values{
		"xhr":            {"1"},
		"isnew":          {"0"},
		"sipactive":      {"on"},
		"sipprovider":    {provider},
		"numberinput1_1": {areaCode},
		"numberinput2_1": {localNumber},
		"username":       {username},
		"password":       {password},
		"uid":            {sipID},
		"sid":            {string(id)},
		"page":           {"sip_edit"},
		"apply":          {""},
	}

	var err error
	var resp *http.Response
	if resp, err = c.httpClient.PostForm(requestUrl, values); err != nil {
		return err
	}

	var updateResult UpdateResult
	if err = json.NewDecoder(resp.Body).Decode(&updateResult); err != nil {
		return err
	}
	if updateResult.Data.Apply == "valerror" {
		return errors.New(updateResult.Data.ValError.Alert)
	}
	if updateResult.Data.Apply != "ok" {
		return errors.New("unknown error while processing response")
	}
	return nil
}
