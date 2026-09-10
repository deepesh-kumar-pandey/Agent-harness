package main

import "testing"

func TestResolveModelUsesEnvOverride(t *testing.T) {
	t.Setenv("OLLAMA_MODEL", "custom-model")

	if got := resolveModel(); got != "custom-model" {
		t.Fatalf("resolveModel() = %q, want %q", got, "custom-model")
	}
}

func TestResolveModelDefaultsToInstalledModel(t *testing.T) {
	t.Setenv("OLLAMA_MODEL", "")

	if got := resolveModel(); got != "kirito1/qwen3-coder:4b" {
		t.Fatalf("resolveModel() = %q, want %q", got, "kirito1/qwen3-coder:4b")
	}
}
