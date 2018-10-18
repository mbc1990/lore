package main

import (
	"testing"
)

func TestParseUserID(t *testing.T) {
	t.Parallel()

	tt := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "Standard",
			input:    "<@U123ABC>",
			expected: "U123ABC",
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			out := parseUserID(tc.input)

			if out != tc.expected {
				t.Fatalf("expected: '%v', got: '%v'", tc.expected, out)
			}
		})
	}
}
