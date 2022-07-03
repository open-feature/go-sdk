package openfeature

// Hook Hooks are a mechanism whereby application developers can add arbitrary behavior to flag evaluation. They operate similarly to middleware in many web frameworks.
// https://github.com/open-feature/spec/blob/main/specification/flag-evaluation/hooks.md
type Hook interface {

}

type HookContext struct {
	flagKey string
	flagType string

}