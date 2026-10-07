package jev

type Decision string

const (
	Allow   Decision = "allow"
	Confirm Decision = "confirm"
	Deny    Decision = "deny"
)

type DecisionResult struct {
	Decision Decision
	Reason   string
}
