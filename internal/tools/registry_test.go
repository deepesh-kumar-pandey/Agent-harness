package tools

import (
	"reflect"
	"testing"
)

func TestToolRegistryGet(t *testing.T) {
	registry := NewToolRegistry()
	registeredTool := &Calculator{}
	registry.Register("custom", registeredTool)

	actualTool, err := registry.Get("custom")
	if err != nil {
		t.Fatalf("Get returned an unexpected error: %v", err)
	}
	if actualTool != registeredTool {
		t.Fatalf("Get returned %T, want the registered tool", actualTool)
	}

	missingTool, err := registry.Get("missing")
	if err == nil {
		t.Fatal("Get returned nil error for a missing tool")
	}
	if missingTool != nil {
		t.Fatalf("Get returned %T for a missing tool, want nil", missingTool)
	}
}

func TestToolRegistryNew(t *testing.T) {
	registry := NewToolRegistry()
	for _, name := range []string{"calculator", "shell", "filesystem"} {
		if !registry.Has(name) {
			t.Errorf("NewToolRegistry did not register %q", name)
		}
	}
}

func TestToolRegistryRegister(t *testing.T) {
	registry := &ToolRegistry{tools: make(map[string]Tool)}
	firstTool := &Calculator{}
	secondTool := &Calculator{}

	name, returnedTool := registry.Register("calculator", firstTool)
	if name != "calculator" || returnedTool != firstTool {
		t.Fatalf("Register returned (%q, %T), want (%q, firstTool)", name, returnedTool, "calculator")
	}

	registry.Register("calculator", secondTool)
	actualTool, err := registry.Get("calculator")
	if err != nil {
		t.Fatalf("Get after replacement returned an unexpected error: %v", err)
	}
	if actualTool != secondTool {
		t.Fatalf("replacement did not update the registered tool")
	}
}

func TestToolRegistryHas(t *testing.T) {
	registry := &ToolRegistry{tools: make(map[string]Tool)}
	if registry.Has("calculator") {
		t.Fatal("Has returned true for an unregistered tool")
	}

	registry.Register("calculator", &Calculator{})
	if !registry.Has("calculator") {
		t.Fatal("Has returned false for a registered tool")
	}
}

func TestToolRegistryList(t *testing.T) {
	registry := &ToolRegistry{tools: make(map[string]Tool)}
	if actual := registry.List(); len(actual) != 0 {
		t.Fatalf("List on an empty registry returned %v", actual)
	}

	for _, name := range []string{"calculator", "shell", "filesystem"} {
		registry.Register(name, &Calculator{})
	}

	actual := registry.List()
	if !reflect.DeepEqual(stringSet(actual), stringSet([]string{"calculator", "shell", "filesystem"})) {
		t.Fatalf("List returned %v, want all registered names", actual)
	}
}

func TestToolRegistryRemove(t *testing.T) {
	registry := &ToolRegistry{tools: make(map[string]Tool)}
	registry.Register("calculator", &Calculator{})

	if err := registry.Remove("calculator"); err != nil {
		t.Fatalf("Remove returned an unexpected error: %v", err)
	}
	if registry.Has("calculator") {
		t.Fatal("removed tool is still registered")
	}
	if _, err := registry.Get("calculator"); err == nil {
		t.Fatal("Get returned nil error for a removed tool")
	}
	if err := registry.Remove("calculator"); err == nil {
		t.Fatal("Remove returned nil error for a missing tool")
	}
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}
