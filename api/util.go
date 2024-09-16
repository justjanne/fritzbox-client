package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func challengeResponse(challenge string, password string) string {
	data := fmt.Sprintf("%s-%s", challenge, password)
	enc := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	hasher := md5.New()
	writer := transform.NewWriter(hasher, enc)
	_, _ = io.WriteString(writer, data)
	return hex.EncodeToString(hasher.Sum(nil))
}

func innerText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}

	result := strings.Builder{}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		_, _ = result.WriteString(innerText(child))
	}
	return result.String()
}

var updateMessageSelector = cascadia.MustCompile("form[name=mainform] > p")
var javascriptSelector = cascadia.MustCompile("script[type=module]")
var jsFunctionSelector = regexp.MustCompile(`postUpload\.redirect\(([0-9]*)\);`)

func parseUpdateResponse(resp *http.Response) (string, error) {
	var err error
	var document *html.Node
	if document, err = html.Parse(resp.Body); err != nil {
		return "", err
	}

	updateMessageNode := cascadia.Query(document, updateMessageSelector)
	if updateMessageNode == nil {
		return "", fmt.Errorf("unable to find update message in document")
	}
	updateMessage := strings.TrimSpace(innerText(updateMessageNode))

	javascriptNode := cascadia.Query(document, javascriptSelector)
	if javascriptNode == nil {
		return "", fmt.Errorf("unable to find post-update JS in document")
	}
	matches := jsFunctionSelector.FindStringSubmatch(innerText(javascriptNode))
	if len(matches) != 2 {
		return "", fmt.Errorf("unable to parse post-update JS")
	}
	redirectDelay := matches[1]
	if redirectDelay != "" {
		return "", fmt.Errorf("failed to update TLS certificate: %s", updateMessage)
	}

	return updateMessage, nil
}
