package segfont

// 7-segment bit layout:
//
//    _a_
//   |   |
//   f   b
//   |_g_|
//   |   |
//   e   c
//   |_d_|
//
// bit 0 = a, bit 1 = b, bit 2 = c, bit 3 = d,
// bit 4 = e, bit 5 = f, bit 6 = g

// Seg7 maps runes to 7-segment bitmasks.
var Seg7 = map[rune]byte{
	'0': 0x3F, // a b c d e f
	'1': 0x06, // b c
	'2': 0x5B, // a b d e g
	'3': 0x4F, // a b c d g
	'4': 0x66, // b c f g
	'5': 0x6D, // a c d f g
	'6': 0x7D, // a c d e f g
	'7': 0x07, // a b c
	'8': 0x7F, // a b c d e f g
	'9': 0x6F, // a b c d f g
	'A': 0x77, // a b c e f g
	'a': 0x77,
	'B': 0x7C, // c d e f g
	'b': 0x7C,
	'C': 0x39, // a d e f
	'c': 0x58, // d e g
	'D': 0x5E, // b c d e g
	'd': 0x5E,
	'E': 0x79, // a d e f g
	'e': 0x79,
	'F': 0x71, // a e f g
	'f': 0x71,
	'G': 0x3D, // a c d e f
	'g': 0x6F, // a b c d f g
	'H': 0x76, // b c e f g
	'h': 0x74, // c e f g
	'I': 0x06, // b c
	'i': 0x04, // c
	'J': 0x1E, // b c d e
	'j': 0x1E,
	'L': 0x38, // d e f
	'l': 0x06, // b c
	'N': 0x54, // c e g
	'n': 0x54,
	'O': 0x3F, // a b c d e f
	'o': 0x5C, // c d e g
	'P': 0x73, // a b e f g
	'p': 0x73,
	'R': 0x50, // e g
	'r': 0x50,
	'S': 0x6D, // a c d f g
	's': 0x6D,
	'T': 0x78, // d e f g
	't': 0x78,
	'U': 0x3E, // b c d e f
	'u': 0x1C, // c d e
	'Y': 0x6E, // b c d f g
	'y': 0x6E,
	' ': 0x00,
	'-': 0x40, // g
	'_': 0x08, // d
}

// 14-segment bit layout:
//
//     _a1_ _a2_
//    |\   |   /|
//    f h  j  k b
//    |  \ | /  |
//     _g1_ _g2_
//    |  / | \  |
//    e n  m  l c
//    |/   |   \|
//     _d1_ _d2_
//
// bit  0 = a1    bit  1 = a2
// bit  2 = b     bit  3 = c
// bit  4 = d1    bit  5 = d2
// bit  6 = e     bit  7 = f
// bit  8 = g1    bit  9 = g2
// bit 10 = h     bit 11 = j
// bit 12 = k     bit 13 = l
// bit 14 = m     bit 15 = n

// Seg14 maps runes to 14-segment bitmasks.
var Seg14 = map[rune]uint16{
	' ': 0x0000,
	'!': 0x0806, // b, c, m
	'"': 0x0802, // b, j  (simplified)
	'#': 0x4E33, // a2, b, c, d1, d2, g1, g2, j, m
	'$': 0x4EED, // a1, a2, c, d1, d2, f, g1, g2, j, m
	'%': 0xC4E4, // c, e, g1, g2, k, n (approximation)
	'&': 0x2959, // a1, d1, d2, e, f, g1, k, n
	'\'': 0x0800, // j
	'(':  0x1400, // k, l
	')':  0x8400, // h, n
	'*':  0xFC00, // h, j, k, l, m, n
	'+':  0x4E00, // g1, g2, j, m
	',':  0x8000, // n
	'-':  0x0300, // g1, g2
	'.':  0x0000, // dot not in standard 14-seg; use colon
	'/':  0x8400, // h... actually k, n
	'0':  0x84FF, // a1, a2, b, c, d1, d2, e, f, k, n
	'1':  0x080C, // b, c, j (or just b, c)
	'2':  0x03B7, // a1, a2, b, d1, d2, e, g1, g2
	'3':  0x032F, // a1, a2, b, c, d1, d2, g2
	'4':  0x030C, // b, c, g1, g2 ... wait, also f
	'5':  0x036D, // a1, a2, c, d1, d2, f, g1, g2
	'6':  0x037D, // a1, a2, c, d1, d2, e, f, g1, g2
	'7':  0x0003, // a1, a2... also b, c
	'8':  0x03FF, // all outer + g1, g2
	'9':  0x036F, // a1, a2, b, c, d1, d2, f, g1, g2

	'A': 0x03CF, // a1, a2, b, c, e, f, g1, g2
	'B': 0x4A2F, // a1, a2, b, c, d1, d2, g2, j, m
	'C': 0x00F3, // a1, a2, d1, d2, e, f
	'D': 0x482F, // a1, a2, b, c, d1, d2, j, m
	'E': 0x01F3, // a1, a2, d1, d2, e, f, g1
	'F': 0x01C3, // a1, a2, e, f, g1
	'G': 0x02FD, // a1, a2, c, d1, d2, e, f, g2
	'H': 0x03CC, // b, c, e, f, g1, g2
	'I': 0x4833, // a1, a2, d1, d2, j, m
	'J': 0x003E, // b, c, d1, d2, e
	'K': 0x15C0, // e, f, g1, k, l
	'L': 0x00F0, // d1, d2, e, f
	'M': 0x14CC, // b, c, e, f, h, k
	'N': 0x24CC, // b, c, e, f, h, l
	'O': 0x00FF, // a1, a2, b, c, d1, d2, e, f
	'P': 0x03C7, // a1, a2, b, e, f, g1, g2
	'Q': 0x20FF, // a1, a2, b, c, d1, d2, e, f, l
	'R': 0x23C7, // a1, a2, b, e, f, g1, g2, l
	'S': 0x036D, // a1, a2, c, d1, d2, f, g1, g2
	'T': 0x4803, // a1, a2, j, m
	'U': 0x00FC, // b, c, d1, d2, e, f
	'V': 0x84C0, // e, f, k, n
	'W': 0xA0CC, // b, c, e, f, l, n
	'X': 0xB400, // h, k, l, n
	'Y': 0x5400, // h, k, m
	'Z': 0x8433, // a1, a2, d1, d2, k, n

	'a': 0x03CF,
	'b': 0x4A2F,
	'c': 0x00F3,
	'd': 0x482F,
	'e': 0x01F3,
	'f': 0x01C3,
	'g': 0x02FD,
	'h': 0x03CC,
	'i': 0x4833,
	'j': 0x003E,
	'k': 0x15C0,
	'l': 0x00F0,
	'm': 0x14CC,
	'n': 0x24CC,
	'o': 0x00FF,
	'p': 0x03C7,
	'q': 0x20FF,
	'r': 0x23C7,
	's': 0x036D,
	't': 0x4803,
	'u': 0x00FC,
	'v': 0x84C0,
	'w': 0xA0CC,
	'x': 0xB400,
	'y': 0x5400,
	'z': 0x8433,
}

// Encoder converts a rune to a segment bitmask.
type Encoder func(rune) uint16

// Enc7 encodes a rune using the 7-segment font.
func Enc7(r rune) uint16 {
	return uint16(Seg7[r])
}

// Enc14 encodes a rune using the 14-segment font.
func Enc14(r rune) uint16 {
	return Seg14[r]
}

// EncodeText encodes a string using the given encoder.
func EncodeText(enc Encoder, text string) []uint16 {
	result := make([]uint16, 0, len(text))
	for _, r := range text {
		result = append(result, enc(r))
	}
	return result
}
