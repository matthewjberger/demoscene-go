package main

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// decodeHDR decodes a Radiance RGBE (.hdr) image into tightly packed float16
// RGBA pixels (8 bytes per pixel, little-endian), ready to upload to an
// RGBA16Float texture. Only the standard "-Y height +X width" orientation is
// supported, which is what Poly Haven serves.
func decodeHDR(data []byte) (int, int, []byte, error) {
	pos := 0
	readLine := func() (string, bool) {
		start := pos
		for pos < len(data) {
			if data[pos] == '\n' {
				line := string(data[start:pos])
				pos++
				return strings.TrimRight(line, "\r"), true
			}
			pos++
		}
		return "", false
	}

	first, ok := readLine()
	if !ok || (!strings.HasPrefix(first, "#?RADIANCE") && !strings.HasPrefix(first, "#?RGBE")) {
		return 0, 0, nil, fmt.Errorf("hdr: missing radiance signature")
	}
	for {
		line, ok := readLine()
		if !ok {
			return 0, 0, nil, fmt.Errorf("hdr: unexpected eof in header")
		}
		if line == "" {
			break
		}
	}

	resolution, ok := readLine()
	if !ok {
		return 0, 0, nil, fmt.Errorf("hdr: missing resolution line")
	}
	fields := strings.Fields(resolution)
	if len(fields) != 4 || fields[0] != "-Y" || fields[2] != "+X" {
		return 0, 0, nil, fmt.Errorf("hdr: unsupported resolution line %q", resolution)
	}
	height, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, nil, fmt.Errorf("hdr: parse height: %w", err)
	}
	width, err := strconv.Atoi(fields[3])
	if err != nil {
		return 0, 0, nil, fmt.Errorf("hdr: parse width: %w", err)
	}
	if width <= 0 || height <= 0 {
		return 0, 0, nil, fmt.Errorf("hdr: invalid dimensions %dx%d", width, height)
	}

	out := make([]byte, width*height*8)
	scan := make([]byte, width*4)
	for y := 0; y < height; y++ {
		if err := readScanline(data, &pos, width, scan); err != nil {
			return 0, 0, nil, fmt.Errorf("hdr: scanline %d: %w", y, err)
		}
		for x := 0; x < width; x++ {
			red, green, blue := rgbeToFloat(scan[x*4+0], scan[x*4+1], scan[x*4+2], scan[x*4+3])
			index := (y*width + x) * 8
			putFloat16(out[index:], red)
			putFloat16(out[index+2:], green)
			putFloat16(out[index+4:], blue)
			putFloat16(out[index+6:], 1.0)
		}
	}
	return width, height, out, nil
}

func readScanline(data []byte, pos *int, width int, scan []byte) error {
	cursor := *pos
	if cursor+4 > len(data) {
		return io.ErrUnexpectedEOF
	}
	header0, header1 := data[cursor], data[cursor+1]
	encodedWidth := int(data[cursor+2])<<8 | int(data[cursor+3])
	if width < 8 || width >= 0x8000 || header0 != 2 || header1 != 2 || encodedWidth != width {
		return readFlatScanline(data, pos, width, scan)
	}
	cursor += 4

	for channel := 0; channel < 4; channel++ {
		x := 0
		for x < width {
			if cursor >= len(data) {
				return io.ErrUnexpectedEOF
			}
			count := int(data[cursor])
			cursor++
			if count > 128 {
				runLength := count - 128
				if cursor >= len(data) {
					return io.ErrUnexpectedEOF
				}
				value := data[cursor]
				cursor++
				if x+runLength > width {
					return fmt.Errorf("hdr: rle run overflow")
				}
				for i := 0; i < runLength; i++ {
					scan[(x+i)*4+channel] = value
				}
				x += runLength
			} else {
				if count == 0 || x+count > width {
					return fmt.Errorf("hdr: rle literal overflow")
				}
				if cursor+count > len(data) {
					return io.ErrUnexpectedEOF
				}
				for i := 0; i < count; i++ {
					scan[(x+i)*4+channel] = data[cursor+i]
				}
				cursor += count
				x += count
			}
		}
	}
	*pos = cursor
	return nil
}

func readFlatScanline(data []byte, pos *int, width int, scan []byte) error {
	cursor := *pos
	shift := 0
	var previous [4]byte
	for x := 0; x < width; {
		if cursor+4 > len(data) {
			return io.ErrUnexpectedEOF
		}
		red, green, blue, exponent := data[cursor], data[cursor+1], data[cursor+2], data[cursor+3]
		cursor += 4
		if red == 1 && green == 1 && blue == 1 {
			count := int(exponent) << shift
			for i := 0; i < count && x < width; i++ {
				copy(scan[x*4:x*4+4], previous[:])
				x++
			}
			shift += 8
		} else {
			scan[x*4+0], scan[x*4+1], scan[x*4+2], scan[x*4+3] = red, green, blue, exponent
			previous = [4]byte{red, green, blue, exponent}
			x++
			shift = 0
		}
	}
	*pos = cursor
	return nil
}

func rgbeToFloat(red, green, blue, exponent byte) (float32, float32, float32) {
	if exponent == 0 {
		return 0, 0, 0
	}
	factor := float32(math.Ldexp(1.0, int(exponent)-(128+8)))
	return float32(red) * factor, float32(green) * factor, float32(blue) * factor
}

func putFloat16(dst []byte, value float32) {
	bits := float32ToFloat16Bits(value)
	dst[0] = byte(bits)
	dst[1] = byte(bits >> 8)
}

func float32ToFloat16Bits(value float32) uint16 {
	if value < 0 {
		value = 0
	}
	if value > 65504 {
		value = 65504
	}
	bits := math.Float32bits(value)
	sign := uint16((bits >> 16) & 0x8000)
	exponent := int32((bits>>23)&0xff) - 127 + 15
	mantissa := bits & 0x7fffff

	if exponent <= 0 {
		if exponent < -10 {
			return sign
		}
		mantissa |= 0x800000
		halfMantissa := mantissa >> uint32(14-exponent)
		if (mantissa>>uint32(13-exponent))&1 != 0 {
			halfMantissa++
		}
		return sign | uint16(halfMantissa)
	}
	if exponent >= 0x1f {
		return sign | 0x7c00
	}
	half := sign | uint16(exponent<<10) | uint16(mantissa>>13)
	if mantissa&0x1000 != 0 {
		half++
	}
	return half
}
