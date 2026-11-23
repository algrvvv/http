package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBody(t *testing.T) {
	const payload = `{"items":[{"enabled":true}]}`
	path := filepath.Join(t.TempDir(), "payload.json")
	if err := os.WriteFile(path, []byte(payload), 0600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		args      []string
		body      string
		bodySet   bool
		jsonMode  bool
		want      string
		wantError bool
	}{
		{name: "fields", args: []string{"message=hello", "count=2", "detail=a=b", "empty="}, jsonMode: true, want: `{"count":"2","detail":"a=b","empty":"","message":"hello"}`},
		{name: "literal", body: payload, bodySet: true, jsonMode: true, want: payload},
		{name: "file", body: "@" + path, bodySet: true, jsonMode: true, want: payload},
		{name: "stdin", body: "@-", bodySet: true, jsonMode: true, want: payload},
		{name: "missing JSON flag", args: []string{"message=hello"}, wantError: true},
		{name: "conflicting input", args: []string{"message=hello"}, bodySet: true, jsonMode: true, wantError: true},
		{name: "invalid field", args: []string{"message"}, jsonMode: true, wantError: true},
		{name: "duplicate field", args: []string{"message=hello", "message=bye"}, jsonMode: true, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBody(tt.args, tt.body, tt.bodySet, tt.jsonMode, strings.NewReader(payload))
			if (err != nil) != tt.wantError || string(got) != tt.want {
				t.Fatalf("got (%q, %v), want (%q, error=%v)", got, err, tt.want, tt.wantError)
			}
		})
	}
}

func TestGetHeaders(t *testing.T) {
	input := stringList{"X-Test: a:b,c", "x-test: second"}
	headers, err := input.getHeaders()
	if err != nil {
		t.Fatal(err)
	}
	values := headers.Values("X-Test")
	if len(values) != 2 || values[0] != "a:b,c" || values[1] != "second" {
		t.Fatalf("unexpected header values: %q", values)
	}

	invalid := stringList{"X-Test: value", "invalid"}
	if _, err := invalid.getHeaders(); err == nil {
		t.Fatal("expected an error for malformed header")
	}
}
