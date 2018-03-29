package main

import (
	"encoding/binary"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/lukechampine/flagg"

	"github.com/lukechampine/jsteg"
)

const magic = "jsteg"

func main() {
	log.SetFlags(0)

	flagg.Root.Usage = flagg.SimpleUsage(flagg.Root, `Usage: jsteg [command] [args]

Commands:
    jsteg hide in.jpg [FILE] [out.jpg]
    jsteg reveal in.jpg [FILE]
`)
	cmdHide := flagg.New("hide", `Usage:
    jsteg hide in.jpg [FILE] [out.jpg]
      Hide FILE (or stdin) in in.jpg, writing the result to out.jpg (or stdout)
`)
	cmdReveal := flagg.New("reveal", `Usage:
    jsteg reveal in.jpg [FILE]
      Write the hidden contents of in.jpg to FILE (or stdout)
`)
	cmd := flagg.Parse(flagg.Tree{
		Cmd: flagg.Root,
		Sub: []flagg.Tree{
			{Cmd: cmdHide},
			{Cmd: cmdReveal},
		},
	})

	switch cmd {
	case cmdHide:
		var in io.Reader
		var out io.Writer
		switch cmd.NArg() {
		// stdin and stdout
		case 1:
			in, out = os.Stdin, os.Stdout

		// either stdin and outfile or infile and stdout
		case 2:
			// detect whether we have stdin
			// (not perfect; doesn't work with e.g. /dev/zero)
			stat, _ := os.Stdin.Stat()
			haveStdin := (stat.Mode() & os.ModeCharDevice) == 0
			if haveStdin {
				fout, err := os.Create(cmd.Arg(1))
				if err != nil {
					log.Fatalln("could not create output file:", err)
				}
				defer fout.Close()
				in, out = os.Stdin, fout
			} else {
				fin, err := os.Open(cmd.Arg(1))
				if err != nil {
					log.Fatalln("could not open file:", err)
				}
				defer fin.Close()
				in, out = fin, os.Stdout
			}

		// infile and outfile
		case 3:
			fin, err := os.Open(cmd.Arg(1))
			if err != nil {
				log.Fatalln("could not open file:", err)
			}
			defer fin.Close()
			fout, err := os.Create(cmd.Arg(2))
			if err != nil {
				log.Fatalln("could not create output file:", err)
			}
			defer fout.Close()
			in, out = fin, fout

		default:
			cmdHide.Usage()
			return
		}

		injpg, err := os.Open(cmd.Arg(0))
		if err != nil {
			log.Fatalln("could not open jpeg:", err)
		}
		img, err := jpeg.Decode(injpg)
		if err != nil {
			log.Fatalln("could not decode jpeg:", err)
		}

		text, err := ioutil.ReadAll(in)
		if err != nil {
			log.Fatalln("could not read input:", err)
		}

		data := make([]byte, 9+len(text))
		copy(data[:5], magic)
		binary.LittleEndian.PutUint32(data[5:9], uint32(len(text)))
		copy(data[9:], text)

		err = jsteg.Hide(out, img, data, nil)
		if err != nil {
			log.Fatalln("could not write output file:", err)
		}

	case cmdReveal:
		var out io.Writer
		switch cmd.NArg() {
		// stdout
		case 1:
			out = os.Stdout

		// outfile
		case 2:
			fout, err := os.Create(cmd.Arg(1))
			if err != nil {
				log.Fatalln("could not create output file:", err)
			}
			defer fout.Close()
			out = fout

		default:
			cmdReveal.Usage()
			return
		}

		injpg, err := os.Open(cmd.Arg(0))
		if err != nil {
			log.Fatalln("could not open file:", err)
		}
		defer injpg.Close()

		data, err := jsteg.Reveal(injpg)
		if err != nil {
			log.Fatalln("could not decode jpeg:", err)
		}
		if len(data) < 5 || string(data[:5]) != magic {
			log.Fatalln("jpeg does not contain hidden data")
		}
		n := binary.LittleEndian.Uint32(data[5:9])
		if n > uint32(len(data)) {
			log.Fatalln("hidden data is malformed")
		}

		text := data[9:][:n]
		if _, err := out.Write(text); err != nil {
			log.Fatalln("could not write hidden data:", err)
		}

	default:
		flagg.Root.Usage()
	}
}
