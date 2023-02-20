package mocks

type MockBroadcastTxResponse struct {
	code        int64
	description string
	error       error
}

func NewMockBroadcastTx(code int64, description string, error error) *MockBroadcastTxResponse {
	return &MockBroadcastTxResponse{
		code:        code,
		description: description,
		error:       error,
	}
}
