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

	"github.com/lukechampine/jsteg"
)

const magic = "slink"

func usage() {
	log.Fatalf(`Usage: %s [command] [args]

Commands:
    claim img.jpg password            Embed a public key
    prove data password               Sign arbitrary data
    verify img.jpg data signature     Verify a signature

Use claim to embed your public key in an image. Later, you can
use prove to demonstrate that you control the private key paired
with the public key. Use verify to verify that a signature was
generated from the same keypair embedded in the image.
`, os.Args[0])
}

func requireArgs(n int) {
	if len(os.Args[2:]) != n {
		usage()
	}
}

func keypair(password string) (ed25519.PublicKey, ed25519.PrivateKey) {
	h := sha256.Sum256([]byte(password))
	pk, sk, _ := ed25519.GenerateKey(bytes.NewReader(h[:]))
	return pk, sk
}

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	default:
		usage()

	case "claim":
		requireArgs(2)

		infile, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		defer infile.Close()
		img, err := jpeg.Decode(infile)
		if err != nil {
			log.Fatal(err)
		}

		outPath := strings.TrimSuffix(os.Args[2], filepath.Ext(os.Args[2])) + ".claimed.jpg"
		outfile, err := os.Create(outPath)
		if err != nil {
			log.Fatal(err)
		}
		defer outfile.Close()

		data := make([]byte, len(magic)+ed25519.PublicKeySize)
		copy(data, magic)
		pk, _ := keypair(os.Args[3])
		copy(data[len(magic):], pk[:])

		err = jsteg.Hide(outfile, img, data, nil)
		if err != nil {
			log.Fatal(err)
		}
		os.Stdout.WriteString("Wrote claimed jpeg to " + outPath + "\n")

	case "prove":
		requireArgs(2)

		_, sk := keypair(os.Args[3])
		sig := ed25519.Sign(sk, []byte(os.Args[2]))
		os.Stdout.WriteString(base64.StdEncoding.EncodeToString(sig) + "\n")

	case "verify":
		requireArgs(3)

		infile, err := os.Open(os.Args[2])
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
		sig, err := base64.StdEncoding.DecodeString(os.Args[4])
		if err != nil {
			log.Fatal("Invalid signature")
		}
		if ed25519.Verify(pk, []byte(os.Args[3]), sig) {
			log.Println("Verified OK")
		} else {
			log.Fatal("Bad signature")
		}
	}
}
