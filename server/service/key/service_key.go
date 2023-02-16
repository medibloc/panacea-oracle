package key

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/crypto"
	key "github.com/medibloc/panacea-oracle/pb/key/v0"
	"github.com/medibloc/panacea-oracle/server/rpc/interceptor/auth"
	log "github.com/sirupsen/logrus"
)

func (s *secretKeyService) GetSecretKey(ctx context.Context, req *key.GetSecretKeyRequest) (*key.GetSecretKeyResponse, error) {
	queryClient := s.GetQueryClient()
	oraclePrivKey := s.GetOraclePrivKey()

	dealID := req.DealId
	dataHashStr := hex.EncodeToString(req.DataHash)

	requesterAddress, err := auth.GetRequestAddress(ctx)
	if err != nil {
		log.Errorf("failed to get request address. %v", err.Error())
		return nil, fmt.Errorf("failed to get request address. %w", err)
	}

	deal, err := queryClient.GetDeal(ctx, dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal(%d): %w", dealID, err)
	}

	if requesterAddress != deal.ConsumerAddress {
		return nil, fmt.Errorf("only consumer request secret key")
	}

	_, err = queryClient.GetConsent(ctx, dealID, dataHashStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get consent(dealID: %d, dataHash %s). %w", dealID, dataHashStr, err)
	}

	consumerAcc, err := queryClient.GetAccount(ctx, deal.ConsumerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer account: %w", err)
	}
	consumerPubKeyBz := consumerAcc.GetPubKey().Bytes()
	consumerPubKey, err := btcec.ParsePubKey(consumerPubKeyBz, btcec.S256())
	if err != nil {
		return nil, fmt.Errorf("failed to parse consumer public key: %w", err)
	}

	sharedKey := crypto.DeriveSharedKey(oraclePrivKey, consumerPubKey, crypto.KDFSHA256)

	secretKey := GetSecretKey(oraclePrivKey.Serialize(), dealID, req.DataHash)
	encryptedSecretKey, err := crypto.Encrypt(sharedKey, nil, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret key with shared key: %w", err)
	}

	return &key.GetSecretKeyResponse{
		EncryptedSecretKey: encryptedSecretKey,
	}, nil
}
