package service

import (
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
)

const publicEmailAppURL = "https://medifinder.online"

func emailLink(path string, query url.Values) (string, error) {
	address := strings.TrimSpace(os.Getenv("EMAIL_APP_URL"))
	if address == "" {
		address = strings.TrimSpace(os.Getenv("APP_URL"))
	}
	if address == "" {
		address = publicEmailAppURL
	}

	base, err := url.Parse(address)
	if err != nil || base.Hostname() == "" || base.User != nil {
		return "", errors.New("invalid email frontend URL")
	}

	host := strings.TrimSuffix(strings.ToLower(base.Hostname()), ".")
	loopback := host == "localhost" || net.ParseIP(host).IsLoopback()
	if loopback && os.Getenv("RAILWAY_DEPLOYMENT_ID") != "" {
		// Email recipients cannot reach the deployed service's loopback address.
		base = &url.URL{Scheme: "https", Host: strings.TrimPrefix(publicEmailAppURL, "https://")}
	} else if base.Scheme != "https" && !(base.Scheme == "http" && loopback) {
		return "", errors.New("email frontend URL must use HTTPS (HTTP is allowed on localhost)")
	}

	base.Path = strings.TrimRight(base.Path, "/") + path
	base.RawPath = ""
	base.RawQuery = query.Encode()
	base.ForceQuery = false
	base.Fragment = ""
	return base.String(), nil
}
