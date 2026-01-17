// Package auth: deals with the authorization and authentication of the http server. For example, it issues JWT and access tokens.
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authorisationHeader, ok := headers["Authorization"]
	if !ok {
		return "", errors.New("no authorization header found")
	} else if len(authorisationHeader) == 0 {
		return "", errors.New("authorization header is empty")
	}

	apiKey, ok := strings.CutPrefix(authorisationHeader[0], "ApiKey ")
	if !ok {
		return "", fmt.Errorf("authorization header doesn't begin with 'ApiKey ': %v", authorisationHeader[0])
	}

	return strings.TrimSpace(apiKey), nil
}
