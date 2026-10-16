package view

import (
	"fmt"

	"github.com/weqqr/rio/pkg/ui/react"
)

func Example() {
	count := react.NewRef(0)
	comp := New(&Config{TextMeasurer: Monospace{AdvanceRatio: 0.5, LineHeightRatio: 1}})
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Text(fmt.Sprintf("count: %d", count.Get()))
			c.Text(fmt.Sprintf("twice: %d", count.Get()*2), Padding(1), Background(Color{A: 1}))
		}, Gap(2), Padding(1))
	})

	frame := func() {
		canvas := &Recorder{}
		size := comp.Measure(Loose(Size{Width: 100, Height: 50}))
		comp.Draw(canvas)
		fmt.Printf("frame %v\n", size)
		for _, op := range canvas.Ops {
			switch op := op.(type) {
			case DrawTextOp:
				fmt.Printf("  text %q at %v\n", op.Text, op.At)
			case FillRectOp:
				fmt.Printf("  fill %v to %v\n", op.Rect.Min, op.Rect.Max)
			}
		}
	}

	frame()
	count.Set(7)
	frame()

	// Output:
	// frame {68 38}
	//   text "count: 0" at {1 1}
	//   fill {2 20} to {66 36}
	//   text "twice: 0" at {2 20}
	// frame {76 38}
	//   text "count: 7" at {1 1}
	//   fill {2 20} to {74 36}
	//   text "twice: 14" at {2 20}
}
