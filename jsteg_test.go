package jsteg

import (
	"bytes"
	"image/jpeg"
	"os"
	"strings"
	"testing"
)

func loadTestImages(t *testing.T) []string {
	testdata, err := os.Open("testdata")
	if err != nil {
		t.Fatal(err)
	}
	names, err := testdata.Readdirnames(-1)
	if err != nil {
		t.Fatal(err)
	}
	return names
}

func TestHideReveal(t *testing.T) {
	for _, name := range loadTestImages(t) {
		// load test jpeg
		f, err := os.Open("testdata/" + name)
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
}

// Test that jpegs post-Hide can still be decoded normally
func TestHideDecode(t *testing.T) {
	for _, name := range loadTestImages(t) {
		// load test jpeg
		f, err := os.Open("testdata/" + name)
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

		// decode img
		_, err = jpeg.Decode(&buf)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// Progressive JPEGs are not supported
func TestRevealProgressive(t *testing.T) {
	for _, name := range loadTestImages(t) {
		if !strings.Contains(name, "progressive") {
			continue
		}
		// load test jpeg
		f, err := os.Open("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		_, err = Reveal(f)
		if _, ok := err.(UnsupportedError); !ok {
			t.Fatal("expected UnsupportedError, got", err)
		}
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
