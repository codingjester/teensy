package main

import (
	"testing"
)

func TestFormatUrl(t *testing.T) {
	if FormatUrl("http", "testing.com", 9000, "test") != "http://testing.com:9000/test" {
		t.Error("Format URL did not equal http://testing.com:9000/test")
	}

	if FormatUrl("http", "testing.com", 80, "test") != "http://testing.com/test" {
		t.Error("Format URL did not equal http://testing.com/test")
	}
}
