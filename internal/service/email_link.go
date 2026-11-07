package service

import (
	"errors"
	"net/url"
	"os"
	"strings"
)

func emailLink(path string, query url.Values) (string, error) {
	address := strings.TrimSpace(os.Getenv("EMAIL_APP_URL"))
	if address == "" {
		address = strings.TrimSpace(os.Getenv("APP_URL"))
	}
	if address == "" {
		address = "https://medifinder.online"
	}

	base, err := url.Parse(address)
	if err != nil || base.Hostname() == "" || base.User != nil ||
		(base.Scheme != "https" && !(base.Scheme == "http" && (base.Hostname() == "localhost" || base.Hostname() == "127.0.0.1"))) {
		return "", errors.New("email frontend URL must use HTTPS (HTTP is allowed on localhost)")
	}

	base.Path = strings.TrimRight(base.Path, "/") + path
	base.RawPath = ""
	base.RawQuery = query.Encode()
	base.ForceQuery = false
	base.Fragment = ""
	return base.String(), nil
}
