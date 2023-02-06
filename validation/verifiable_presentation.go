package validation

import (
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

	// TODO: validate VP with the PresentationDefinition

	return nil
}
