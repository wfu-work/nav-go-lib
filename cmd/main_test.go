package main

import "testing"

func TestIsHelpArg(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{name: "short flag", arg: "-h", want: true},
		{name: "long flag", arg: "--help", want: true},
		{name: "word", arg: "help", want: true},
		{name: "nav args", arg: "2025/12/07 00:00:00@2025/12/07 09:00:00@/data/raw@/data/nav@SITE@1@", want: false},
		{name: "empty", arg: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHelpArg(tt.arg); got != tt.want {
				t.Fatalf("isHelpArg(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestIsValidNavArgs(t *testing.T) {
	tests := []struct {
		name string
		argv string
		want bool
	}{
		{name: "valid", argv: "2025/12/07 00:00:00@2025/12/07 09:00:00@/data/raw@/data/nav@SITE@1@", want: true},
		{name: "missing ext separator", argv: "2025/12/07 00:00:00@2025/12/07 09:00:00@/data/raw@/data/nav@SITE@1", want: false},
		{name: "empty", argv: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidNavArgs(tt.argv); got != tt.want {
				t.Fatalf("isValidNavArgs(%q) = %v, want %v", tt.argv, got, tt.want)
			}
		})
	}
}
