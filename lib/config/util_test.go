package config

import (
	"flag"
	"slices"
	"testing"
)

// withCommandLine swaps the global flag set, so a test can declare and parse flags in isolation.
func withCommandLine(t *testing.T) {
	t.Helper()

	orig := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(t.Name(), flag.ContinueOnError)
	t.Cleanup(func() { flag.CommandLine = orig })
}

func TestStringListVarUsesEnvDefault(t *testing.T) {
	withCommandLine(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", " https://a.example , ,https://b.example ")

	var origins StringList
	StringListVar(&origins, "corsAllowedOrigins", "CORS_ALLOWED_ORIGINS", "", "")

	if err := flag.CommandLine.Parse(nil); err != nil {
		t.Fatalf("Can't parse flags: %v", err)
	}

	want := []string{"https://a.example", "https://b.example"}
	if !slices.Equal(origins, want) {
		t.Fatalf("Expected %q, got %q", want, origins)
	}
}

func TestStringListVarFlagOverridesEnv(t *testing.T) {
	withCommandLine(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://from-env.example")

	var origins StringList
	StringListVar(&origins, "corsAllowedOrigins", "CORS_ALLOWED_ORIGINS", "", "")

	// The env default must be replaced, not extended: an allow list merged from two sources is
	// wider than either of them.
	if err := flag.CommandLine.Parse([]string{"-corsAllowedOrigins", "https://from-flag.example"}); err != nil {
		t.Fatalf("Can't parse flags: %v", err)
	}

	want := []string{"https://from-flag.example"}
	if !slices.Equal(origins, want) {
		t.Fatalf("Expected %q, got %q", want, origins)
	}
}

func TestStringListVarRepeatedFlagAppends(t *testing.T) {
	withCommandLine(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://from-env.example")

	var origins StringList
	StringListVar(&origins, "corsAllowedOrigins", "CORS_ALLOWED_ORIGINS", "", "")

	args := []string{"-corsAllowedOrigins", "https://a.example,https://b.example", "-corsAllowedOrigins", "https://c.example"}
	if err := flag.CommandLine.Parse(args); err != nil {
		t.Fatalf("Can't parse flags: %v", err)
	}

	// Repeated occurrences accumulate, but the env default is still dropped by the first one.
	want := []string{"https://a.example", "https://b.example", "https://c.example"}
	if !slices.Equal(origins, want) {
		t.Fatalf("Expected %q, got %q", want, origins)
	}
}

func TestStringListVarEmptyValue(t *testing.T) {
	withCommandLine(t)

	var origins StringList
	StringListVar(&origins, "corsAllowedOrigins", "CORS_ALLOWED_ORIGINS", "", "")

	if err := flag.CommandLine.Parse(nil); err != nil {
		t.Fatalf("Can't parse flags: %v", err)
	}

	if len(origins) != 0 {
		t.Fatalf("Expected no origins, got %q", origins)
	}
}

func TestStringListString(t *testing.T) {
	var empty *StringList
	if got := empty.String(); got != "" {
		t.Fatalf("Expected an empty string for a nil list, got %q", got)
	}

	origins := StringList{"https://a.example", "https://b.example"}
	if got := origins.String(); got != "https://a.example,https://b.example" {
		t.Fatalf("Unexpected string representation %q", got)
	}
}
