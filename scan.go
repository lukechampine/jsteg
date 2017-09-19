// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsteg

import "image/jpeg"

const blockSize = 64 // A DCT block is 8x8.

type block [blockSize]int32

// Specified in section B.2.3.
func (d *decoder) processSOS(n int) error {
	if d.nComp == 0 {
		return jpeg.FormatError("missing SOF marker")
	}
	if n < 6 || 4+2*d.nComp < n || n%2 != 0 {
		return jpeg.FormatError("SOS has wrong length")
	}
	if err := d.readFull(d.tmp[:n]); err != nil {
		return err
	}
	nComp := int(d.tmp[0])
	if n != 4+2*nComp {
		return jpeg.FormatError("SOS length inconsistent with number of components")
	}
	var scan [maxComponents]struct {
		compIndex uint8
		td        uint8 // DC table selector.
		ta        uint8 // AC table selector.
	}
	totalHV := 0
	for i := 0; i < nComp; i++ {
		cs := d.tmp[1+2*i] // Component selector.
		compIndex := -1
		for j, comp := range d.comp[:d.nComp] {
			if cs == comp.c {
				compIndex = j
			}
		}
		if compIndex < 0 {
			return jpeg.FormatError("unknown component selector")
		}
		scan[i].compIndex = uint8(compIndex)
		// Section B.2.3 states that "the value of Cs_j shall be different from
		// the values of Cs_1 through Cs_(j-1)". Since we have previously
		// verified that a frame's component identifiers (C_i values in section
		// B.2.2) are unique, it suffices to check that the implicit indexes
		// into d.comp are unique.
		for j := 0; j < i; j++ {
			if scan[i].compIndex == scan[j].compIndex {
				return jpeg.FormatError("repeated component selector")
			}
		}
		totalHV += d.comp[compIndex].h * d.comp[compIndex].v

		// The baseline t <= 1 restriction is specified in table B.3.
		scan[i].td = d.tmp[2+2*i] >> 4
		if t := scan[i].td; t > maxTh || (d.baseline && t > 1) {
			return jpeg.FormatError("bad Td value")
		}
		scan[i].ta = d.tmp[2+2*i] & 0x0f
		if t := scan[i].ta; t > maxTh || (d.baseline && t > 1) {
			return jpeg.FormatError("bad Ta value")
		}
	}
	// Section B.2.3 states that if there is more than one component then the
	// total H*V values in a scan must be <= 10.
	if d.nComp > 1 && totalHV > 10 {
		return jpeg.FormatError("total sampling factors too large")
	}

	// mxx and myy are the number of MCUs (Minimum Coded Units) in the image.
	h0, v0 := d.comp[0].h, d.comp[0].v // The h and v values from the Y components.
	mxx := (d.width + 8*h0 - 1) / (8 * h0)
	myy := (d.height + 8*v0 - 1) / (8 * v0)

	d.bits = bits{}
	mcu, expectedRST := 0, uint8(rst0Marker)
	for my := 0; my < myy; my++ {
		for mx := 0; mx < mxx; mx++ {
			for i := 0; i < nComp; i++ {
				compIndex := scan[i].compIndex
				hi := d.comp[compIndex].h
				vi := d.comp[compIndex].v
				for j := 0; j < hi*vi; j++ {
					// Decode the DC coefficient, as specified in section F.2.2.1.
					value, err := d.decodeHuffman(&d.huff[dcTable][scan[i].td])
					if err != nil {
						return err
					}
					if value > 16 {
						return jpeg.UnsupportedError("excessive DC component")
					}
					if _, err = d.receiveExtend(value); err != nil {
						return err
					}

					// Decode the AC coefficients, as specified in section F.2.2.2.
					huff := &d.huff[acTable][scan[i].ta]
					for zig := 1; zig < blockSize; zig++ {
						value, err := d.decodeHuffman(huff)
						if err != nil {
							return err
						}
						val0 := value >> 4
						val1 := value & 0x0f
						if val1 != 0 {
							zig += int(val0)
							if zig > blockSize {
								break
							}
							ac, err := d.receiveExtend(val1)
							if err != nil {
								return err
							}

							// steganography
							if i == 0 && (ac < -1 || ac > 1) {
								if d.databit == 0 {
									d.data = append(d.data, 0)
								}
								d.data[len(d.data)-1] |= byte((ac & 1) << d.databit)
								d.databit = (d.databit + 1) % 8
							}

						} else {
							if val0 != 0x0f {
								break
							}
							zig += 0x0f
						}
					}
				} // for j
			} // for i
			mcu++
			if d.ri > 0 && mcu%d.ri == 0 && mcu < mxx*myy {
				// A more sophisticated decoder could use RST[0-7] markers to resynchronize from corrupt input,
				// but this one assumes well-formed input, and hence the restart marker follows immediately.
				if err := d.readFull(d.tmp[:2]); err != nil {
					return err
				}
				if d.tmp[0] != 0xff || d.tmp[1] != expectedRST {
					return jpeg.FormatError("bad RST marker")
				}
				expectedRST++
				if expectedRST == rst7Marker+1 {
					expectedRST = rst0Marker
				}
				// Reset the Huffman decoder.
				d.bits = bits{}
			}
		} // for mx
	} // for my

	return nil
}
