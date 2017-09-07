package jsteg

import (
	"bytes"
	"image/jpeg"
	"os"
	"testing"
)

func TestHideReveal(t *testing.T) {
	// load test jpeg
	f, err := os.Open("testdata/video-001.jpeg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	// hide data in img
	var buf bytes.Buffer
	data := []byte("foo bar baz quux")
	err = Hide(&buf, img, data, nil)
	if err != nil {
		t.Fatal(err)
	}

	// reveal data
	revealed, err := Reveal(&buf)
	if err != nil {
		t.Fatal(err)
	}
	revealed = revealed[:len(data)]
	if !bytes.Equal(data, revealed) {
		t.Fatal("revealed bytes do not match original")
	}
}

func TestTooSmall(t *testing.T) {
	// load test jpeg
	f, err := os.Open("testdata/video-001.jpeg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	// attempt hide data in img
	var buf bytes.Buffer
	data := make([]byte, 10e6)
	err = Hide(&buf, img, data, nil)
	if err != ErrTooSmall {
		t.Fatal("expected ErrTooSmall, got", err)
	}
}
