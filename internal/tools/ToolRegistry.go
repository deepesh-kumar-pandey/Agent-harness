package tools

import (
	"fmt"
)

type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

func (tr *ToolRegistry) Register(name string, tool Tool) (string, Tool) {
	fmt.Printf("Registering tool %s\n", name)
	tr.tools[name] = tool
	return name, tool
}

func (tr *ToolRegistry) Get(name string) (Tool, error) {
	tool, exists := tr.tools[name]

	if !exists {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	return tool, nil
}
