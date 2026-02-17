package main

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestIdFormat(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{1, "%d"},
		{5, "%d"},
		{9, "%d"},
		{10, "%02d"},
		{50, "%02d"},
		{99, "%02d"},
		{100, "%03d"},
		{999, "%03d"},
	}
	for _, tt := range tests {
		got := idFormat(tt.count)
		if got != tt.want {
			t.Errorf("idFormat(%d) = %q, want %q", tt.count, got, tt.want)
		}
	}
}

func TestFormatPrefix(t *testing.T) {
	tests := []struct {
		name   string
		custom string
		id     int
		idFmt  string
		want   string
	}{
		{"default single digit", "", 3, "%d", "[3]"},
		{"default zero padded", "", 3, "%02d", "[03]"},
		{"default triple padded", "", 3, "%03d", "[003]"},
		{"custom prefix", "Worker-{id}", 5, "%d", "Worker-5"},
		{"custom prefix padded", "Worker-{id}", 5, "%02d", "Worker-05"},
		{"custom no placeholder", "Static", 1, "%d", "Static"},
		{"custom multiple placeholders", "{id}-{id}", 7, "%d", "7-7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPrefix(tt.custom, tt.id, tt.idFmt)
			if got != tt.want {
				t.Errorf("formatPrefix(%q, %d, %q) = %q, want %q",
					tt.custom, tt.id, tt.idFmt, got, tt.want)
			}
		})
	}
}

func TestColorCycling(t *testing.T) {
	if len(colors) != 36 {
		t.Errorf("expected 36 colors, got %d", len(colors))
	}

	// Verify colors wrap around correctly
	for i := 0; i < 72; i++ {
		got := colors[i%len(colors)]
		want := colors[i%36]
		if got != want {
			t.Errorf("color at index %d: got %q, want %q", i, got, want)
		}
	}
}

func TestStreamOutput(t *testing.T) {
	tests := []struct {
		name   string
		inst   *instance
		input  string
		expect []string
	}{
		{
			name:   "single line with color",
			inst:   &instance{id: 1, color: "\033[31m", prefix: "[1]"},
			input:  "hello world\n",
			expect: []string{"\033[31m[1] hello world\033[0m"},
		},
		{
			name:   "multiple lines",
			inst:   &instance{id: 2, color: "\033[32m", prefix: "[2]"},
			input:  "line one\nline two\n",
			expect: []string{"\033[32m[2] line one\033[0m", "\033[32m[2] line two\033[0m"},
		},
		{
			name:   "no color",
			inst:   &instance{id: 1, color: "", prefix: "[1]"},
			input:  "plain text\n",
			expect: []string{"[1] plain text\033[0m"},
		},
		{
			name:   "empty input",
			inst:   &instance{id: 1, color: "\033[31m", prefix: "[1]"},
			input:  "",
			expect: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var buf bytes.Buffer
			var wg sync.WaitGroup
			wg.Add(1)

			streamOutput(tt.inst, reader, &buf, &wg)
			wg.Wait()

			output := buf.String()
			for _, exp := range tt.expect {
				if !strings.Contains(output, exp) {
					t.Errorf("output missing expected string %q\ngot: %q", exp, output)
				}
			}

			// Verify line count
			if len(tt.expect) > 0 {
				lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
				if len(lines) != len(tt.expect) {
					t.Errorf("expected %d lines, got %d: %q", len(tt.expect), len(lines), output)
				}
			}
		})
	}
}
