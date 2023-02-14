package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/proof"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/vc-sdk/pkg/vc"
)

const tempVerificationMethod = "did:panacea:76e12ec712ebc6f1c221ebfeb1f#key1"

// ValidateData validates the data in the form of verifiable presentation
// This function validates the belows:
// 1. proof of VP
// 2. if the VP meets the requirements of PD
// 3. proof of VC
// 4. if the holder of VP is the one of VC
func ValidateData(vpBytes []byte, queryClient panacea.QueryClient) error {
	ctx := context.Background()

	holderDID := strings.Split(tempVerificationMethod, "#")[0]
	holderDIDDoc, err := queryClient.GetDID(ctx, holderDID)
	if err != nil {
		return err
	}

	holderVerificationMethod, err := GetVerificationMethod(holderDIDDoc, tempVerificationMethod)
	if err != nil {
		return err
	}

	holderPubKey := base58.Decode(holderVerificationMethod.PublicKeyBase58)
	if err != nil {
		return err
	}

	framework, err := vc.NewFramework()
	if err != nil {
		return err
	}

	pres, err := framework.VerifyPresentation(vpBytes, holderPubKey, "EcdsaSecp256k1VerificationKey2019", nil)
	if err != nil {
		return err
	}

	presHolder := pres.Holder

	creds := pres.Credentials()

	for _, cred := range creds {
		c, ok := cred.(verifiable.Credential)
		if !ok {
			return err
		}

		issuerVerificationMethod := c.Proofs.(proof.Proof).VerificationMethod

		issuerID := c.Issuer.ID
		issuerDIDDoc, err := queryClient.GetDID(ctx, issuerID)
		if err != nil {
			return err
		}

		issuerVerificationMethod, err := GetVerificationMethod(issuerDIDDoc, issuerID)
		if err != nil {
			return err
		}

		issuerPubKey := base58.Decode(issuerVerificationMethod.PublicKeyBase58)

		mCred, err := c.MarshalJSON()
		if err != nil {
			return err
		}

		if err := framework.VerifyCredential(mCred, issuerPubKey, "EcdsaSecp256k1VerificationKey2019"); err != nil {
			return err
		}
	}

	// TODO: validate VP with the PresentationDefinition

	return nil
}

func GetVerificationMethod(seq *didtypes.DIDDocumentWithSeq, verificationMethod string) (*didtypes.VerificationMethod, error) {
	for _, method := range seq.Document.VerificationMethods {
		if method.Id == verificationMethod {
			return method, nil
		}
	}

	return nil, fmt.Errorf("verification method not found")
}
