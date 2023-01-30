package status

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (s *statusService) GetStatus(w http.ResponseWriter, _ *http.Request) {
	resp := statusResponse{
		OracleAccountAddress: s.OracleAcc().GetAddress(),
		API: statusAPI{
			ListenAddr: s.Config().API.ListenAddr,
		},
		EnclaveInfo: statusEnclaveInfo{
			//ProductIDBase64: base64.StdEncoding.EncodeToString(s.EnclaveInfo().ProductID),
			//UniqueIDBase64:  base64.StdEncoding.EncodeToString(s.EnclaveInfo().UniqueID),
		},
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("failed to marshal response: %v", err)
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonResp); err != nil {
		log.Errorf("failed to write response: %s", err.Error())
		return
	}
}
