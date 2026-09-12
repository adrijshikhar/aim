package runner

import (
	"fmt"
	"testing"
)

func makeMockEnviron(count int) []string {
	env := make([]string, count)
	for i := 0; i < count; i++ {
		env[i] = fmt.Sprintf("VAR_%03d=value_%03d_some_longer_path_data", i, i)
	}
	return env
}

func BenchmarkBuildEnv_Typical(b *testing.B) {
	environ := makeMockEnviron(50)
	launchEnv := map[string]string{
		"HOME":        "/Users/nemesis/.aim/profiles/bench",
		"AIM_PROFILE": "bench",
		"AIM_AGENT":   "agy",
		"AIM_HOME":    "/Users/nemesis/.aim",
		"PS1":         "[aim:agy:bench] $ ",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res := BuildEnv(environ, launchEnv)
		if len(res) < 50 {
			b.Fatalf("expected at least 50 variables, got %d", len(res))
		}
	}
}

func BenchmarkBuildEnv_Large(b *testing.B) {
	environ := makeMockEnviron(150)
	launchEnv := map[string]string{
		"HOME":           "/Users/nemesis/.aim/profiles/bench",
		"AIM_PROFILE":    "bench",
		"AIM_AGENT":      "agy",
		"AIM_HOME":       "/Users/nemesis/.aim",
		"PS1":            "[aim:agy:bench] $ ",
		"SSH_CONNECTION": "127.0.0.1 50000 127.0.0.1 22",
		"CUSTOM_KEY_1":   "custom_value_1",
		"CUSTOM_KEY_2":   "custom_value_2",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res := BuildEnv(environ, launchEnv)
		if len(res) < 150 {
			b.Fatalf("expected at least 150 variables, got %d", len(res))
		}
	}
}
