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

func TestDecodeHash(t *testing.T) {
	_, err := DecodeHash("£") // Not a value we would be using, but allowed in urls
	if err == nil {
		t.Errorf("Value was decoded incorrectly")
	}

	value, err := DecodeHash("1")
	if err != nil {
		t.Errorf("Value was decoded incorrectly with error: %s", err.Error())
	}

	if value != 1 {
		t.Error("value did not equal 1, it equaled %d", value)
	}
}

func TestEncodeHash(t *testing.T) {
	if EncodeHash(1) != "1" {
		t.Error("value did not equal 1")
	}

	if EncodeHash(1234567890) != "kf12oi" {
		t.Error("value 1234567890 did not equal kf12oi")
	}
}
