package pubsub

import (
	"fmt"
	"reflect"
	"testing"

	ut "github.com/linkingthing/cement/unittest"
)

func TestSub(t *testing.T) {
	ps := New(1)
	ch1 := ps.Sub("t1")
	ch2 := ps.Sub("t1")
	ch3 := ps.Sub("t2")

	ps.Pub("hi", "t1")
	ps.Pub("hello", "t2")

	ps.Shutdown()

	checkContents(t, ch1, []string{"hi"})
	checkContents(t, ch2, []string{"hi"})
	checkContents(t, ch3, []string{"hello"})
}

func TestAddSub(t *testing.T) {
	ps := New(3)
	ch1 := ps.Sub("t1")
	ch2 := ps.Sub("t2")

	ps.Pub("hi1", "t1")
	ps.Pub("hi2", "t2")

	ps.AddSub(ch1, "t2", "t3")
	ps.Pub("hi3", "t2")
	ps.Pub("hi4", "t3")

	ps.Shutdown()

	checkContents(t, ch1, []string{"hi1", "hi3", "hi4"})
	checkContents(t, ch2, []string{"hi2", "hi3"})
}

func TestUnsub(t *testing.T) {
	ps := New(1)
	defer ps.Shutdown()

	ch := ps.Sub("t1")

	ps.Pub("hi", "t1")
	ps.Unsub(ch, "t1")
	checkContents(t, ch, []string{"hi"})
}

func TestUnsubAll(t *testing.T) {
	ps := New(1)
	ch1 := ps.Sub("t1", "t2", "t3")
	ch2 := ps.Sub("t1", "t3")

	ps.Unsub(ch1)
	checkContents(t, ch1, []string{})

	ps.Pub("hi", "t1")
	ps.Shutdown()

	checkContents(t, ch2, []string{"hi"})
}

func TestClose(t *testing.T) {
	ps := New(1)
	ch1 := ps.Sub("t1")
	ch2 := ps.Sub("t1")
	ch3 := ps.Sub("t2")
	ch4 := ps.Sub("t3")

	ps.Pub("hi", "t1")
	ps.Pub("hello", "t2")
	ps.Close("t1", "t2")

	checkContents(t, ch1, []string{"hi"})
	checkContents(t, ch2, []string{"hi"})
	checkContents(t, ch3, []string{"hello"})

	ps.Pub("welcome", "t3")
	ps.Shutdown()

	checkContents(t, ch4, []string{"welcome"})
}

func TestUnsubAfterClose(t *testing.T) {
	ps := New(1)
	ch := ps.Sub("t1")
	defer func() {
		ps.Unsub(ch, "t1")
		ps.Shutdown()
	}()

	ps.Close("t1")
	checkContents(t, ch, []string{})
}

func TestShutdown(t *testing.T) {
	ps := New(10)
	ch1 := ps.Sub("t1", "t2")
	ch2 := ps.Sub("t3")
	ps.Shutdown()
	_, ok := <-ch1
	ut.Assert(t, ok == false, "shutdown should close ch")

	_, ok = <-ch2
	ut.Assert(t, ok == false, "shutdown should close ch")
}

func TestMultiSub(t *testing.T) {
	ps := New(2)
	ch := ps.Sub("t1", "t2")

	ps.Pub("hi", "t1")
	ps.Pub("hello", "t2")
	ps.Shutdown()

	checkContents(t, ch, []string{"hi", "hello"})
}

func TestMultiPub(t *testing.T) {
	ps := New(2)
	ch1 := ps.Sub("t1")
	ch2 := ps.Sub("t2")

	ps.Pub("hi", "t1", "t2")
	ps.Shutdown()

	checkContents(t, ch1, []string{"hi"})
	checkContents(t, ch2, []string{"hi"})
}

func TestTryPub(t *testing.T) {
	ps := New(1)
	defer ps.Shutdown()

	ch := ps.Sub("t1")
	ps.TryPub("hi", "t1")
	ps.TryPub("there", "t1")

	<-ch
	extraMsg := false
	select {
	case <-ch:
		extraMsg = true
	default:
	}

	ut.Assert(t, extraMsg == false, "Extra message was found in channel")
}

func TestMultiUnsub(t *testing.T) {
	ps := New(1)
	defer ps.Shutdown()

	ch := ps.Sub("t1", "t2", "t3")

	ps.Unsub(ch, "t1")
	ps.Pub("hi", "t1")
	ps.Pub("hello", "t2")
	ps.Unsub(ch, "t2", "t3")

	checkContents(t, ch, []string{"hello"})
}

func TestMultiClose(t *testing.T) {
	ps := New(2)
	defer ps.Shutdown()

	ch := ps.Sub("t1", "t2")

	ps.Pub("hi", "t1")
	ps.Close("t1")

	ps.Pub("hello", "t2")
	ps.Close("t2")

	checkContents(t, ch, []string{"hi", "hello"})
}

func checkContents(t *testing.T, ch chan interface{}, vals []string) {
	contents := []string{}
	for v := range ch {
		contents = append(contents, v.(string))
	}

	ut.Assert(t, reflect.DeepEqual(contents, vals), fmt.Sprintf("Invalid channel contents. Expected: %v, but was: %v.", vals, contents))
}
