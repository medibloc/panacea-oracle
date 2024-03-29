package validation

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"

	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"

	"github.com/medibloc/vc-sdk/pkg/vc"
)

type didResolver interface {
	Resolve(did string, opts ...vdr.DIDMethodOption) (*did.DocResolution, error)
}

// ValidateVP validates verifiable presentation
func ValidateVP(vdr didResolver, vpBytes, pdBytes []byte) error {
	f, err := vc.NewFramework(vdr)
	if err != nil {
		return fmt.Errorf("failed to create a framework for VP verification: %w", err)
	}

	if _, err := f.VerifyPresentation(vpBytes, vc.WithPresentationDefinition(pdBytes)); err != nil {
		return fmt.Errorf("invalid VP: %w", err)
	}

	return nil
}
