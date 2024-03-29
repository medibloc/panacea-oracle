package consumer_service

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/medibloc/panacea-oracle/panacea"
)

type FileStorage interface {
	Add(endpoint string, dealID uint64, dataHash string, data []byte) error
}

var _ FileStorage = &ConsumerServiceFileStorage{}

type ConsumerServiceFileStorage struct {
	oraclePrivKey *btcec.PrivateKey
	oracleAcc     *panacea.OracleAccount
	timeout       time.Duration
}

func NewConsumerService(oraclePrivKey *btcec.PrivateKey, oracleAcc *panacea.OracleAccount, timeout time.Duration) FileStorage {
	return &ConsumerServiceFileStorage{
		oraclePrivKey: oraclePrivKey,
		oracleAcc:     oracleAcc,
		timeout:       timeout,
	}
}

func (s *ConsumerServiceFileStorage) Add(endpoint string, dealID uint64, dataHash string, data []byte) error {
	// dataUrl is /v0/deals/{dealId}/data/{dataHash}
	dataUrl := endpoint + "/v0/deals/" + strconv.FormatUint(dealID, 10) + "/data/" + dataHash
	token, err := generateJWT(s.oraclePrivKey, s.oracleAcc, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to generate jwt: %v", err)
	}
	if err := s.postData(data, dataUrl, token); err != nil {
		return fmt.Errorf("failed to post request: %v", err)
	}

	return nil
}

func (s *ConsumerServiceFileStorage) postData(data []byte, dataUrl string, jwt []byte) error {
	buff := bytes.NewBuffer(data)

	request, err := http.NewRequest("POST", dataUrl, buff)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+string(jwt))

	client := http.Client{
		Timeout: s.timeout,
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
	return nil
}

func generateJWT(privKey *btcec.PrivateKey, oracleAcc *panacea.OracleAccount, expiration time.Duration) ([]byte, error) {
	now := time.Now().Truncate(time.Second)
	token, err := jwt.NewBuilder().
		Issuer(oracleAcc.GetAddress()).
		IssuedAt(now).
		NotBefore(now).
		Expiration(now.Add(expiration)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build jwt error: %v", err)
	}
	signedJWT, err := jwt.Sign(token, jwt.WithKey(jwa.ES256K, privKey.ToECDSA()))
	if err != nil {
		return nil, fmt.Errorf("jwt signing error: %v", err)
	}
	return signedJWT, nil
}
