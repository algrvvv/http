package internal

import "testing"

func TestHostAliases(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("AppData", dir)

	if err := saveHostAlias("myapp", "localhost:8000/api/v1/"); err != nil {
		t.Fatal(err)
	}
	checks := map[string]string{
		"myapp/login":         "http://localhost:8000/api/v1/login",
		"MYAPP/users?page=2":  "http://localhost:8000/api/v1/users?page=2",
		"myapp":               "http://localhost:8000/api/v1/",
		"http://myapp/login":  "http://myapp/login",
		"localhost:9000/test": "http://localhost:9000/test",
		"unknown/path":        "http://unknown/path",
	}
	for input, want := range checks {
		got, err := resolveHostAlias(input)
		if err != nil || got != want {
			t.Errorf("resolveHostAlias(%q) = %q, %v; want %q", input, got, err, want)
		}
	}

	if err := saveHostAlias("MYAPP", "https://example.test/new"); err != nil {
		t.Fatal(err)
	}
	got, err := resolveHostAlias("myapp/login")
	if err != nil || got != "https://example.test/new/login" {
		t.Fatalf("updated alias = %q, %v", got, err)
	}
	if err := saveHostAlias("../invalid", "localhost"); err == nil {
		t.Fatal("expected an error for invalid alias")
	}
}
