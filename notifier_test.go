package van

import "testing"

func TestNotify(t *testing.T) {
	notify("foo")
	notify("bar")
}
