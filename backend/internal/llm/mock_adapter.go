package llm

import "context"

// MockInterpreter is a test double for the Interpreter interface.
type MockInterpreter struct {
	Result     *InterpretationResult
	Err        error
	CalledWith *InterpretationInput
}

// Interpret records the input and returns the preconfigured result or error.
func (m *MockInterpreter) Interpret(_ context.Context, input *InterpretationInput) (*InterpretationResult, error) {
	m.CalledWith = input
	return m.Result, m.Err
}
