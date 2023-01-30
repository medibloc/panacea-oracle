package validation

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/medibloc/vc-sdk/pkg/vc"
)

func ValidateVerifiablePresentation(vpBytes []byte) error {
	framework, err := vc.NewFrameWork()
	if err != nil {
		return err
	}

	proofs, err := framework.GetPresentationProofs(vpBytes)
	if err != nil {
		return err
	}

	var vp verifiable.Presentation

	err = json.Unmarshal(vpBytes, &vp)
	if err != nil {
		return err
	}

	vc.ProofOptions{}

	err = validatePresentationProofs(proofs, )
	if err != nil {
		return err
	}

	framework.VerifyPresentation()

	return nil
}

func validatePresentationProofs(proofs *vc.ProofIterator, vp *verifiable.Proof) error {
	for !proofs.HasNext() {
		proof := proofs.Next()

		if proof.VerificationMethod !=
	}
	return nil
}
