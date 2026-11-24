package viiperdevice

import "testing"

type rawState []byte

func (r rawState) MarshalBinary() ([]byte, error) { return []byte(r), nil }

// A full send buffer must drop the oldest pending state, never the newest:
// losing the newest one leaves the game stuck on an old input (e.g. stick held forward).
func TestQueueStateSendKeepsNewestWhenFull(t *testing.T) {
	d := &Device{
		stateChan: make(chan []byte, 3),
		done:      make(chan struct{}),
	}
	for _, s := range []string{"a", "b", "c", "d", "released"} {
		d.state = rawState(s)
		d.QueueStateSend()
	}

	var got []string
	for len(d.stateChan) > 0 {
		got = append(got, string(<-d.stateChan))
	}
	if len(got) != 3 || got[len(got)-1] != "released" {
		t.Fatalf("pending states = %v, want the last one to be \"released\"", got)
	}
}
