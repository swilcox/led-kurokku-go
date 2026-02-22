package font

//go:generate go run ../cmd/genfont

// IsHangul reports whether r is a precomposed Hangul syllable (U+AC00–U+D7A3).
func IsHangul(r rune) bool {
	return r >= 0xAC00 && r <= 0xD7A3
}

// HangulGlyph returns the 8-column bitmap for a Hangul syllable.
// Returns false if r is not in the hangul font map.
func HangulGlyph(r rune) ([8]byte, bool) {
	g, ok := hangulFont[r]
	return g, ok
}
