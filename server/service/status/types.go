package status

// statusResponse is a response type for GET /v0/status.
// Not using existing structs directly, such as sgx.EnclaveInfo, since they may contain sensitive values in the future.
// TODO: feel free to add more values, such as version
type statusResponse struct {
	OracleAccountAddress string            `json:"oracle_account_address"`
	API                  statusAPI         `json:"api"`
	EnclaveInfo          statusEnclaveInfo `json:"enclave_info"`
}

type statusAPI struct {
	ListenAddr string `json:"listen_addr"`
}

type statusEnclaveInfo struct {
	ProductIDBase64 string `json:"product_id_base64"`
	UniqueIDBase64  string `json:"unique_id_base64"`
}
