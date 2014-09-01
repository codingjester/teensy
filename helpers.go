package main

import (
	"fmt"
)

func FormatUrl(proto string, host string, port int, hash string) string {
	if port != 80 {
		host = fmt.Sprintf("%s:%d", host, port)
	}
	return fmt.Sprintf("%s://%s/%s", proto, host, hash)
}
