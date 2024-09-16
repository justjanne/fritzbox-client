package api

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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
	return result, err
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
