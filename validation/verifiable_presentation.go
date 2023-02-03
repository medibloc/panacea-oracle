package validation

import (
	"fmt"

	"github.com/medibloc/vc-sdk/pkg/vc"
)

func ValidateVerifiablePresentation(vpBytes, pubKey []byte) error {
	framework, err := vc.NewFramework()
	if err != nil {
		return err
	}

	err = framework.VerifyPresentation(vpBytes, pubKey, "EcdsaSecp256k1VerificationKey2019")
	if err != nil {
		return err
	}

	proofs, err := framework.GetPresentationProofs(vpBytes)
	if err != nil {
		return err
	}

	err = validateVPProofs(proofs)
	if err != nil {
		return err
	}

	// TODO: validate VP with the PresentationDefinition

	return nil
}

func validateVPProofs(proofs *vc.ProofIterator) error {
	for proofs.HasNext() {
		proof := proofs.Next()
		if proof.VerificationMethod == "" {
			return fmt.Errorf("verification method is empty")
		}
		if proof.Type == "" {
			return fmt.Errorf("proof type is empty")
		}
		if proof.ProofPurpose == "" {
			return fmt.Errorf("proof purpose is empty")
		}
		if proof.Domain == "" {
			return fmt.Errorf("proof domain is empty")
		}
		if proof.Challenge == "" {
			return fmt.Errorf("proof challenge is empty")
		}
		if proof.Created == "" {
			return fmt.Errorf("proof created date is empty")
		}
	}

	return nil
}
