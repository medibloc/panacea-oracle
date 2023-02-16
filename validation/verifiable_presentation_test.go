package validation

import (
	"fmt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/vc-sdk/pkg/vc"
	"github.com/stretchr/testify/require"
)

func TestValidateVerifiablePresentation(t *testing.T) {
	cred := `{"@context": ["https://www.w3.org/2018/credentials/v1","https://www.w3.org/2018/credentials/examples/v1"],
	"issuer": "did:panacea:BFbUAkxqj3cXXYdNK9FAF9UuEmm7jCT5T77rXhBCvy2K",
	"id": "https://abc.com/1",
	"issuanceDate": "2010-01-01T19:13:24Z",
    "credentialSubject": {
      "id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
		"degree": {
		  "type": "BachelorDegree",
		  "name": "Bachelor of Science and Arts"
		}
    },
    "type": [
      "VerifiableCredential",
      "UniversityDegreeCredential"
    ]}`

	privKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	mockVDR := NewMockVDR(privKey.PubKey().SerializeUncompressed(), "EcdsaSecp256k1VerificationKey2019")
	frameWork, err := vc.NewFramework(mockVDR)
	require.NoError(t, err)

	vcBytes, err := frameWork.SignCredential([]byte(cred), privKey.Serialize(), &vc.ProofOptions{
		VerificationMethod: "did:panacea:BFbUAkxqj3cXXYdNK9FAF9UuEmm7jCT5T77rXhBCvy2K#key1",
		SignatureType:      "EcdsaSecp256k1Signature2019",
	})
	require.NoError(t, err)

	proofs, err := frameWork.GetCredentialProofs(vcBytes)
	require.NoError(t, err)
	require.True(t, proofs.HasNext())
	proof := proofs.Next()
	require.NotNil(t, proof)
	require.Equal(t, "did:panacea:BFbUAkxqj3cXXYdNK9FAF9UuEmm7jCT5T77rXhBCvy2K#key1", proof.VerificationMethod)
	require.Equal(t, "EcdsaSecp256k1Signature2019", proof.Type)
	require.Equal(t, "assertionMethod", proof.ProofPurpose)
	require.Empty(t, proof.Domain)
	require.Empty(t, proof.Challenge)
	require.NotEmpty(t, proof.Created) // automatically set as current time by PanaceaFramework
	require.False(t, proofs.HasNext())
	require.Nil(t, proofs.Next())

	err = frameWork.VerifyCredential(vcBytes)
	require.NoError(t, err)

	pres := fmt.Sprintf(`{"@context": ["https://www.w3.org/2018/credentials/v1"],
		"id": "https://abc.com/vp/1",
		"type": ["VerifiablePresentation"],
		"verifiableCredential": [%s]
	}`, string(vcBytes))

	vpBytes, err := frameWork.SignPresentation([]byte(pres), privKey.Serialize(), &vc.ProofOptions{
		VerificationMethod: "did:panacea:BFbUAkxqj3cXXYdNK9FAF9UuEmm7jCT5T77rXhBCvy2K#key1",
		SignatureType:      "EcdsaSecp256k1Signature2019",
		Domain:             "https://my-domain.com",
		Challenge:          "this is a challenge",
		Created:            "2017-06-18T21:19:10Z",
	})
	require.NoError(t, err)

	err = ValidateVP(mockVDR, vpBytes, nil)
	require.NoError(t, err)
}

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

func (v *MockVDR) Create(_ string, _ *did.Doc, _ ...vdr.DIDMethodOption) (*did.DocResolution, error) {
	return nil, nil
}

func (v *MockVDR) Update(_ *did.Doc, _ ...vdr.DIDMethodOption) error {
	return nil
}

func (v *MockVDR) Deactivate(_ string, _ ...vdr.DIDMethodOption) error {
	return nil
}

func (v *MockVDR) Close() error {
	return nil
}
