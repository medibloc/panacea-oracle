package panacea

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec"
	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v4/modules/core/23-commitment/types"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/sgx"
	sgxdb "github.com/medibloc/panacea-oracle/store/sgxleveldb"
	log "github.com/sirupsen/logrus"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"
	"github.com/tendermint/tendermint/light/provider"
	tmhttp "github.com/tendermint/tendermint/light/provider/http"
	dbs "github.com/tendermint/tendermint/light/store/db"
	"github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

type QueryClient interface {
	Close() error
	GetAccount(context.Context, string) (authtypes.AccountI, error)
	GetDID(context.Context, string) (*didtypes.DIDDocumentWithSeq, error)
	GetOracleRegistration(context.Context, string, string) (*oracletypes.OracleRegistration, error)
	GetLightBlock(height int64) (*tmtypes.LightBlock, error)
	GetOracleParamsPublicKey(context.Context) (*btcec.PublicKey, error)
	GetDeal(context.Context, uint64) (*datadealtypes.Deal, error)
	GetConsent(context.Context, uint64, string) (*datadealtypes.Consent, error)
	GetLastBlockHeight(context.Context) (int64, error)
	GetCachedLastBlockHeight() int64
	GetOracleUpgrade(context.Context, string, string) (*oracletypes.OracleUpgrade, error)
	GetOracleUpgradeInfo(context.Context) (*oracletypes.OracleUpgradeInfo, error)
	GetOracle(context.Context, string) (*oracletypes.Oracle, error)
	VerifyTrustedBlockInfo(int64, []byte) error
}

const (
	trustedPeriod       = 2 * 7 * 24 * time.Hour
	refreshIntervalTime = time.Second * 3
)

type TrustedBlockInfo struct {
	TrustedBlockHeight int64
	TrustedBlockHash   []byte
}

var _ QueryClient = &verifiedQueryClient{}

type verifiedQueryClient struct {
	rpcClient             *rpchttp.HTTP
	lightClient           *light.Client
	db                    dbm.DB
	mutex                 *sync.Mutex
	cdc                   *codec.ProtoCodec
	aminoCdc              *codec.AminoCodec
	cachedLastBlockHeight int64
}

// makeInterfaceRegistry
func makeInterfaceRegistry() sdk.InterfaceRegistry {
	interfaceRegistry := sdk.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	return interfaceRegistry
}

// NewVerifiedQueryClient set verifiedQueryClient with rpcClient & and returns, if successful,
// a verifiedQueryClient that can be used to add query function.
func NewVerifiedQueryClient(ctx context.Context, config *config.Config, info *TrustedBlockInfo, sgx sgx.Sgx) (QueryClient, error) {
	return newVerifiedQueryClientWithSgxLevelDB(ctx, config, info, sgx)
}

func LoadVerifiedQueryClient(ctx context.Context, config *config.Config, sgx sgx.Sgx) (QueryClient, error) {
	return newVerifiedQueryClientWithSgxLevelDB(ctx, config, nil, sgx)
}

func newVerifiedQueryClientWithSgxLevelDB(ctx context.Context, config *config.Config, info *TrustedBlockInfo, sgx sgx.Sgx) (QueryClient, error) {
	db, err := sgxdb.NewSgxLevelDB("light-client", config.AbsDataDirPath(), sgx)
	if err != nil {
		return nil, err
	}
	return NewVerifiedQueryClientWithDB(ctx, config, info, db)
}

// NewVerifiedQueryClientWithDB creates a verifiedQueryClient using a provided DB.
// If TrustedBlockInfo exists, a new lightClient is created based on this information,
// and if TrustedBlockInfo is nil, a lightClient is created with information obtained from TrustedStore.
func NewVerifiedQueryClientWithDB(ctx context.Context, config *config.Config, info *TrustedBlockInfo, db dbm.DB) (QueryClient, error) {
	lcMutex := sync.Mutex{}
	chainID := config.Panacea.ChainID
	rpcClient, err := rpchttp.New(config.Panacea.RPCAddr, "/websocket")
	if err != nil {
		return nil, err
	}

	pv, err := tmhttp.New(chainID, config.Panacea.LightClientPrimaryAddr)
	if err != nil {
		return nil, err
	}

	var pvs []provider.Provider
	for _, witnessAddr := range config.Panacea.LightClientWitnessAddrs {
		witness, err := tmhttp.New(chainID, witnessAddr)
		if err != nil {
			return nil, err
		}
		pvs = append(pvs, witness)
	}

	store := dbs.New(db, chainID)

	var lc *light.Client
	logger := light.Logger(newTMLogger(config))

	if info == nil {
		lc, err = light.NewClientFromTrustedStore(
			chainID,
			trustedPeriod,
			pv,
			pvs,
			store,
			light.SkippingVerification(light.DefaultTrustLevel),
			logger,
		)
	} else {
		trustOptions := light.TrustOptions{
			Period: trustedPeriod,
			Height: info.TrustedBlockHeight,
			Hash:   info.TrustedBlockHash,
		}
		lc, err = light.NewClient(
			ctx,
			chainID,
			trustOptions,
			pv,
			pvs,
			store,
			light.SkippingVerification(light.DefaultTrustLevel),
			logger,
		)
	}

	if err != nil {
		return nil, err
	}

	queryClient := &verifiedQueryClient{
		rpcClient:   rpcClient,
		lightClient: lc,
		db:          db,
		mutex:       &lcMutex,
		cdc:         codec.NewProtoCodec(makeInterfaceRegistry()),
		aminoCdc:    codec.NewAminoCodec(codec.NewLegacyAmino()),
	}

	// If the last height is not present when oracle initially start the server, all queries will fail.
	// So if Oracle doesn't get the block information when it first starts the server, it will fail.
	if err := queryClient.lastBlockCaching(); err != nil {
		return nil, err
	}

	go queryClient.startSchedulingLastBlockCaching()

	return queryClient, nil
}

func newTMLogger(conf *config.Config) tmlog.Logger {
	logger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))

	switch strings.ToLower(conf.Panacea.LightClientLogLevel) {
	case "panic", "fatal", "error":
		logger = tmlog.NewFilter(logger, tmlog.AllowError())
	case "warn", "warning", "info":
		logger = tmlog.NewFilter(logger, tmlog.AllowInfo())
	default: // "debug", "trace", and so on
		logger = tmlog.NewFilter(logger, tmlog.AllowDebug())
	}

	return logger
}

func (q *verifiedQueryClient) lastBlockCaching() error {
	lastHeight, err := q.GetLastBlockHeight(context.Background())

	if err != nil {
		return fmt.Errorf("failed to refresh last block. %w", err)
	}
	log.Debugf("Refresh last block. Height(%d)", lastHeight)
	q.cachedLastBlockHeight = lastHeight

	return nil
}

// startSchedulingLastBlockCaching updates the LightClient with the last block information and stores the height of this block.
func (q *verifiedQueryClient) startSchedulingLastBlockCaching() {
	for {
		time.Sleep(refreshIntervalTime)

		if err := q.lastBlockCaching(); err != nil {
			log.Error(err)
		}
	}
}

func (q *verifiedQueryClient) safeUpdateLightClient(ctx context.Context) (*tmtypes.LightBlock, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.lightClient.Update(ctx, time.Now())
}

func (q *verifiedQueryClient) safeVerifyLightBlockAtHeight(ctx context.Context, height int64) (*tmtypes.LightBlock, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.lightClient.VerifyLightBlockAtHeight(ctx, height, time.Now())
}

// GetStoreData get data from panacea with storeKey and key, then verify queried data with light client and merkle proof.
// the returned data type is ResponseQuery.value ([]byte), so recommend to convert to expected type
func (q *verifiedQueryClient) GetStoreData(ctx context.Context, storeKey string, key []byte) ([]byte, error) {
	queryHeight := q.getQueryBlockHeight()

	//set queryOption prove to true
	option := client.ABCIQueryOptions{
		Prove:  true,
		Height: queryHeight,
	}
	// query to kv store with proof option
	result, err := q.abciQueryWithOptions(ctx, fmt.Sprintf("/store/%s/key", storeKey), key, option)
	if err != nil {
		return nil, err
	}

	// for merkle proof, the apphash of the next block is needed.
	// requests the next block every second, and generates an error after more than 12sec
	var nextTrustedBlock *tmtypes.LightBlock
	i := 0
	for {
		nextTrustedBlock, err = q.safeVerifyLightBlockAtHeight(ctx, queryHeight+1)
		if errors.Is(err, provider.ErrHeightTooHigh) {
			time.Sleep(1 * time.Second)
			i++
		} else if err != nil {
			return nil, err
		} else {
			break
		}
		if i > 12 {
			return nil, fmt.Errorf("can not get nextTrustedBlock")
		}
	}
	// verify query result with merkle proof & trusted block info
	merkleProof, err := types.ConvertProofs(result.Response.ProofOps)
	if err != nil {
		return nil, err
	}

	sdkSpecs := []*ics23.ProofSpec{ics23.IavlSpec, ics23.TendermintSpec}
	merkleRootKey := types.NewMerkleRoot(nextTrustedBlock.AppHash.Bytes())

	keyPath := url.PathEscape(string(key))
	merklePath := types.NewMerklePath(storeKey, keyPath)
	err = merkleProof.VerifyMembership(sdkSpecs, merkleRootKey, merklePath, result.Response.Value)
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}

func (q *verifiedQueryClient) getQueryBlockHeight() int64 {
	return q.GetCachedLastBlockHeight() - 1
}

func (q *verifiedQueryClient) Close() error {
	return q.db.Close()
}

// abciQueryWithOptions is a wrapper of rpcClient.ABCIQueryWithOptions,
// but validates the details of result.Response even if rpcClient.ABCIQueryWithOptions returns no error.
func (q *verifiedQueryClient) abciQueryWithOptions(ctx context.Context, path string, data tmbytes.HexBytes, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	res, err := q.rpcClient.ABCIQueryWithOptions(ctx, path, data, opts)
	if err != nil {
		return nil, err
	}
	resp := res.Response

	// Validate the response.
	if resp.IsErr() {
		return nil, fmt.Errorf("err response. code(%v) codeSpace(%v) log(%v)", resp.Code, resp.Codespace, resp.Log)
	}
	if len(resp.Key) == 0 {
		return nil, ErrEmptyKey
	}
	if len(resp.Value) == 0 {
		return nil, ErrEmptyValue
	}
	if opts.Prove && (resp.ProofOps == nil || len(resp.ProofOps.Ops) == 0) {
		return nil, errors.New("no proof ops")
	}
	if resp.Height <= 0 {
		return nil, ErrNegativeOrZeroHeight
	}

	return res, nil
}

func (q *verifiedQueryClient) GetLightBlock(height int64) (*tmtypes.LightBlock, error) {
	return q.safeVerifyLightBlockAtHeight(context.Background(), height)
}

// Below are query functions that use GetStoreData function to verify queried result.
// Need to set storeKey and key inside the query function, and change type to expected type.

// GetAccount returns account from address.
func (q *verifiedQueryClient) GetAccount(ctx context.Context, address string) (authtypes.AccountI, error) {
	acc, err := GetAccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	key := authtypes.AddressStoreKey(acc)
	bz, err := q.GetStoreData(ctx, authtypes.StoreKey, key)
	if err != nil {
		return nil, err
	}

	var account authtypes.AccountI
	err = q.cdc.UnmarshalInterface(bz, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (q *verifiedQueryClient) GetDID(ctx context.Context, did string) (*didtypes.DIDDocumentWithSeq, error) {
	key := append(didtypes.DIDKeyPrefix, []byte(did)...)
	bz, err := q.GetStoreData(ctx, didtypes.StoreKey, key)
	if err != nil {
		return nil, err
	}

	var didDoc didtypes.DIDDocumentWithSeq
	if err := q.cdc.UnmarshalLengthPrefixed(bz, &didDoc); err != nil {
		return nil, err
	}

	return &didDoc, nil
}

func (q *verifiedQueryClient) GetDeal(ctx context.Context, dealID uint64) (*datadealtypes.Deal, error) {
	key := datadealtypes.GetDealKey(dealID)

	bz, err := q.GetStoreData(ctx, datadealtypes.StoreKey, key)
	if err != nil {
		return nil, err
	}

	var deal datadealtypes.Deal
	if err = q.cdc.UnmarshalLengthPrefixed(bz, &deal); err != nil {
		return nil, err
	}

	return &deal, nil
}

func (q *verifiedQueryClient) GetConsent(ctx context.Context, dealID uint64, dataHash string) (*datadealtypes.Consent, error) {

	key := datadealtypes.GetConsentKey(dealID, dataHash)

	bz, err := q.GetStoreData(ctx, datadealtypes.StoreKey, key)
	if err != nil {
		return nil, err
	}

	var consent datadealtypes.Consent
	err = q.cdc.UnmarshalLengthPrefixed(bz, &consent)
	if err != nil {
		return nil, err
	}

	return &consent, nil
}

func (q *verifiedQueryClient) GetOracleRegistration(ctx context.Context, uniqueID, oracleAddr string) (*oracletypes.OracleRegistration, error) {
	acc, err := GetAccAddressFromBech32(oracleAddr)
	if err != nil {
		return nil, err
	}

	key := oracletypes.GetOracleRegistrationKey(uniqueID, acc)

	bz, err := q.GetStoreData(ctx, oracletypes.StoreKey, key)
	if err != nil {
		return nil, err
	}

	var oracleRegistration oracletypes.OracleRegistration
	err = q.cdc.UnmarshalLengthPrefixed(bz, &oracleRegistration)
	if err != nil {
		return nil, err
	}

	return &oracleRegistration, nil
}

func (q *verifiedQueryClient) GetOracle(ctx context.Context, oracleAddr string) (*oracletypes.Oracle, error) {
	acc, err := GetAccAddressFromBech32(oracleAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get AccAddr from Bech32 address(%s): %w", oracleAddr, err)
	}

	bz, err := q.GetStoreData(ctx, oracletypes.StoreKey, oracletypes.GetOracleKey(acc))
	if err != nil {
		return nil, fmt.Errorf("failed to get oracle from data: %w", err)
	}

	var oracle oracletypes.Oracle
	if err := q.cdc.UnmarshalLengthPrefixed(bz, &oracle); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oracle data: %w", err)
	}

	return &oracle, nil
}

func (q *verifiedQueryClient) GetOracleParamsPublicKey(ctx context.Context) (*btcec.PublicKey, error) {
	pubKeyBase64Bz, err := q.GetStoreData(ctx, paramstypes.StoreKey, append(append([]byte(oracletypes.StoreKey), '/'), oracletypes.KeyOraclePublicKey...))
	if err != nil {
		return nil, err
	}

	if pubKeyBase64Bz == nil {
		return nil, errors.New("the oracle public key's value is nil")
	}

	// If you get a value from params, you should not use protoCodec, but use legacyAmino.
	var pubKeyBase64 string
	err = q.aminoCdc.LegacyAmino.UnmarshalJSON(pubKeyBase64Bz, &pubKeyBase64)
	if err != nil {
		return nil, err
	}

	pubKeyBz, err := base64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 pubkey: %w", err)
	}

	return btcec.ParsePubKey(pubKeyBz, btcec.S256())
}

func (q *verifiedQueryClient) GetOracleUpgrade(ctx context.Context, uniqueID, oracleAddr string) (*oracletypes.OracleUpgrade, error) {
	acc, err := GetAccAddressFromBech32(oracleAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Bech32 address: %w", err)
	}

	key := oracletypes.GetOracleUpgradeKey(uniqueID, acc)

	bz, err := q.GetStoreData(ctx, oracletypes.StoreKey, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from panacea: %w", err)
	}

	var oracleUpgrade oracletypes.OracleUpgrade
	if err := q.cdc.UnmarshalLengthPrefixed(bz, &oracleUpgrade); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return &oracleUpgrade, nil
}

func (q *verifiedQueryClient) GetOracleUpgradeInfo(ctx context.Context) (*oracletypes.OracleUpgradeInfo, error) {
	bz, err := q.GetStoreData(ctx, oracletypes.StoreKey, oracletypes.OracleUpgradeInfoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracle upgrade info from panacea: %w", err)
	}

	var oracleUpgradeInfo oracletypes.OracleUpgradeInfo
	if err := q.cdc.UnmarshalLengthPrefixed(bz, &oracleUpgradeInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return &oracleUpgradeInfo, nil
}

// GetLastBlockHeight updates the lightClient with the latest block and gets the height of that latest block.
func (q *verifiedQueryClient) GetLastBlockHeight(ctx context.Context) (int64, error) {
	// get recent light block
	// if the latest block has already been updated, get LastTrustedHeight
	trustedBlock, err := q.safeUpdateLightClient(ctx)
	if err != nil {
		return 0, err
	}

	if trustedBlock == nil {
		return q.lightClient.LastTrustedHeight()
	}

	return trustedBlock.Height, nil
}

func (q *verifiedQueryClient) GetCachedLastBlockHeight() int64 {
	return q.cachedLastBlockHeight
}

func (q *verifiedQueryClient) VerifyTrustedBlockInfo(height int64, blockHash []byte) error {
	block, err := q.GetLightBlock(height)
	if err != nil {
		switch err {
		case provider.ErrLightBlockNotFound, provider.ErrHeightTooHigh:
			return fmt.Errorf("not found light block. %w", err)
		default:
			return err
		}
	}

	if !bytes.Equal(block.Hash().Bytes(), blockHash) {
		return fmt.Errorf("failed to verify trusted block information. height(%v), expected block hash(%s), got block hash(%s)",
			height,
			hex.EncodeToString(block.Hash().Bytes()),
			hex.EncodeToString(blockHash),
		)
	}

	return nil
}
