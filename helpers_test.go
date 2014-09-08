package main

import (
	"net/url"
	"testing"
)

func TestValidateUrl(t *testing.T) {
	if !ValidateURL("http://testing.com/yeah") {
		t.Error("Validate error failed to parse http://testing.com/yeah")
	}

	if !ValidateURL("https://testing.com/yeah") {
		t.Error("Validate error failed to parse https://testing.com/yeah")
	}

	if ValidateURL("ftp://ftp.awesomesauce.com") {
		t.Error("Validate should not have allowed a ftp url to work.")
	}

	if ValidateURL("git://awesome.git.host.com") {
		t.Error("Validate should not have allowed a git url to work.")
	}

}

func TestFormatUrl(t *testing.T) {
	if FormatUrl("http", "testing.com", 9000, "test") != "http://testing.com:9000/test" {
		t.Error("Format URL did not equal http://testing.com:9000/test")
	}

	if FormatUrl("http", "testing.com", 80, "test") != "http://testing.com/test" {
		t.Error("Format URL did not equal http://testing.com/test")
	}
}

func TestDecodeHash(t *testing.T) {
	// Not a value we would be using or represented in our hash, but allowed in urls
	if _, err := DecodeHash("£"); err == nil {
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

func TestGetWithDefault(t *testing.T) {
	hash := url.Values{}
	hash.Set("offset", "25")
	if GetWithDefault(hash, "offset", "0") != "25" {
		t.Error("value did not return 25 for the offset")
	}

	if GetWithDefault(hash, "magic", "0") != "0" {
		t.Error("value did not return 0 for magic")
	}
}
