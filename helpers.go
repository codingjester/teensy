package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

func ValidateURL(url string) bool {
	regex, err := regexp.Compile("http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*(),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+")
	if err != nil {
		return false
	}
	if regex.MatchString(url) {
		return true
	}
	return false
}

func FormatUrl(proto string, host string, port int, hash string) string {
	if port != 80 {
		host = fmt.Sprintf("%s:%d", host, port)
	}
	return fmt.Sprintf("%s://%s/%s", proto, host, hash)
}

// The decoding/encoding is simplistic but abstracted away
// so that we can (if we want) build something a bit more
// robust.

func DecodeHash(hash string) (int64, error) {
	id, err := strconv.ParseInt(hash, 36, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func EncodeHash(id int64) string {
	return strconv.FormatInt(id, 36)
}

func WriteJSON(w http.ResponseWriter, js []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
