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
	"rain":   func() widget.Widget { return &Rain{} },
	"random": func() widget.Widget { return &Random{} },
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
