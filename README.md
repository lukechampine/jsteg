jsteg
-----

[![GoDoc](https://godoc.org/github.com/lukechampine/jsteg?status.svg)](https://godoc.org/github.com/lukechampine/jsteg)
[![Go Report Card](http://goreportcard.com/badge/github.com/lukechampine/jsteg)](https://goreportcard.com/report/github.com/lukechampine/jsteg)

```
go get github.com/lukechampine/jsteg
```

`jsteg` is a package for hiding data inside jpeg files, a technique known as
[steganography](https://en.wikipedia.org/wiki/steganography). This is accomplished
by copying each bit of the data into the least-significant bits of the image.
The amount of data that can be hidden depends on the filesize of the jpeg; it
takes about 10-14 bytes of jpeg to store each byte of the hidden data.

## Example

```go
// open an existing jpeg
f, _ := os.Open(filename)
img, _ := jpeg.Decode(f)

// add hidden data to it
out, _ := os.Create(outfilename)
data := []byte("my secret data")
jsteg.Hide(out, img, data, nil)

// read hidden data:
hidden, _ := jsteg.Reveal(out)
```

Note that the data is not demarcated in any way; the caller is responsible for
determining which bytes of `hidden` it cares about. The easiest way to do this
is to prepend the data with its length.

A `jsteg` command is also included, providing a simple wrapper around the
functions of this package. It can hide and reveal data in jpeg files and
supports input/output redirection. It automatically handles length prefixes
and uses a magic header to identify jpegs that were produced by `jsteg`.
Binaries can be found [here](https://github.com/lukechampine/jsteg/releases).

---

This package reuses a significant amount of code from the image/jpeg package.
The BSD-style license that governs the use of that code can be found in the
`go_LICENSE` file.
