package react

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestRefNotifiesOnChange(t *testing.T) {
	count := NewRef(0)
	runs := 0
	e := NewEffect(func() {
		runs++
		_ = count.Get()
	}, nil)

	if runs != 1 {
		t.Fatalf("effect ran %d times on creation, want 1", runs)
	}
	count.Set(1)
	if runs != 2 {
		t.Fatalf("effect ran %d times after a write, want 2", runs)
	}
	count.Set(1)
	if runs != 2 {
		t.Fatalf("effect ran %d times after an equal write, want 2", runs)
	}
	e.Stop()
	count.Set(2)
	if runs != 2 {
		t.Fatalf("stopped effect ran again, %d runs", runs)
	}
	if got := count.Peek(); got != 2 {
		t.Fatalf("count = %d, want 2", got)
	}
}

func TestRefComparesSlicesByIdentity(t *testing.T) {
	values := NewRef([]int{1})
	runs := 0
	e := NewEffect(func() {
		runs++
		_ = values.Get()
	}, nil)

	values.Set(values.Peek())
	if runs != 1 {
		t.Fatalf("writing the same slice re-ran the effect, %d runs", runs)
	}
	values.Set([]int{1})
	if runs != 2 {
		t.Fatalf("writing another slice did not re-run the effect, %d runs", runs)
	}
	e.Stop()

	byContent := NewRefEqual([]int{1}, func(newValue, oldValue []int) bool {
		return len(newValue) == len(oldValue) && newValue[0] == oldValue[0]
	})
	e = NewEffect(func() {
		runs++
		_ = byContent.Get()
	}, nil)
	byContent.Set([]int{1})
	if runs != 3 {
		t.Fatalf("content-equal write re-ran the effect, %d runs", runs)
	}
	e.Stop()
}

func TestPeekAndUntrackDoNotSubscribe(t *testing.T) {
	count := NewRef(0)
	runs := 0
	e := NewEffect(func() {
		runs++
		_ = count.Peek()
		_ = Untrack(count.Get)
	}, nil)

	count.Set(1)
	if runs != 1 {
		t.Fatalf("an untracked read subscribed the effect, %d runs", runs)
	}
	e.Stop()
}

func TestEffectCleanupRunsBeforeRerun(t *testing.T) {
	source := NewRef(0)
	var events []string
	e := NewEffect(func() {
		value := source.Get()
		events = append(events, fmt.Sprintf("run %d", value))
		OnCleanup(func() { events = append(events, fmt.Sprintf("cleanup %d", value)) })
	}, nil)

	source.Set(1)
	e.Stop()

	if got, want := strings.Join(events, "|"), "run 0|cleanup 0|run 1|cleanup 1"; got != want {
		t.Fatalf("events = %q, want %q", got, want)
	}
}

func TestEffectDoesNotReenterItself(t *testing.T) {
	count := NewRef(0)
	runs := 0
	e := NewEffect(func() {
		runs++
		value := count.Get()
		if value < 3 {
			count.Set(value + 1)
		}
	}, nil)

	if runs != 1 {
		t.Fatalf("the effect re-entered itself, %d runs", runs)
	}
	if got := count.Peek(); got != 1 {
		t.Fatalf("count = %d, want 1", got)
	}
	count.Set(10)
	if runs != 2 {
		t.Fatalf("effect ran %d times, want 2", runs)
	}
	e.Stop()
}

func TestEffectCanStopItself(t *testing.T) {
	source := NewRef(0)
	runs := 0
	var e *Effect
	e = NewEffect(func() {
		runs++
		source.Get()
		if runs == 2 {
			e.Stop()
		}
	}, nil)

	source.Set(1)
	source.Set(2)
	if runs != 2 {
		t.Fatalf("effect ran %d times after stopping itself, want 2", runs)
	}
}

func TestPanicInEffectLeavesTheSystemUsable(t *testing.T) {
	source := NewRef(0)
	crash := NewEffect(func() {
		if source.Get() != 0 {
			panic("boom")
		}
	}, nil)
	runs := 0
	other := NewEffect(func() {
		runs++
		_ = source.Get()
	}, nil)

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("the effect did not panic")
			}
		}()
		source.Set(1)
	}()
	crash.Stop()

	source.Set(2)
	source.Set(3)
	if runs != 3 {
		t.Fatalf("other effect ran %d times, want 3", runs)
	}
	other.Stop()
}

func TestBatchDefersReruns(t *testing.T) {
	first, second := NewRef(0), NewRef(0)
	runs := 0
	e := NewEffect(func() {
		runs++
		_, _ = first.Get(), second.Get()
	}, nil)

	Batch(func() {
		first.Set(1)
		second.Set(2)
		if runs != 1 {
			t.Fatalf("the effect ran inside the batch, %d runs", runs)
		}
	})
	if runs != 2 {
		t.Fatalf("the effect ran %d times after the batch, want 2", runs)
	}
	e.Stop()
}

func TestNextTickRunsAfterTheBatch(t *testing.T) {
	source := NewRef(0)
	var order []string
	e := NewEffect(func() {
		source.Get()
		order = append(order, "effect")
	}, nil)

	order = nil
	Batch(func() {
		source.Set(1)
		NextTick(func() { order = append(order, "tick") })
	})
	if got := strings.Join(order, "|"); got != "effect|tick" {
		t.Fatalf("order = %q, want %q", got, "effect|tick")
	}

	order = nil
	NextTick(func() { order = append(order, "tick") })
	if got := strings.Join(order, "|"); got != "tick" {
		t.Fatalf("idle NextTick order = %q, want %q", got, "tick")
	}
	e.Stop()
}

func TestComputedIsLazyAndCached(t *testing.T) {
	calls := 0
	source := NewRef(1)
	double := NewComputed(func() int {
		calls++
		return source.Get() * 2
	})

	if calls != 0 {
		t.Fatalf("the computed ran before it was read, %d calls", calls)
	}
	if got := double.Get(); got != 2 {
		t.Fatalf("double = %d, want 2", got)
	}
	if got := double.Get(); got != 2 {
		t.Fatalf("double = %d, want 2", got)
	}
	if calls != 1 {
		t.Fatalf("the cached value was recomputed, %d calls", calls)
	}
	source.Set(2)
	if calls != 1 {
		t.Fatalf("the computed ran with no reader, %d calls", calls)
	}
	if got := double.Get(); got != 4 {
		t.Fatalf("double = %d, want 4", got)
	}
	if calls != 2 {
		t.Fatalf("the computed ran %d times, want 2", calls)
	}
}

func TestComputedNotifiesOnlyOnRealChange(t *testing.T) {
	source := NewRef(1)
	parity := NewComputed(func() int { return source.Get() % 2 })
	var seen []int
	e := NewEffect(func() { seen = append(seen, parity.Get()) }, nil)

	source.Set(3)
	if len(seen) != 1 {
		t.Fatalf("seen = %v, an unchanged computed value woke the effect", seen)
	}
	source.Set(4)
	if len(seen) != 2 || seen[1] != 0 {
		t.Fatalf("seen = %v, want [1 0]", seen)
	}
	e.Stop()
}

func TestComputedChains(t *testing.T) {
	price := NewRef(10)
	quantity := NewRef(2)
	total := NewComputed(func() int { return price.Get() * quantity.Get() })
	doubled := NewComputed(func() int { return total.Get() * 2 })

	var seen []int
	e := NewEffect(func() { seen = append(seen, doubled.Get()) }, nil)

	quantity.Set(3)
	price.Set(20)
	if want := []int{40, 60, 120}; fmt.Sprint(seen) != fmt.Sprint(want) {
		t.Fatalf("seen = %v, want %v", seen, want)
	}
	e.Stop()
}

func TestComputedStopKeepsLastValue(t *testing.T) {
	source := NewRef(1)
	calls := 0
	double := NewComputed(func() int {
		calls++
		return source.Get() * 2
	})

	if got := double.Get(); got != 2 {
		t.Fatalf("double = %d, want 2", got)
	}
	double.Stop()
	source.Set(5)
	if calls != 1 {
		t.Fatalf("a stopped computed recomputed, %d calls", calls)
	}
	if got := double.Get(); got != 2 {
		t.Fatalf("double = %d, want the last value 2", got)
	}
}

func TestWatchReportsChanges(t *testing.T) {
	source := NewRef(1)
	var changes []string
	e := Watch(func() int { return source.Get() }, func(newValue, oldValue int) {
		changes = append(changes, fmt.Sprintf("%d->%d", oldValue, newValue))
	}, nil)

	if len(changes) != 0 {
		t.Fatalf("the callback ran without Immediate, %v", changes)
	}
	source.Set(2)
	source.Set(2)
	source.Set(3)
	e.Stop()
	source.Set(4)

	if got, want := strings.Join(changes, "|"), "1->2|2->3"; got != want {
		t.Fatalf("changes = %q, want %q", got, want)
	}
}

func TestWatchImmediateAndCleanup(t *testing.T) {
	source := NewRef(0)
	var events []string
	e := Watch(func() int { return source.Get() }, func(newValue, oldValue int) {
		events = append(events, fmt.Sprintf("cb %d->%d", oldValue, newValue))
		OnCleanup(func() { events = append(events, fmt.Sprintf("cleanup %d", newValue)) })
	}, &WatchConfig{Immediate: true})

	source.Set(1)
	source.Set(1)
	e.Stop()

	want := "cb 0->0|cleanup 0|cb 0->1|cleanup 1"
	if got := strings.Join(events, "|"); got != want {
		t.Fatalf("events = %q, want %q", got, want)
	}
}

func TestWatchPostFlushOrder(t *testing.T) {
	source := NewRef(0)
	var order []string
	watch := WatchEffect(func() {
		source.Get()
		order = append(order, "watch")
	}, FlushPost)
	effect := NewEffect(func() {
		source.Get()
		order = append(order, "effect")
	}, nil)

	order = nil
	Batch(func() { source.Set(1) })
	if got := strings.Join(order, "|"); got != "effect|watch" {
		t.Fatalf("order = %q, want %q", got, "effect|watch")
	}

	watch.Stop()
	effect.Stop()
}

func TestEffectScopeStopsEverything(t *testing.T) {
	source := NewRef(0)
	runs, disposed, childRuns := 0, 0, 0
	scope := NewEffectScope()
	scope.Run(func() {
		NewEffect(func() {
			runs++
			source.Get()
		}, nil)
		OnScopeDispose(func() { disposed++ })
		child := NewEffectScope()
		child.Run(func() {
			NewEffect(func() {
				childRuns++
				source.Get()
			}, nil)
		})
	})
	outside := NewEffect(func() { source.Get() }, nil)

	source.Set(1)
	if runs != 2 || childRuns != 2 {
		t.Fatalf("runs = %d, childRuns = %d, want 2 each", runs, childRuns)
	}
	scope.Stop()
	if disposed != 1 {
		t.Fatalf("disposed = %d, want 1", disposed)
	}
	scope.Stop()
	if disposed != 1 {
		t.Fatalf("second Stop disposed again, %d", disposed)
	}
	source.Set(2)
	if runs != 2 || childRuns != 2 {
		t.Fatalf("scope effects ran after Stop: runs = %d, childRuns = %d", runs, childRuns)
	}
	if !outside.Active() {
		t.Fatal("Stop stopped an effect created outside the scope")
	}
	outside.Stop()
}

func TestEffectScopeRunDoesNothingAfterStop(t *testing.T) {
	scope := NewEffectScope()
	scope.Stop()
	ran := false
	scope.Run(func() {
		ran = true
	})
	if ran {
		t.Fatal("Run executed the function of a stopped scope")
	}
}

func TestReactiveMapTracksKeys(t *testing.T) {
	m := NewReactiveMap[string, int]()
	runs := 0
	e := NewEffect(func() {
		runs++
		_ = m.Get("a")
	}, nil)

	m.Set("b", 1)
	if runs != 1 {
		t.Fatalf("writing another key re-ran the effect, %d runs", runs)
	}
	m.Set("a", 1)
	if runs != 2 {
		t.Fatalf("writing the key did not re-run the effect, %d runs", runs)
	}
	m.Set("a", 1)
	if runs != 2 {
		t.Fatalf("equal write re-ran the effect, %d runs", runs)
	}
	m.Set("a", 2)
	if runs != 3 {
		t.Fatalf("changed write did not re-run the effect, %d runs", runs)
	}
	m.Delete("a")
	if runs != 4 {
		t.Fatalf("deleting the key did not re-run the effect, %d runs", runs)
	}
	m.Delete("a")
	if runs != 4 {
		t.Fatalf("deleting a missing key re-ran the effect, %d runs", runs)
	}
	e.Stop()
}

func TestReactiveMapTracksSizeAndEntries(t *testing.T) {
	m := NewReactiveMap[string, int]()
	m.Set("a", 1)
	length, values := 0, 0
	e := NewEffect(func() {
		length++
		_ = m.Len()
	}, nil)

	m.Set("a", 2)
	if length != 1 {
		t.Fatalf("updating a value re-ran the length reader, %d runs", length)
	}
	m.Set("b", 3)
	if length != 2 {
		t.Fatalf("adding a key did not re-run the length reader, %d runs", length)
	}
	e.Stop()

	entries := NewEffect(func() {
		values++
		_ = m.Values()
	}, nil)
	m.Set("a", 9)
	if values != 2 {
		t.Fatalf("updating a value did not re-run the entries reader, %d runs", values)
	}
	m.Clear()
	if values != 3 {
		t.Fatalf("clearing did not re-run the entries reader, %d runs", values)
	}
	if got := m.Len(); got != 0 {
		t.Fatalf("Len() = %d after Clear, want 0", got)
	}
	entries.Stop()
}

func TestReactiveSliceTracksIndices(t *testing.T) {
	s := NewReactiveSlice(1, 2, 3)
	runs := 0
	e := NewEffect(func() {
		runs++
		_ = s.Get(0)
	}, nil)

	s.Set(1, 9)
	if runs != 1 {
		t.Fatalf("writing another index re-ran the effect, %d runs", runs)
	}
	s.Set(0, 9)
	if runs != 2 {
		t.Fatalf("writing the index did not re-run the effect, %d runs", runs)
	}
	s.Set(0, 9)
	if runs != 2 {
		t.Fatalf("equal write re-ran the effect, %d runs", runs)
	}
	e.Stop()
}

func TestReactiveSliceTracksStructure(t *testing.T) {
	s := NewReactiveSlice[int]()
	runs := 0
	e := NewEffect(func() {
		runs++
		_ = s.Len()
	}, nil)

	s.Append(1)
	s.Insert(0, 0)
	s.Remove(0)
	if runs != 4 {
		t.Fatalf("runs = %d, want one per structural change", runs)
	}
	if got := s.Slice(); fmt.Sprint(got) != "[1]" {
		t.Fatalf("Slice() = %v, want [1]", got)
	}
	s.Clear()
	if runs != 5 {
		t.Fatalf("Clear did not re-run the reader, %d runs", runs)
	}
	e.Stop()
}

func TestReactiveSliceShiftsIndices(t *testing.T) {
	s := NewReactiveSlice("a", "b", "c")
	var seen []string
	e := NewEffect(func() { seen = append(seen, s.Get(0)) }, nil)

	s.Insert(0, "z")
	s.Remove(0)
	if got := strings.Join(seen, "|"); got != "a|z|a" {
		t.Fatalf("seen = %q, want %q", got, "a|z|a")
	}
	e.Stop()
}

func TestReactiveSliceKeepsIndicesAcrossShrinkAndGrow(t *testing.T) {
	s := NewReactiveSlice("a", "b")
	var seen []string
	e := NewEffect(func() {
		if s.Len() < 2 {
			seen = append(seen, "-")
			return
		}
		seen = append(seen, s.Get(1))
	}, nil)

	s.Remove(1)
	s.Append("c")
	if got := strings.Join(seen, "|"); got != "b|-|c" {
		t.Fatalf("seen = %q, want %q", got, "b|-|c")
	}
	e.Stop()
}

func TestSame(t *testing.T) {
	nan := math.NaN()
	if !same[float64]()(nan, nan) {
		t.Fatal("NaN must equal NaN")
	}
	if same[float64]()(nan, 1) || same[float64]()(1, 2) {
		t.Fatal("distinct floats must differ")
	}
	if !same[int]()(1, 1) {
		t.Fatal("equal ints must be the same")
	}

	s := []int{1, 2}
	if !same[[]int]()(s, s) {
		t.Fatal("the same slice must be the same value")
	}
	if same[[]int]()(s, []int{1, 2}) {
		t.Fatal("a new slice must be a change")
	}
	m := map[string]int{"a": 1}
	if !same[map[string]int]()(m, m) || same[map[string]int]()(m, map[string]int{"a": 1}) {
		t.Fatal("maps must compare by identity")
	}

	type withAny struct{ Value any }
	if !same[withAny]()(withAny{Value: 1}, withAny{Value: 1}) {
		t.Fatal("structs holding comparable values must be the same value")
	}
	if same[withAny]()(withAny{Value: s}, withAny{Value: s}) {
		t.Fatal("structs holding an uncomparable value must count as changed")
	}
	type record struct {
		N    int
		Name string
	}
	if !same[record]()(record{1, "a"}, record{1, "a"}) {
		t.Fatal("equal structs must be the same value")
	}

	if !same[any]()(1, 1) || !same[any]()(nan, nan) {
		t.Fatal("boxed ints and NaN must compare like their static types")
	}
	if !same[any]()(any(s), any(s)) || same[any]()(any(s), any([]int{1, 2})) {
		t.Fatal("boxed slices must compare by identity")
	}
	if same[any]()(1, "1") {
		t.Fatal("values of different dynamic types must differ")
	}
	f := func() {}
	if same[func()]()(f, f) {
		t.Fatal("funcs are never the same value")
	}
}
