package main

import (
	"encoding/binary"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/lukechampine/jsteg"
)

func usage() {
	log.Fatalf(`Usage:
    %[1]s hide in.jpg [FILE] [out.jpg]
      Hide FILE (or stdin) in in.jpg, writing the result to out.jpg (or stdout)
    %[1]s reveal in.jpg [FILE]
      Write the hidden contents of in.jpg to FILE (or stdout)
`, os.Args[0])
}

const magic = "jsteg"

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "hide":
		var in io.Reader
		var out io.Writer
		switch len(os.Args) {
		// stdin and stdout
		case 3:
			in, out = os.Stdin, os.Stdout

		// either stdin and outfile or infile and stdout
		case 4:
			// detect whether we have stdin
			// (not perfect; doesn't work with e.g. /dev/zero)
			stat, _ := os.Stdin.Stat()
			haveStdin := (stat.Mode() & os.ModeCharDevice) == 0
			if haveStdin {
				fout, err := os.Create(os.Args[3])
				if err != nil {
					log.Fatal("could not create output file:", err)
				}
				defer fout.Close()
				in, out = os.Stdin, fout
			} else {
				fin, err := os.Open(os.Args[3])
				if err != nil {
					log.Fatal("could not open file:", err)
				}
				defer fin.Close()
				in, out = fin, os.Stdout
			}

		// infile and outfile
		case 5:
			fin, err := os.Open(os.Args[3])
			if err != nil {
				log.Fatal("could not open file:", err)
			}
			defer fin.Close()
			fout, err := os.Create(os.Args[4])
			if err != nil {
				log.Fatal("could not create output file:", err)
			}
			defer fout.Close()
			in, out = fin, fout

		default:
			usage()
		}

		injpg, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatal("could not open jpeg:", err)
		}
		img, err := jpeg.Decode(injpg)
		if err != nil {
			log.Fatal("could not decode jpeg:", err)
		}

		text, err := ioutil.ReadAll(in)
		if err != nil {
			log.Fatal("could not read input:", err)
		}

		data := make([]byte, 9+len(text))
		copy(data[:5], magic)
		binary.LittleEndian.PutUint32(data[5:9], uint32(len(text)))
		copy(data[9:], text)

		err = jsteg.Hide(out, img, data, nil)
		if err != nil {
			log.Fatal("could not write output file:", err)
		}

	case "reveal":
		var out io.Writer
		switch len(os.Args) {
		// stdout
		case 3:
			out = os.Stdout

		// outfile
		case 4:
			fout, err := os.Create(os.Args[3])
			if err != nil {
				log.Fatal("could not create output file:", err)
			}
			defer fout.Close()
			out = fout

		default:
			usage()
		}

		injpg, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatal("could not open file:", err)
		}
		defer injpg.Close()

		data, err := jsteg.Reveal(injpg)
		if err != nil {
			log.Fatal("could not decode jpeg:", err)
		}
		if len(data) < 5 || string(data[:5]) != magic {
			log.Fatal("jpeg does not contain hidden data")
		}
		n := binary.LittleEndian.Uint32(data[5:9])
		if n > uint32(len(data)) {
			log.Fatal("hidden data is malformed")
		}

		text := data[9:][:n]
		if _, err := out.Write(text); err != nil {
			log.Fatal("could not write hidden data:", err)
		}

	default:
		usage()
	}
}
