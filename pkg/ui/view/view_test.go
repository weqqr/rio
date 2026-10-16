package view

import (
	"fmt"
	"strings"
	"testing"

	"github.com/weqqr/rio/pkg/ui/react"
)

var (
	red   = Color{R: 1, A: 1}
	blue  = Color{B: 1, A: 1}
	green = Color{G: 1, A: 1}
)

func monospace() *Config {
	return &Config{TextMeasurer: Monospace{AdvanceRatio: 0.5, LineHeightRatio: 1}}
}

func drawOps(comp *Composition) []any {
	recorder := &Recorder{}
	comp.Draw(recorder)
	return recorder.Ops
}

func texts(ops []any) []string {
	var out []string
	for _, op := range ops {
		if text, ok := op.(DrawTextOp); ok {
			out = append(out, fmt.Sprintf("%s@%v", text.Text, text.At))
		}
	}
	return out
}

func rects(ops []any) []Rect {
	var out []Rect
	for _, op := range ops {
		if fill, ok := op.(FillRectOp); ok {
			out = append(out, fill.Rect)
		}
	}
	return out
}

func TestColumnStacksChildren(t *testing.T) {
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Text("aa")
			c.Text("bbbb", Padding(1))
		}, Gap(2))
	})

	if got, want := comp.Measure(Unbounded), (Size{Width: 34, Height: 36}); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "aa@{0 0}|bbbb@{1 19}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
}

func TestRowWeightsSplitTheRemainingSpace(t *testing.T) {
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Row(func(c *Scope) {
			c.Spacer(Sized(10, 10), Background(red))
			c.Spacer(Weight(1), Sized(0, 10), Background(blue))
			c.Spacer(Weight(2), Sized(0, 10), Background(green))
		})
	})

	if got, want := comp.Measure(Tight(Size{Width: 130, Height: 10})), (Size{Width: 130, Height: 10}); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	got := fmt.Sprint(rects(drawOps(comp)))
	want := "[{{0 0} {10 10}} {{10 0} {50 10}} {{50 0} {130 10}}]"
	if got != want {
		t.Fatalf("rects = %s, want %s", got, want)
	}
}

func TestColumnAlignsChildren(t *testing.T) {
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Spacer(Sized(60, 10))
			c.Spacer(Sized(20, 10), Align(AlignCenter), Background(red))
			c.Spacer(Sized(20, 10), Align(AlignEnd), Background(blue))
		}, Sized(60, 30))
	})

	comp.Measure(Unbounded)
	got := fmt.Sprint(rects(drawOps(comp)))
	want := "[{{20 10} {40 20}} {{40 20} {60 30}}]"
	if got != want {
		t.Fatalf("rects = %s, want %s", got, want)
	}
}

func TestBoxOverlapsAndAligns(t *testing.T) {
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Box(func(c *Scope) {
			c.Spacer(Sized(60, 20), Background(red))
			c.Spacer(Sized(20, 10), Align(AlignCenter), Background(blue))
		}, Sized(60, 20))
	})

	if got, want := comp.Measure(Unbounded), (Size{Width: 60, Height: 20}); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	got := fmt.Sprint(rects(drawOps(comp)))
	want := "[{{0 0} {60 20}} {{20 5} {40 15}}]"
	if got != want {
		t.Fatalf("rects = %s, want %s", got, want)
	}
}

func TestFillMaxSizeExpandsToTheConstraints(t *testing.T) {
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Box(func(c *Scope) {
			c.Spacer(FillMaxSize(), Background(red))
		}, FillMaxSize())
	})

	if got, want := comp.Measure(Loose(Size{Width: 100, Height: 50})), (Size{Width: 100, Height: 50}); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	if got, want := fmt.Sprint(rects(drawOps(comp))), "[{{0 0} {100 50}}]"; got != want {
		t.Fatalf("rects = %s, want %s", got, want)
	}

	if got, want := comp.Measure(Loose(Size{Width: 40, Height: 20})), (Size{Width: 40, Height: 20}); got != want {
		t.Fatalf("size after constraining = %v, want %v", got, want)
	}
}

func TestModifierOrderDecidesNesting(t *testing.T) {
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Text("x", Padding(2), Background(red))
			c.Text("y", Background(blue), Padding(2))
		}, Gap(2))
	})

	comp.Measure(Unbounded)
	got := fmt.Sprint(rects(drawOps(comp)))
	want := "[{{2 2} {10 18}} {{0 22} {12 42}}]"
	if got != want {
		t.Fatalf("rects = %s, want %s", got, want)
	}
}

func TestTextSizeChangesMeasurement(t *testing.T) {
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Text("ab")
			c.Text("ab", TextSize(32))
		})
	})

	if got, want := comp.Measure(Unbounded), (Size{Width: 32, Height: 48}); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "ab@{0 0}|ab@{0 16}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
}

type countingMeasurer struct {
	size  Size
	calls int
}

func (m *countingMeasurer) Measure(text string, style TextStyle, constraints Constraints) Size {
	m.calls++
	return m.size
}

func TestMeasureIsCachedUntilSomethingChanges(t *testing.T) {
	measurer := &countingMeasurer{size: Size{Width: 30, Height: 10}}
	comp := New(&Config{TextMeasurer: measurer})
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) { c.Text("whatever") })
	})

	if got, want := comp.Measure(Unbounded), (Size{Width: 30, Height: 10}); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	comp.Measure(Unbounded)
	if measurer.calls != 1 {
		t.Fatalf("measurer called %d times for an unchanged tree, want 1", measurer.calls)
	}
	if got, want := comp.Measure(Tight(Size{Width: 50, Height: 40})), (Size{Width: 50, Height: 40}); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	if measurer.calls != 2 {
		t.Fatalf("measurer called %d times for new constraints, want 2", measurer.calls)
	}
}

func TestDrawWithoutMeasureLaysOutUnbounded(t *testing.T) {
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Text("aaaa")
			c.Text("bb")
		})
	})

	ops := drawOps(comp)
	if got, want := strings.Join(texts(ops), "|"), "aaaa@{0 0}|bb@{0 16}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
}

type probeView struct {
	constraints Constraints
	offset      Point
	measures    int
	draws       int
	size        Size
}

func (v *probeView) Measure(constraints Constraints) Size {
	v.measures++
	v.constraints = constraints
	if v.size == (Size{}) {
		return Size{Width: 20, Height: 5}
	}
	return v.size
}

func (v *probeView) Place(offset Point) { v.offset = offset }

func (v *probeView) Draw(canvas Canvas) { v.draws++ }

func TestCustomViewTakesPartInLayout(t *testing.T) {
	probe := &probeView{}
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Row(func(c *Scope) {
			c.Text("ab")
			c.Custom(probe)
		}, Gap(4))
	})

	if got, want := comp.Measure(Loose(Size{Width: 100, Height: 40})), (Size{Width: 40, Height: 16}); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	comp.Draw(&Recorder{})
	if probe.offset != (Point{X: 20}) {
		t.Fatalf("probe placed at %v, want {20 0}", probe.offset)
	}
	if probe.constraints != (Constraints{MaxWidth: 100, MaxHeight: 40}) {
		t.Fatalf("probe measured with %v", probe.constraints)
	}
	if probe.draws != 1 {
		t.Fatalf("probe drawn %d times, want 1", probe.draws)
	}
}

func TestCustomViewTakesThePlaceOfThePreviousOne(t *testing.T) {
	first, second := &probeView{}, &probeView{}
	useFirst := react.NewRef(true)
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		if useFirst.Get() {
			c.Custom(first)
			return
		}
		c.Custom(second)
	})

	comp.Measure(Unbounded)
	useFirst.Set(false)
	comp.Draw(&Recorder{})
	if second.measures == 0 || second.draws != 1 {
		t.Fatalf("the second view measured %d times and drew %d times, want the swap to take effect", second.measures, second.draws)
	}
}

func TestRecompositionIsScopedToTheStateReader(t *testing.T) {
	count := react.NewRef(0)
	outer, inner := 0, 0
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Component(func(c *Scope) {
			outer++
			c.Text(fmt.Sprintf("outer %d", outer))
			c.Component(func(c *Scope) {
				inner++
				c.Text(fmt.Sprintf("count %d", count.Get()))
			})
		})
	})

	comp.Measure(Unbounded)
	count.Set(1)
	comp.Measure(Unbounded)

	if outer != 1 || inner != 2 {
		t.Fatalf("outer = %d, inner = %d, want 1 and 2", outer, inner)
	}
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "outer 1@{0 0}|count 1@{0 16}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
}

func TestRememberKeepsValuesAcrossRecompositions(t *testing.T) {
	count := react.NewRef(0)
	inits := 0
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Component(func(c *Scope) {
			kept := c.Remember(func() string {
				inits++
				return "kept"
			})
			c.Text(fmt.Sprintf("%s %d", kept, count.Get()))
		})
	})

	comp.Measure(Unbounded)
	count.Set(1)
	comp.Measure(Unbounded)

	if inits != 1 {
		t.Fatalf("remember initialised %d times, want 1", inits)
	}
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "kept 1@{0 0}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
}

func TestRememberKeyReinitialisesOnKeyChange(t *testing.T) {
	key := react.NewRef("a")
	inits := 0
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		value := c.RememberKey(key.Get(), func() string {
			inits++
			return fmt.Sprintf("made for %s", key.Peek())
		})
		c.Text(value)
	})

	comp.Measure(Unbounded)
	key.Set("b")
	comp.Measure(Unbounded)
	key.Set("b")
	comp.Measure(Unbounded)

	if inits != 2 {
		t.Fatalf("remember initialised %d times, want 2", inits)
	}
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "made for b@{0 0}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
}

func TestStateIsRememberedAndDrivesRendering(t *testing.T) {
	var state *react.Ref[int]
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		state = c.State(3)
		c.Text(fmt.Sprintf("value %d", state.Get()))
	})

	if got := strings.Join(texts(drawOps(comp)), "|"); got != "value 3@{0 0}" {
		t.Fatalf("texts = %q", got)
	}
	state.Set(4)
	if got := strings.Join(texts(drawOps(comp)), "|"); got != "value 4@{0 0}" {
		t.Fatalf("texts after write = %q", got)
	}
}

func TestDerivedRecomputesOnlyWhenItsSourcesChange(t *testing.T) {
	count := react.NewRef(1)
	renders := 0
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		label := c.Derived(func() string { return fmt.Sprintf("parity %d", count.Get()%2) })
		c.Text(label.Get())
		renders++
	})

	comp.Measure(Unbounded)
	count.Set(3)
	comp.Measure(Unbounded)
	if renders != 1 {
		t.Fatalf("rendered %d times for a derived value that did not change, want 1", renders)
	}
	count.Set(4)
	if got := strings.Join(texts(drawOps(comp)), "|"); got != "parity 0@{0 0}" {
		t.Fatalf("texts = %q", got)
	}
	if renders != 2 {
		t.Fatalf("rendered %d times, want 2", renders)
	}
}

func TestDisposableEffectLifecycle(t *testing.T) {
	key := react.NewRef("a")
	var events []string
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Component(func(c *Scope) {
			current := key.Get()
			c.DisposableEffect(current, func() func() {
				events = append(events, "start "+current)
				return func() { events = append(events, "stop "+current) }
			})
			c.Text(current)
		})
	})

	comp.Measure(Unbounded)
	key.Set("b")
	comp.Measure(Unbounded)
	comp.Dispose()

	want := "start a|stop a|start b|stop b"
	if got := strings.Join(events, "|"); got != want {
		t.Fatalf("events = %q, want %q", got, want)
	}
}

func TestUnmountedComponentStopsReacting(t *testing.T) {
	show := react.NewRef(true)
	count := react.NewRef(0)
	invalidations := 0
	comp := New(monospace())
	comp.config.OnInvalidate = func() { invalidations++ }
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Text("header")
			if show.Get() {
				c.Component(func(c *Scope) {
					c.Text(fmt.Sprintf("count %d", count.Get()))
				})
			}
		})
	})

	comp.Measure(Unbounded)
	if invalidations != 0 {
		t.Fatalf("invalidations = %d before any write", invalidations)
	}
	show.Set(false)
	if invalidations != 1 {
		t.Fatalf("invalidations = %d after one write, want 1", invalidations)
	}
	count.Set(1)
	count.Set(2)
	if invalidations != 1 {
		t.Fatalf("invalidations = %d, want the write on an unmounted node to be silent", invalidations)
	}
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "header@{0 0}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
}

func TestInvalidateFiresOncePerBurst(t *testing.T) {
	count := react.NewRef(0)
	invalidations := 0
	comp := New(monospace())
	comp.config.OnInvalidate = func() { invalidations++ }
	comp.SetContent(func(c *Scope) { c.Text(fmt.Sprintf("%d", count.Get())) })

	comp.Measure(Unbounded)
	count.Set(1)
	count.Set(2)
	if invalidations != 1 {
		t.Fatalf("invalidations = %d, want 1 for a burst of writes", invalidations)
	}
	comp.Measure(Unbounded)
	count.Set(2)
	if invalidations != 1 {
		t.Fatalf("invalidations = %d, want an equal write to be silent", invalidations)
	}
	count.Set(3)
	if invalidations != 2 {
		t.Fatalf("invalidations = %d, want 2", invalidations)
	}
}

func TestSideEffectRunsAfterTheBodyEmits(t *testing.T) {
	count := react.NewRef(0)
	var order []string
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Component(func(c *Scope) {
			order = append(order, "body")
			c.SideEffect(func() { order = append(order, "side") })
			c.Text(fmt.Sprintf("%d", count.Get()))
		})
	})

	comp.Measure(Unbounded)
	count.Set(1)
	comp.Measure(Unbounded)

	want := "body|side|body|side"
	if got := strings.Join(order, "|"); got != want {
		t.Fatalf("order = %q, want %q", got, want)
	}
}

func TestKeyedItemsKeepTheirStateWhenReordered(t *testing.T) {
	items := react.NewRef([]string{"a", "b", "c"})
	creations := 0
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Items(items.Get(), func(value string) any { return value }, func(c *Scope, value string) {
				label := c.Remember(func() string {
					creations++
					return "state of " + value
				})
				c.Text(label)
			})
		})
	})

	comp.Measure(Unbounded)
	items.Set([]string{"c", "a", "b"})
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "state of c@{0 0}|state of a@{0 16}|state of b@{0 32}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
	if creations != 3 {
		t.Fatalf("items created %d times, want 3 (their state must follow the key)", creations)
	}
}

func TestAppendingAnItemKeepsTheOtherItemsState(t *testing.T) {
	items := react.NewRef([]string{"a"})
	creations := 0
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Items(items.Get(), func(value string) any { return value }, func(c *Scope, value string) {
				label := c.Remember(func() string {
					creations++
					return "kept " + value
				})
				c.Text(label)
			})
		})
	})

	comp.Measure(Unbounded)
	items.Set([]string{"a", "b"})
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "kept a@{0 0}|kept b@{0 16}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
	if creations != 2 {
		t.Fatalf("remembered values created %d times, want 2: the existing item keeps its state", creations)
	}
}

func TestSiblingComponentsKeepSeparateState(t *testing.T) {
	first := func(c *Scope) {
		label := c.Remember(func() string { return "first" })
		c.Text(label)
	}
	second := func(c *Scope) {
		label := c.Remember(func() string { return "second" })
		c.Text(label)
	}
	flag := react.NewRef(true)
	comp := New(monospace())
	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			if flag.Get() {
				c.Component(first)
				return
			}
			c.Component(second)
		})
	})

	comp.Measure(Unbounded)
	flag.Set(false)
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "second@{0 0}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
}

func TestDisposeDetachesEverything(t *testing.T) {
	count := react.NewRef(0)
	invalidations := 0
	comp := New(monospace())
	comp.config.OnInvalidate = func() { invalidations++ }
	comp.SetContent(func(c *Scope) { c.Text(fmt.Sprintf("%d", count.Get())) })

	comp.Measure(Unbounded)
	comp.Dispose()
	count.Set(1)

	if invalidations != 0 {
		t.Fatalf("invalidations = %d after Dispose, want 0", invalidations)
	}
	if got := comp.Measure(Unbounded); got != (Size{}) {
		t.Fatalf("size = %v after Dispose, want the zero size", got)
	}
	comp.Draw(&Recorder{})
	comp.Dispose()
}

func TestSetContentReplacesTheTree(t *testing.T) {
	step := react.NewRef(0)
	comp := New(monospace())
	comp.SetContent(func(c *Scope) { c.Text("first") })
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "first@{0 0}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}

	comp.SetContent(func(c *Scope) {
		c.Column(func(c *Scope) {
			c.Text("second")
			c.Text(fmt.Sprintf("step %d", step.Get()))
		})
	})
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "second@{0 0}|step 0@{0 16}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
	step.Set(1)
	if got, want := strings.Join(texts(drawOps(comp)), "|"), "second@{0 0}|step 1@{0 16}"; got != want {
		t.Fatalf("texts = %q, want %q", got, want)
	}
}
