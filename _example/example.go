// _example/example.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/kenshaw/ffcookies"
	//_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

func main() {
	profile := flag.String("profile", "", "profile")
	host := flag.String("host", "", "host")
	flag.Parse()
	if err := run(context.Background(), *profile, *host); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, profile, host string) error {
	cookies, err := ffcookies.ReadContext(ctx, profile, host)
	if err != nil {
		return err
	}
	for i, cookie := range cookies {
		fmt.Printf("%d:\n", i)
		fmt.Printf("  domain: %s\n", cookie.Domain)
		fmt.Printf("  name: %q\n", cookie.Name)
		fmt.Printf("  expires: %q\n", cookie.Expires)
		fmt.Printf("  path: %q\n", cookie.Path)
		fmt.Printf("  value: %q\n", cookie.Value)
	}
	return nil
}
