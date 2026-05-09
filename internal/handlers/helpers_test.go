package handlers

import (
	"strings"
	"testing"
)

func TestSanitizeWords(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "replaces kerfuffle case insensitively",
			in:   "Kerfuffle is bad",
			want: "**** is bad",
		},
		{
			name: "replaces sharbert",
			in:   "sharbert here",
			want: "**** here",
		},
		{
			name: "replaces fornax",
			in:   "FORNAX mention",
			want: "**** mention",
		},
		{
			name: "multiple banned words",
			in:   "kerfuffle and fornax",
			want: "**** and ****",
		},
		{
			name: "normal text unchanged",
			in:   "hello world",
			want: "hello world",
		},
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "word as substring is not replaced",
			in:   "prekerfuffle",
			want: "prekerfuffle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeWords(tt.in)
			if got != tt.want {
				t.Errorf("sanitizeWords(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidateChirp(t *testing.T) {
	longBody := strings.Repeat("a", 141)

	tests := []struct {
		name    string
		body    string
		wantOK  bool
		wantMsg string
		wantOut string // expected sanitized chirp body when valid
	}{
		{
			name:    "too long",
			body:    longBody,
			wantOK:  false,
			wantMsg: "Chirp is too long",
		},
		{
			name:    "max length accepted",
			body:    strings.Repeat("a", 140),
			wantOK:  true,
			wantOut: strings.Repeat("a", 140),
		},
		{
			name:    "sanitizes and valid",
			body:    "hello kerfuffle",
			wantOK:  true,
			wantOut: "hello ****",
		},
		{
			name:    "empty body valid after sanitize",
			body:    "",
			wantOK:  true,
			wantOut: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateChirp(tt.body)
			if got.isValid != tt.wantOK {
				t.Errorf("validateChirp(%q).isValid = %v; want %v", tt.body, got.isValid, tt.wantOK)
			}
			if !tt.wantOK && got.msg != tt.wantMsg {
				t.Errorf("validateChirp(%q).msg = %q; want %q", tt.body, got.msg, tt.wantMsg)
			}
			if tt.wantOK && got.chirp != tt.wantOut {
				t.Errorf("validateChirp(%q).chirp = %q; want %q", tt.body, got.chirp, tt.wantOut)
			}
		})
	}
}
