package segment

import (
	"context"
	"math/rand"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/widget"
)

// Registry maps segment animation type names to constructors.
var Registry = map[string]func() widget.Widget{
	"rain":    func() widget.Widget { return &Rain{} },
	"random":  func() widget.Widget { return &Random{} },
	"scanner": func() widget.Widget { return &Scanner{} },
	"race":    func() widget.Widget { return &Race{} },
}

// Rain simulates segments lighting up and "falling" through each digit.
// For 7-segment: top → middle → bottom segments cycle.
// For 14-segment: top → middle → bottom segments cycle with diagonals.
type Rain struct{}

func (r *Rain) Name() string { return "segment-rain" }

func (r *Rain) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	n := sd.DisplayLength()

	// Each digit tracks a "drop" position (0-5 = falling stages, -1 = inactive)
	drops := make([]int, n)
	for i := range drops {
		drops[i] = -1
	}

	// 7-segment rain stages: top(a) → upper-sides(f,b) → middle(g) → lower-sides(e,c) → bottom(d) → off
	seg7Stages := []uint16{
		0x01, // a
		0x22, // f, b
		0x40, // g
		0x14, // e, c
		0x08, // d
		0x00, // off
	}

	// 14-segment rain stages: top(A) → diagonals(H,J)+verticals(F,B) → middle(G1,G2) →
	//   diagonals(K,M)+verticals(E,C) → bottom(D) → off
	seg14Stages := []uint16{
		0x0001, // A
		0x0522, // F, B, H, J
		0x00C0, // G1, G2
		0x2814, // E, C, K, M
		0x0008, // D
		0x0000, // off
	}

	stages := seg7Stages
	if n > 0 {
		// Detect 14-seg by trying a write and checking the display type
		// Since we can't query, use a simple heuristic: TerminalSegment14 and HT16K33
		// always get Enc14 from the engine, but here we just check segment count.
		// Use the 14-seg stages if we have them available.
		// The engine sets the correct animation, so we use 7-seg by default
		// and the registry can be extended. For now, use display length as hint is insufficient.
		// Keep it simple: 7-seg stages work on both (low byte only on 14-seg still lights segments).
	}
	_ = seg14Stages // available for future use or explicit 14-seg rain variant

	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		// Randomly spawn new drops
		for i := range drops {
			if drops[i] < 0 && rand.Intn(6) == 0 {
				drops[i] = 0
			}
		}

		segments := make([]uint16, n)
		for i := range drops {
			if drops[i] < 0 {
				continue
			}
			if drops[i] < len(stages) {
				segments[i] = stages[drops[i]]
			}
			drops[i]++
			if drops[i] >= len(stages)+2 {
				drops[i] = -1
			}
		}

		sd.WriteSegments(segments, false)
	}
}

// Rain14 is a 14-segment-specific rain animation that uses diagonal segments.
type Rain14 struct{}

func (r *Rain14) Name() string { return "segment-rain14" }

func (r *Rain14) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	n := sd.DisplayLength()

	drops := make([]int, n)
	for i := range drops {
		drops[i] = -1
	}

	stages := []uint16{
		0x0001, // A
		0x0522, // F, B, H, J
		0x00C0, // G1, G2
		0x2814, // E, C, K, M
		0x0008, // D
		0x0000, // off
	}

	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		for i := range drops {
			if drops[i] < 0 && rand.Intn(6) == 0 {
				drops[i] = 0
			}
		}

		segments := make([]uint16, n)
		for i := range drops {
			if drops[i] < 0 {
				continue
			}
			if drops[i] < len(stages) {
				segments[i] = stages[drops[i]]
			}
			drops[i]++
			if drops[i] >= len(stages)+2 {
				drops[i] = -1
			}
		}

		sd.WriteSegments(segments, false)
	}
}

// Random displays random segments lighting up across all digits.
type Random struct{}

func (r *Random) Name() string { return "segment-random" }

func (r *Random) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	n := sd.DisplayLength()

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		segments := make([]uint16, n)
		for i := range segments {
			segments[i] = uint16(rand.Intn(0x80)) // 7-seg range (7 bits)
		}

		sd.WriteSegments(segments, rand.Intn(2) == 0)
	}
}

// Scanner sweeps a vertical bar back and forth across the digits.
type Scanner struct{}

func (s *Scanner) Name() string { return "segment-scanner" }

func (s *Scanner) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	n := sd.DisplayLength()
	if n == 0 {
		return nil
	}

	// 7-segment vertical bar: b + c + e + f = all four vertical segments
	bar := uint16(0x36)

	// Build bounce sequence: 0,1,2,3,2,1 for n=4
	seq := make([]int, 0, 2*n-2)
	for i := 0; i < n; i++ {
		seq = append(seq, i)
	}
	for i := n - 2; i > 0; i-- {
		seq = append(seq, i)
	}

	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	pos := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		segments := make([]uint16, n)
		segments[seq[pos]] = bar
		sd.WriteSegments(segments, false)

		pos = (pos + 1) % len(seq)
	}
}

// Scanner14 sweeps a vertical bar using 14-segment patterns (includes center verticals).
type Scanner14 struct{}

func (s *Scanner14) Name() string { return "segment-scanner14" }

func (s *Scanner14) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	n := sd.DisplayLength()
	if n == 0 {
		return nil
	}

	// 14-segment vertical bar: B + C + E + F + J + M (outer + center verticals)
	// B=bit1, C=bit2, E=bit4, F=bit5, J=bit10, M=bit13
	bar := uint16(0x2436)

	seq := make([]int, 0, 2*n-2)
	for i := 0; i < n; i++ {
		seq = append(seq, i)
	}
	for i := n - 2; i > 0; i-- {
		seq = append(seq, i)
	}

	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	pos := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		segments := make([]uint16, n)
		segments[seq[pos]] = bar
		sd.WriteSegments(segments, false)

		pos = (pos + 1) % len(seq)
	}
}

// trackStep represents one position on the racetrack: which digit and which segment bit.
type trackStep struct {
	digit int
	bit   uint16
}

// buildTrack7 builds the perimeter track for a 7-segment display with n digits.
// Path: top L→R (a), down right side (b,c), bottom R→L (d), up left side (e,f).
func buildTrack7(n int) []trackStep {
	track := make([]trackStep, 0, 2*n+4)
	// Top: a (bit 0) across all digits left to right
	for i := 0; i < n; i++ {
		track = append(track, trackStep{i, 0x01})
	}
	// Right side down: b (bit 1), c (bit 2) on last digit
	track = append(track, trackStep{n - 1, 0x02})
	track = append(track, trackStep{n - 1, 0x04})
	// Bottom: d (bit 3) across all digits right to left
	for i := n - 1; i >= 0; i-- {
		track = append(track, trackStep{i, 0x08})
	}
	// Left side up: e (bit 4), f (bit 5) on first digit
	track = append(track, trackStep{0, 0x10})
	track = append(track, trackStep{0, 0x20})
	return track
}

// buildTrack14 builds the perimeter track for a 14-segment display with n digits.
// Same shape as 7-seg but uses 14-seg bit positions.
func buildTrack14(n int) []trackStep {
	track := make([]trackStep, 0, 2*n+4)
	// Top: A (bit 0)
	for i := 0; i < n; i++ {
		track = append(track, trackStep{i, 0x01})
	}
	// Right side down: B (bit 1), C (bit 2)
	track = append(track, trackStep{n - 1, 0x02})
	track = append(track, trackStep{n - 1, 0x04})
	// Bottom: D (bit 3)
	for i := n - 1; i >= 0; i-- {
		track = append(track, trackStep{i, 0x08})
	}
	// Left side up: E (bit 4), F (bit 5)
	track = append(track, trackStep{0, 0x10})
	track = append(track, trackStep{0, 0x20})
	return track
}

// Race animates two segments chasing each other around the perimeter of the display.
type Race struct{}

func (r *Race) Name() string { return "segment-race" }

func (r *Race) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	n := sd.DisplayLength()
	if n == 0 {
		return nil
	}

	track := buildTrack7(n)
	return runRace(ctx, sd, n, track)
}

// Race14 animates two segments chasing each other using 14-segment bit positions.
type Race14 struct{}

func (r *Race14) Name() string { return "segment-race14" }

func (r *Race14) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	n := sd.DisplayLength()
	if n == 0 {
		return nil
	}

	track := buildTrack14(n)
	return runRace(ctx, sd, n, track)
}

func runRace(ctx context.Context, sd display.SegmentDisplay, n int, track []trackStep) error {
	trackLen := len(track)
	// Two chasers half the track apart
	pos1 := 0
	pos2 := trackLen / 2

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		segments := make([]uint16, n)
		s1 := track[pos1]
		s2 := track[pos2]
		segments[s1.digit] |= s1.bit
		segments[s2.digit] |= s2.bit

		sd.WriteSegments(segments, false)

		pos1 = (pos1 + 1) % trackLen
		pos2 = (pos2 + 1) % trackLen
	}
}

// Random14 displays random 14-segment patterns across all digits.
type Random14 struct{}

func (r *Random14) Name() string { return "segment-random14" }

func (r *Random14) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	n := sd.DisplayLength()

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		segments := make([]uint16, n)
		for i := range segments {
			segments[i] = uint16(rand.Intn(0x4000)) // 14-seg range (14 bits)
		}

		sd.WriteSegments(segments, rand.Intn(2) == 0)
	}
}
