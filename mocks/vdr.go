package mocks

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
)

type MockVDR struct {
	pubKeyBz   []byte
	pubKeyType string
}

func NewMockVDR(pubKeyBz []byte, pubKeyType string) *MockVDR {
	return &MockVDR{
		pubKeyBz:   pubKeyBz,
		pubKeyType: pubKeyType,
	}
}

func (v *MockVDR) Resolve(didID string, _ ...vdr.DIDMethodOption) (*did.DocResolution, error) {
	signingKey := did.VerificationMethod{
		ID:         didID + "#key1",
		Type:       v.pubKeyType,
		Controller: didID,
		Value:      v.pubKeyBz,
	}

	return &did.DocResolution{
		DIDDocument: &did.Doc{
			Context:            []string{"https://w3id.org/did/v1"},
			ID:                 didID,
			VerificationMethod: []did.VerificationMethod{signingKey},
		},
	}, nil
}
