package orchestrator

type Orchestrator interface {
	Run(name string, args map[string]any) (any, error)
}

type DefaultOrchestrator struct {
}
