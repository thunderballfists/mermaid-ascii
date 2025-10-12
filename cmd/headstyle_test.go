package cmd

import "testing"

func TestParseHeadStyle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected HeadStyle
		wantErr  bool
	}{
		{name: "default empty", input: "", expected: HeadStylePointer},
		{name: "default whitespace", input: "  ", expected: HeadStylePointer},
		{name: "pointer case insensitive", input: "PoInTeR", expected: HeadStylePointer},
		{name: "open", input: "open", expected: HeadStyleOpen},
		{name: "none", input: "none", expected: HeadStyleNone},
		{name: "wide", input: "wide", expected: HeadStyleWide},
		{name: "invalid", input: "triangle", expected: HeadStylePointer, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseHeadStyle(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q but got none", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestHeadStyleString(t *testing.T) {
	mapping := map[HeadStyle]string{
		HeadStylePointer: "pointer",
		HeadStyleOpen:    "open",
		HeadStyleNone:    "none",
		HeadStyleWide:    "wide",
	}

	for style, want := range mapping {
		got := headStyleString(style)
		if got != want {
			t.Fatalf("expected %q for style %v, got %q", want, style, got)
		}
	}
}
