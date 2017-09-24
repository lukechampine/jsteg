slink
-----

`slink` is a tool for proving pseudo-ownership of JPEG files. It is intended
as a way for content creators to invisibly watermark their images.

`slink` embeds a public key in a JPEG file, and makes it easy to sign
arbitrary data with the corresponding private key, producing a signature that
`slink` can also verify. Keypairs are derived from password strings. The
intended scenario is:

1. Content creator creates a JPEG file and embeds their public key in it
2. Content creator posts JPEG publicly
3. Someone else claims to have created the JPEG
4. Content creator proves ownership by signing arbitrary data with their key

To strengthen the claim, the content creator should sign a piece of data
supplied by the challenger. Furthermore, they can prove that the image was
created no later than a certain date by registering it with a notary service,
such as https://opentimestamps.org.

This is not a perfect way of assigning ownership to files. Anyone can take an
image in the public domain, embed their public key, and claim that they
created it. The only way to disprove such a claim is to find a copy of the
image that does not include the public key, or that predates it. Users should
be aware of what `slink` guarantees and what it does not.
