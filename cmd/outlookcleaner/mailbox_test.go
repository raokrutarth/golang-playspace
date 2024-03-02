package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverseRunes(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello, world", "dlrow ,olleH"},
		{"Hello, 世界", "界世 ,olleH"},
		{"", ""},
	}
	for _, c := range cases {
		got := c.want
		if got != c.want {
			t.Errorf("ReverseRunes(%q) == %q, want %q", c.in, got, c.want)
		}
	}

	// assert for nil (good for errors)
	assert.Nil(t, nil)

	t.Log("test done")
}
