package animation

import (
	"context"
	"math/rand"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/framebuf"
)

// Life runs Conway's Game of Life on the 32x8 display with toroidal wrapping.
// When the grid stagnates (no change for 3 generations), it re-seeds randomly.
type Life struct{}

func (l *Life) Name() string { return "life" }

func (l *Life) Run(ctx context.Context, disp display.Display) error {
	pd := disp.(display.PixelDisplay)

	grid := lifeNewGrid()
	stagnant := 0

	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		next := lifeStep(grid)

		var f framebuf.Frame
		for x := 0; x < 32; x++ {
			for y := 0; y < 8; y++ {
				if next[x][y] {
					f.SetPixel(x, y, true)
				}
			}
		}
		pd.WriteFramebuffer(f.Bytes())

		if next == grid {
			stagnant++
		} else {
			stagnant = 0
		}
		if stagnant >= 3 {
			grid = lifeNewGrid()
			stagnant = 0
		} else {
			grid = next
		}
	}
}

func lifeNewGrid() [32][8]bool {
	var g [32][8]bool
	for x := 0; x < 32; x++ {
		for y := 0; y < 8; y++ {
			g[x][y] = rand.Intn(3) == 0 // ~33% alive
		}
	}
	return g
}

func lifeStep(grid [32][8]bool) [32][8]bool {
	var next [32][8]bool
	for x := 0; x < 32; x++ {
		for y := 0; y < 8; y++ {
			n := lifeNeighbors(grid, x, y)
			if grid[x][y] {
				next[x][y] = n == 2 || n == 3
			} else {
				next[x][y] = n == 3
			}
		}
	}
	return next
}

func lifeNeighbors(grid [32][8]bool, x, y int) int {
	count := 0
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := (x + dx + 32) % 32
			ny := (y + dy + 8) % 8
			if grid[nx][ny] {
				count++
			}
		}
	}
	return count
}
