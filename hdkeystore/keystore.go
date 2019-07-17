package hdkeystore

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"

	"yekaiwallet/util"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type HDkeyStore struct {
	KeysDirPath     string
	ScryptN         int
	ScryptP         int
	PrivateKeyECDSA *ecdsa.PrivateKey
}

func NewKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey) *keystore.Key {
	id := utils.NewRandom()
	key := &keystore.Key{
		Id:         []byte(id),
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return key
}

func NewHDKeyStore(dirPath string, privateKeyECDSA *ecdsa.PrivateKey) *HDkeyStore {
	return &HDkeyStore{
		KeysDirPath:     dirPath,
		ScryptN:         keystore.LightScryptN,
		ScryptP:         keystore.LightScryptP,
		PrivateKeyECDSA: privateKeyECDSA,
	}
}

func (ks *HDkeyStore) StoreKey(address, auth string) error {

	key := NewKeyFromECDSA(ks.PrivateKeyECDSA)
	filename := ks.JoinPath(address)

	keyjson, err := keystore.EncryptKey(key, auth, ks.ScryptN, ks.ScryptP)
	if err != nil {
		return err
	}
	return utils.WriteKeyFile(filename, keyjson)
}

func (ks *HDkeyStore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(ks.KeysDirPath, filename)
}

func (ks *HDkeyStore) GetKey(addr common.Address, filename, auth string) (*keystore.Key, error) {
	// Load the key from the keystore and decrypt its contents
	keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(keyjson, auth)
	if err != nil {
		return nil, err
	}

	// Make sure we're really operating on the requested key (no swap attacks)
	if key.Address != addr {
		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Address, addr)
	}

	ks.PrivateKeyECDSA = key.PrivateKey
	return key, nil
}

// SignTx implements accounts.Wallet, which allows the account to sign an Ethereum transaction.
func (ks *HDkeyStore) SignTx(account common.Address, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {

	fmt.Printf("%+v\n", ks)
	// Sign the transaction and verify the sender to avoid hardware fault surprises
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, ks.PrivateKeyECDSA)
	if err != nil {
		return nil, err
	}

	msg, err := signedTx.AsMessage(types.HomesteadSigner{})
	if err != nil {
		return nil, err
	}

	sender := msg.From()
	if sender != account {
		return nil, fmt.Errorf("signer mismatch: expected %s, got %s", account.Hex(), sender.Hex())
	}

	return signedTx, nil
}
