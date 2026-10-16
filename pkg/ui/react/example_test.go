package react

import "fmt"

func Example() {
	scope := NewEffectScope()

	count := NewRef(0)
	label := NewComputed(func() string {
		if count.Get() == 0 {
			return "empty"
		}
		return fmt.Sprintf("%d items", count.Get())
	})

	var frames []string
	scope.Run(func() {
		WatchEffect(func() { frames = append(frames, "frame: "+label.Get()) }, FlushSync)
	})

	count.Set(1)
	Batch(func() {
		count.Set(2)
		count.Set(3)
	})
	count.Set(3)
	scope.Stop()
	count.Set(4)

	for _, frame := range frames {
		fmt.Println(frame)
	}
}

// Output:
// frame: empty
// frame: 1 items
// frame: 3 items
