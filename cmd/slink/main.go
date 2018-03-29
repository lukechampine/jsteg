package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ed25519"

	"github.com/lukechampine/flagg"
	"github.com/lukechampine/jsteg"
)

const magic = "slink"

func keypair(password string) (ed25519.PublicKey, ed25519.PrivateKey) {
	h := sha256.Sum256([]byte(password))
	pk, sk, _ := ed25519.GenerateKey(bytes.NewReader(h[:]))
	return pk, sk
}

func main() {
	log.SetFlags(0)

	flagg.Root.Usage = flagg.SimpleUsage(flagg.Root, `Usage: slink [command] [args]

Commands:
    claim img.jpg password            Embed a public key
    prove data password               Sign arbitrary data
    verify img.jpg data signature     Verify a signature

Use claim to embed your public key in an image. Later, you can
use prove to demonstrate that you control the private key paired
with the public key. Use verify to verify that a signature was
generated from the same keypair embedded in the image.
`)
	cmdClaim := flagg.New("claim", `Usage:
    slink claim img.jpg password
      Embed a public key (derived from password) in img.jpg
`)
	cmdProve := flagg.New("prove", `Usage:
    slink prove data password
      Sign arbitrary data using the private key derived from password
`)
	cmdVerify := flagg.New("verify", `Usage:
    slink verify img.jpg data signature
      Verify that data was signed by the same key embedded in img.jpg
`)
	cmd := flagg.Parse(flagg.Tree{
		Cmd: flagg.Root,
		Sub: []flagg.Tree{
			{Cmd: cmdClaim},
			{Cmd: cmdProve},
			{Cmd: cmdVerify},
		},
	})

	switch cmd {
	case cmdClaim:
		if cmd.NArg() != 2 {
			cmd.Usage()
			return
		}

		infile, err := os.Open(cmd.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer infile.Close()
		img, err := jpeg.Decode(infile)
		if err != nil {
			log.Fatal(err)
		}

		outPath := strings.TrimSuffix(cmd.Arg(0), filepath.Ext(cmd.Arg(0))) + ".claimed.jpg"
		outfile, err := os.Create(outPath)
		if err != nil {
			log.Fatal(err)
		}
		defer outfile.Close()

		data := make([]byte, len(magic)+ed25519.PublicKeySize)
		copy(data, magic)
		pk, _ := keypair(cmd.Arg(1))
		copy(data[len(magic):], pk[:])

		err = jsteg.Hide(outfile, img, data, nil)
		if err != nil {
			log.Fatal(err)
		}
		os.Stdout.WriteString("Wrote claimed jpeg to " + outPath + "\n")

	case cmdProve:
		if cmd.NArg() != 2 {
			cmd.Usage()
			return
		}

		_, sk := keypair(cmd.Arg(1))
		sig := ed25519.Sign(sk, []byte(cmd.Arg(0)))
		os.Stdout.WriteString(base64.StdEncoding.EncodeToString(sig) + "\n")

	case cmdVerify:
		if cmd.NArg() != 3 {
			cmd.Usage()
			return
		}

		infile, err := os.Open(cmd.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer infile.Close()
		data, err := jsteg.Reveal(infile)
		if err != nil {
			log.Fatal(err)
		}

		if len(data) < len(magic)+ed25519.PublicKeySize || string(data[:len(magic)]) != magic {
			log.Fatal("Image was not signed with slink")
		}
		pk := ed25519.PublicKey(data[len(magic):][:ed25519.PublicKeySize])
		sig, err := base64.StdEncoding.DecodeString(cmd.Arg(2))
		if err != nil {
			log.Fatal("Invalid signature")
		}
		if ed25519.Verify(pk, []byte(cmd.Arg(1)), sig) {
			log.Println("Verified OK")
		} else {
			log.Fatal("Bad signature")
		}

	default:
		flagg.Root.Usage()
	}
}
