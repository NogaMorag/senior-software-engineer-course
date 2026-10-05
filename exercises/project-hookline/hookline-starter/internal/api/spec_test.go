package api_test

import (
	"os"
	"strings"
	"testing"

	"example.com/hookline/internal/api"
)

// specOperations reads the "paths:" block of the OpenAPI file with plain string handling (two-space
// indentation: a path at depth 1, its methods at depth 2) and returns "METHOD /path" entries.
func specOperations(spec string) map[string]bool {
	ops := map[string]bool{}
	inPaths, path := false, ""
	for _, line := range strings.Split(spec, "\n") {
		switch {
		case line == "paths:":
			inPaths = true
		case inPaths && line != "" && !strings.HasPrefix(line, " "):
			inPaths = false
		case inPaths && strings.HasPrefix(line, "  /"):
			path = strings.TrimSuffix(strings.TrimSpace(line), ":")
		case inPaths && path != "" && strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "     "):
			m := strings.TrimSuffix(strings.TrimSpace(line), ":")
			switch m {
			case "get", "post", "put", "patch", "delete":
				ops[strings.ToUpper(m)+" "+path] = true
			}
		}
	}
	return ops
}

func readSpec(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	return string(b)
}

func TestSpecCoversEveryRoute(t *testing.T) {
	ops := specOperations(readSpec(t))
	for _, r := range api.Routes {
		key := r.Method + " " + r.Path
		if !ops[key] {
			t.Errorf("route %s is served but not in api/openapi.yaml", key)
		}
		delete(ops, key)
	}
	for op := range ops {
		t.Errorf("api/openapi.yaml documents %s but the server does not register it", op)
	}
}

func TestSpecDefinesProblemsAndAuth(t *testing.T) {
	spec := readSpec(t)
	for _, want := range []string{
		"openapi: 3.1.",
		"\n    Problem:\n",
		"required: [type, title, status]",
		"application/problem+json",
		"scheme: bearer",
		"Idempotency-Key",
		"idempotency-key-reused",
	} {
		if !strings.Contains(spec, want) {
			t.Errorf("api/openapi.yaml lacks %q", want)
		}
	}
}
