// Package ffookies provides a quick way to read cookies from a firefox browser
// profile.
package ffcookies

//go:generate ./gen.sh

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kenshaw/ffcookies/models"
	"golang.org/x/net/publicsuffix"
)

// DefaultOpenParams are the default open parameters to use.
var DefaultOpenParams = "?nolock=1&immutable=1&mode=ro"

/*

sq:/home/ken/cookies.sqlite=> \d moz_cookies
                   TABLE "moz_cookies"
           Name            |  Type   | Nullable | Default
---------------------------+---------+----------+---------
 creationTime              | INTEGER | "YES"    |
 expiry                    | INTEGER | "YES"    |
 host                      | TEXT    | "YES"    |
 id                        | INTEGER | "YES"    |
 inBrowserElement          | INTEGER | "YES"    | 0
 isHttpOnly                | INTEGER | "YES"    |
 isPartitionedAttributeSet | INTEGER | "YES"    | 0
 isSecure                  | INTEGER | "YES"    |
 lastAccessed              | INTEGER | "YES"    |
 name                      | TEXT    | "YES"    |
 originAttributes          | TEXT    | "NO"     | ''
 path                      | TEXT    | "YES"    |
 rawSameSite               | INTEGER | "YES"    | 0
 sameSite                  | INTEGER | "YES"    | 0
 schemeMap                 | INTEGER | "YES"    | 0
 value                     | TEXT    | "YES"    |
Indexes:
  "sqlite_autoindex_moz_cookies_1" UNIQUE,  (name, host, path, originAttributes)

*/

// ReadFileContext reads the cookies from the provided sqlite3 file on disk.
func ReadFileContext(ctx context.Context, file, host string) ([]*http.Cookie, error) {
	// check sqlite driver
	driver := driverName()
	if driver == "" {
		return nil, errors.New("code using ffookies must import a sqlite driver!")
	}
	// open database
	db, err := sql.Open(driver, file)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// query func and params
	f := models.Cookies
	if host != "" {
		f, host = models.CookiesLikeHost, "%"+strings.TrimPrefix(host, "%")
	}
	// exec and convert
	res, err := f(ctx, db, host)
	if err != nil {
		return nil, err
	}
	return models.Convert(res), nil
}

// ReadFile reads the cookies from the provided sqlite3 file on disk.
func ReadFile(file, host string) ([]*http.Cookie, error) {
	return ReadFileContext(context.Background(), file, host)
}

// ReadContext reads the cookies for the provided Firefox profile name, or the
// default Firefox profile.
func ReadContext(ctx context.Context, profile, host string) ([]*http.Cookie, error) {
	profileDir := profileDir()
	if profileDir == "" {
		return nil, errors.New("cannot determine the firefox profile directory")
	}
	cookiePath, err := cookiePath(profileDir, profile)
	if err != nil {
		return nil, err
	}
	return ReadFileContext(ctx, "file:"+cookiePath+DefaultOpenParams, host)
}

// Read reads the cookies for the provided Firefox profile name.
func Read(profile, host string) ([]*http.Cookie, error) {
	return ReadContext(context.Background(), profile, host)
}

// Jar builds a cookie jar for the url from provided cookies.
func Jar(u *url.URL, cookies ...*http.Cookie) (http.CookieJar, error) {
	// build jar
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}
	jar.SetCookies(u, cookies)
	return jar, nil
}

// ReadJarContext reads the cookies from the provided sqlite3 file for the provided
// url into a cookie jar usable with http.Client.
func ReadJarContext(ctx context.Context, profile, urlstr string) (http.CookieJar, error) {
	// read cookies
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https", "ws", "wss":
	default:
		return nil, fmt.Errorf("invalid url scheme %q", u.Scheme)
	}
	cookies, err := ReadContext(ctx, profile, u.Host)
	if err != nil {
		return nil, err
	}
	return Jar(u, cookies...)
}

// ReadJar reads the cookies from the provided sqlite3 file for the provided
// url into a cookie jar usable with http.Client.
func ReadJar(profile, urlstr string) (http.CookieJar, error) {
	return ReadJarContext(context.Background(), profile, urlstr)
}

// ReadJarFilteredContext reads the cookies from the provided sqlite3 file for
// the provided url into a cookie jar (usable with http.Client) consisting of
// cookies passed through filter func f.
func ReadJarFilteredContext(ctx context.Context, profile, urlstr string, f func(*http.Cookie) bool) (http.CookieJar, error) {
	// read cookies
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https", "ws", "wss":
	default:
		return nil, fmt.Errorf("invalid url scheme %q", u.Scheme)
	}
	cookies, err := ReadContext(ctx, profile, u.Host)
	if err != nil {
		return nil, err
	}
	// filter
	var c []*http.Cookie
	for _, cookie := range cookies {
		if f(cookie) {
			c = append(c, cookie)
		}
	}
	return Jar(u, c...)
}

// ReadJarFiltered reads the cookies from the provided sqlite3 file for the
// provided url into a cookie jar (usable with http.Client) consisting of
// cookies passed through filter func f.
func ReadJarFiltered(ctx context.Context, profile, urlstr string, f func(*http.Cookie) bool) (http.CookieJar, error) {
	return ReadJarFilteredContext(context.Background(), profile, urlstr, f)
}

// driverName returns the first sqlite3 driver name it encounters.
func driverName() string {
	for _, n := range sql.Drivers() {
		switch n {
		case "sqlite3", "sqlite":
			return n
		}
	}
	return ""
}

// profileDir returns the base profile directory for firefox.
func profileDir() string {
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".mozilla", "firefox")
	}
	return ""
}

// cookiePath determines the cookie file path.
func cookiePath(dir, profile string) (string, error) {
	if profile == "" {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			if name := entry.Name(); entry.IsDir() && strings.HasSuffix(name, ".default-release") {
				dir = filepath.Join(dir, name)
				break
			}
		}
	}
	return filepath.Join(dir, profile, "cookies.sqlite"), nil
}
