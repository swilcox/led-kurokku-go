package animation

import "github.com/swilcox/led-kurokku-go/widget"

// Registry maps animation type names to constructors.
var Registry = map[string]func() widget.Widget{
	"rain":    func() widget.Widget { return &Rain{} },
	"random":  func() widget.Widget { return &Random{} },
	"bounce":  func() widget.Widget { return &Bounce{} },
	"sine":    func() widget.Widget { return &Sine{} },
	"scanner": func() widget.Widget { return &Scanner{} },
	"life":    func() widget.Widget { return &Life{} },
}
