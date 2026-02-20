package animation

import "github.com/swilcox/led-kurokku-go/widget"

// Registry maps animation type names to constructors.
var Registry = map[string]func() widget.Widget{
	"rain":   func() widget.Widget { return &Rain{} },
	"random": func() widget.Widget { return &Random{} },
}
