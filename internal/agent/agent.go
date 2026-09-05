package tools

import (
	"agent-harness/internal/tools"
)

type Agent struct {
	registry *tools.ToolRegistry
}

func NewAgent(registry *tools.ToolRegistry) *Agent {
	return &Agent{
		registry: registry,
	}
}

func (a *Agent) ExecuteTool(name string, args map[string]any) (any, error) {
	tool, err := a.registry.Get(name)

	if err != nil {
		return nil, err
	}

	return tool.Execute(args)
}

func (a *Agent) Run(
	name string,
	args map[string]any,
) (any, error) {
	return a.ExecuteTool(name, args)
}
