package httpapi_test

import (
	"os"
	"strings"
	"testing"

	"example.com/ch09-catalog-api/internal/httpapi"
)

// specOperations reads the "paths:" block of the OpenAPI file with plain string handling (two-space
// indentation: a path at depth 1, its methods at depth 2) and returns "METHOD /path" entries.
func specOperations(t *testing.T, spec string) map[string]bool {
	t.Helper()
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
	ops := specOperations(t, readSpec(t))
	for _, r := range httpapi.Routes {
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

func TestSpecDefinesProblemDetails(t *testing.T) {
	spec := readSpec(t)
	for _, want := range []string{
		"\n    Problem:\n",                // components.schemas.Problem
		"required: [type, title, status]", // the RFC 9457 members every problem carries
		"application/problem+json",        // the media type of every error response
		"$ref: \"#/components/schemas/Problem\"",
	} {
		if !strings.Contains(spec, want) {
			t.Errorf("api/openapi.yaml lacks %q: add the problem details schema (chapter 9, section 9.3)", want)
		}
	}
}
