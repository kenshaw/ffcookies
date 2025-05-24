package models

import (
	"net/http"
	"time"
)

// Convert converts a slice of Cookie to http.Cookie.
func Convert(res []*Cookie) []*http.Cookie {
	var cookies []*http.Cookie
	for _, c := range res {
		cookies = append(cookies, &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Host,
			Expires:  time.Unix(c.Expiry, 0),
			Secure:   c.IsSecure,
			HttpOnly: c.IsHTTPOnly,
			// SameSite: c.SameSite,
		})
	}
	return cookies
}
