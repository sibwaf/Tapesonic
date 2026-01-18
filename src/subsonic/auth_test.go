package subsonic_test

import (
	"tapesonic/subsonic"
	"testing"
)

func TestEncodePassword_Ascii(t *testing.T) {
	encoded := subsonic.EncodePassword("qwerty123!")
	expected := "enc:71776572747931323321"

	if encoded != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, encoded)
	}
}

func TestDecodePassword_Ascii(t *testing.T) {
	decoded, err := subsonic.DecodePassword("enc:71776572747931323321")
	expected := "qwerty123!"

	if err != nil {
		t.Fatalf("Decode failed: %s", err.Error())
	}

	if decoded != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, decoded)
	}
}

func TestEncodePassword_Unicode(t *testing.T) {
	encoded := subsonic.EncodePassword("コワブンガ")
	expected := "enc:e382b3e383afe38396e383b3e382ac"

	if encoded != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, encoded)
	}
}

func TestDecodePassword_Unicode(t *testing.T) {
	decoded, err := subsonic.DecodePassword("enc:e382b3e383afe38396e383b3e382ac")
	expected := "コワブンガ"

	if err != nil {
		t.Fatalf("Decode failed: %s", err.Error())
	}

	if decoded != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, decoded)
	}
}
