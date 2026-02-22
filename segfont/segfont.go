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
// Sourced from the Python led-kurokku TM1637 driver.
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
	'b': 0x7C, // c d e f g
	'c': 0x58, // d e g
	'C': 0x39, // a d e f
	'd': 0x5E, // b c d e g
	'E': 0x79, // a d e f g
	'F': 0x71, // a e f g
	'G': 0x3D, // a c d e f
	'H': 0x76, // b c e f g
	'h': 0x74, // c e f g
	'I': 0x30, // e f
	'J': 0x1E, // b c d e
	'k': 0x76, // b c e f g (same as H)
	'L': 0x38, // d e f
	'm': 0x55, // a c e g
	'n': 0x54, // c e g
	'o': 0x5C, // c d e g
	'O': 0x3F, // a b c d e f
	'P': 0x73, // a b e f g
	'q': 0x67, // a b c f g
	'r': 0x50, // e g
	'S': 0x6D, // a c d f g
	't': 0x78, // d e f g
	'U': 0x3E, // b c d e f
	'v': 0x1C, // c d e
	'w': 0x2A, // b d f (alternating segments)
	'x': 0x76, // b c e f g (same as H)
	'y': 0x6E, // b c d f g
	'z': 0x5B, // a b d e g (same as 2)
	'-': 0x40, // g
	'_': 0x08, // d
	'*': 0x63, // a b f g (degree symbol)
	'°': 0x63, // a b f g (degree symbol)
	' ': 0x00,
}

// 14-segment bit layout:
//
//        A
//    F  H I J  B
//      G1  G2
//    E  K L M  C
//        D
//
// bit  0 = A   (top horizontal)
// bit  1 = B   (top right vertical)
// bit  2 = C   (bottom right vertical)
// bit  3 = D   (bottom horizontal)
// bit  4 = E   (bottom left vertical)
// bit  5 = F   (top left vertical)
// bit  6 = G1  (middle left horizontal)
// bit  7 = G2  (middle right horizontal)
// bit  8 = H   (top left diagonal)
// bit  9 = I   (top center vertical)
// bit 10 = J   (top right diagonal)
// bit 11 = K   (bottom left diagonal)
// bit 12 = L   (bottom center vertical)
// bit 13 = M   (bottom right diagonal)

// Seg14 maps runes to 14-segment bitmasks.
// Sourced from the Python led-kurokku HT16K33 driver.
var Seg14 = map[rune]uint16{
	// Numbers 0-9
	'0': 0x003F, // A+B+C+D+E+F
	'1': 0x0006, // B+C
	'2': 0x00DB, // A+B+G1+G2+E+D
	'3': 0x00CF, // A+B+C+D+G1+G2
	'4': 0x00E6, // F+G1+G2+B+C
	'5': 0x00ED, // A+F+G1+G2+C+D
	'6': 0x00FD, // A+F+G1+G2+E+C+D
	'7': 0x0007, // A+B+C
	'8': 0x00FF, // All standard segments
	'9': 0x00EF, // A+F+G1+G2+B+C+D

	// Uppercase alphabet
	'A': 0x00F7, // A+F+G1+G2+B+C+E
	'B': 0x128F, // A+B+C+D+G2+I+L
	'C': 0x0039, // A+F+E+D
	'D': 0x120F, // A+B+C+D+I+L
	'E': 0x00F9, // A+F+G1+G2+E+D
	'F': 0x00F1, // A+F+G1+G2+E
	'G': 0x00BD, // A+F+E+D+C+G2
	'H': 0x00F6, // F+G1+G2+B+C+E
	'I': 0x1209, // A+D+I+L
	'J': 0x001E, // B+C+D+E
	'K': 0x2470, // F+G1+E+J+M
	'L': 0x0038, // F+E+D
	'M': 0x0536, // F+B+C+E+H+J
	'N': 0x2136, // F+B+C+E+H+M
	'O': 0x003F, // A+B+C+D+E+F
	'P': 0x00F3, // A+F+G1+G2+B+E
	'Q': 0x203F, // A+B+C+D+E+F+M
	'R': 0x20F3, // A+F+G1+G2+B+E+M
	'S': 0x00ED, // A+F+G1+G2+C+D
	'T': 0x1201, // A+I+L
	'U': 0x003E, // F+B+C+D+E
	'V': 0x0C30, // F+E+K+J
	'W': 0x2836, // F+B+C+E+K+M
	'X': 0x2D00, // H+J+K+M
	'Y': 0x1500, // H+J+L
	'Z': 0x0C09, // A+D+K+J

	// Lowercase alphabet
	'a': 0x00F7,
	'b': 0x00FC, // F+G1+G2+E+C+D
	'c': 0x00D8, // G1+G2+E+D
	'd': 0x00DE, // B+G1+G2+C+D+E
	'e': 0x00F9,
	'f': 0x00F1,
	'g': 0x00EF, // A+F+G1+G2+B+C+D
	'h': 0x00F4, // F+G1+G2+C+E
	'i': 0x1000, // L
	'j': 0x000E, // B+C+D
	'k': 0x2470,
	'l': 0x1200, // I+L
	'm': 0x10D4, // G1+G2+C+E+L
	'n': 0x1050, // G1+E+L
	'o': 0x00DC, // G1+G2+E+C+D
	'p': 0x00F3,
	'q': 0x00E7, // A+F+G1+G2+B+C
	'r': 0x00D0, // G1+G2+E
	's': 0x00ED,
	't': 0x00F8, // F+G1+G2+E+D
	'u': 0x001C, // C+D+E
	'v': 0x0810, // E+K
	'w': 0x2814, // C+E+K+M
	'x': 0x2D00,
	'y': 0x008E, // B+C+G1+G2+D
	'z': 0x0C09,

	// Special characters
	' ':  0x0000,
	'-':  0x00C0, // G1+G2
	'_':  0x0008, // D
	'°':  0x00E3, // A+B+F+G1+G2 (degree)
	'*':  0x2DC0, // H+J+K+M+G1+G2 (asterisk)
	'.':  0x4000, // DP (if supported)
	',':  0x4000,
	'!':  0x1206, // B+C+I+L
	'?':  0x1083, // A+B+G2+I
	'/':  0x0C00, // K+J
	'\\': 0x2100, // H+M
	'+':  0x12C0, // G1+G2+I+L
	'=':  0x00C8, // G1+G2+D
	'\'': 0x0400, // J
	'"':  0x0500, // H+J  (simplified)
	'(':  0x0C00, // K+J  (angled)
	')':  0x2100, // H+M  (angled)
	'[':  0x0039, // A+F+E+D
	']':  0x000F, // A+B+C+D
	'<':  0x2100, // H+M
	'>':  0x0C00, // K+J
	':':  0x1200, // I+L
	'^':  0x0023, // A+B+F (simplified)
	'&':  0x2359, // complex approximation
	'#':  0x12CE,
	'@':  0x12BB,
	'%':  0x6C24,
	'$':  0x12ED, // S with center vertical
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
