package validation

import (
	"fmt"
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

	frameWork, err := vc.NewFrameWork()
	require.NoError(t, err)

	privKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	vcBytes, err := frameWork.SignCredential([]byte(cred), privKey.Serialize(), &vc.ProofOptions{
		VerificationMethod: "did:panacea:BFbUAkxqj3cXXYdNK9FAF9UuEmm7jCT5T77rXhBCvy2K#key1",
		SignatureType:      "EcdsaSecp256k1Signature2019",
	})
	require.NoError(t, err)
	fmt.Println(string(vcBytes))

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

	err = frameWork.VerifyCredential(vcBytes, privKey.PubKey().SerializeUncompressed(), "EcdsaSecp256k1VerificationKey2019")
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

	err = ValidateVerifiablePresentation(vpBytes, privKey.PubKey().SerializeUncompressed())
	require.NoError(t, err)
}
