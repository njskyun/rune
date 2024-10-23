package main

import (
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"unicode/utf8"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/bxelab/runestone"
	"lukechampine.com/uint128"
)

type Config struct {
	WalletName  string
	PrivateKey  string
	FeePerByte  int64
	UtxoAmount  int64
	IsAutoSpeed int64
	Network     string
	RpcUrl      string
	LocalRpcUrl string
	Etching     *struct {
		Rune              string
		Logo              string
		Symbol            *string
		Premine           *uint64
		Amount            *uint64
		Cap               *uint64
		Divisibility      *int
		HeightStart       *int
		HeightEnd         *int
		HeightOffsetStart *int
		HeightOffsetEnd   *int
	}
	Mint *struct {
		RuneId  string
		MintNum int64
	}
}

func DefaultConfig() Config {
	return Config{
		Network: "mainnet",
	}

}
func (c Config) GetFeePerByte() int64 {
	if c.FeePerByte == 0 {
		return 0
	}
	return c.FeePerByte
}

func (c Config) GetLocalRpcUrl() string {
	return c.LocalRpcUrl
}

func (c Config) GetIsAutoSpeed() int64 {
	return c.IsAutoSpeed
}

func (c Config) GetUtxoAmount() int64 {
	if c.UtxoAmount == 0 {
		return 330
	}
	return c.UtxoAmount
}

func (c Config) GetWalletName() string {
	return c.WalletName
}

func (c Config) GetEtching() (*runestone.Etching, error) {
	if c.Etching == nil {
		return nil, errors.New("Etching config is required")
	}
	if c.Etching.Rune == "" {
		return nil, errors.New("Rune is required")
	}
	if c.Etching.Symbol != nil {
		runeCount := utf8.RuneCountInString(*c.Etching.Symbol)
		if runeCount != 1 {
			return nil, errors.New("Symbol must be a single character")
		}
	}
	etching := &runestone.Etching{}
	r, err := runestone.SpacedRuneFromString(c.Etching.Rune)
	if err != nil {
		return nil, err
	}
	etching.Rune = &r.Rune
	etching.Spacers = &r.Spacers
	if c.Etching.Symbol != nil {
		symbolStr := *c.Etching.Symbol
		symbol := ([]rune(symbolStr))[0]
		etching.Symbol = &symbol
	}
	if c.Etching.Premine != nil {
		premine := uint128.From64(*c.Etching.Premine)
		etching.Premine = &premine
	}
	if c.Etching.Amount != nil {
		amount := uint128.From64(*c.Etching.Amount)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Amount = &amount
	}
	if c.Etching.Cap != nil {
		cap := uint128.From64(*c.Etching.Cap)
		etching.Terms.Cap = &cap
	}
	if c.Etching.Divisibility != nil {
		d := uint8(*c.Etching.Divisibility)
		etching.Divisibility = &d
	}
	if c.Etching.HeightStart != nil {
		h := uint64(*c.Etching.HeightStart)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Height[0] = &h
	}
	if c.Etching.HeightEnd != nil {
		h := uint64(*c.Etching.HeightEnd)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Height[1] = &h
	}
	if c.Etching.HeightOffsetStart != nil {
		h := uint64(*c.Etching.HeightOffsetStart)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Offset[0] = &h
	}
	if c.Etching.HeightOffsetEnd != nil {
		h := uint64(*c.Etching.HeightOffsetEnd)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Offset[1] = &h
	}
	return etching, nil
}
func (c Config) GetMint() (*runestone.RuneId, int64, error) {
	if c.Mint == nil {
		return nil, 0, errors.New("Mint config is required")
	}
	if c.Mint.RuneId == "" {
		return nil, 0, errors.New("RuneId is required")
	}
	if c.Mint.MintNum == 0 {
		return nil, 0, errors.New("MintNum is required")
	}

	runeId, err := runestone.RuneIdFromString(c.Mint.RuneId)
	if err != nil {
		return nil, 0, err
	}
	return runeId, c.Mint.MintNum, nil
}

func (c Config) GetNetwork() *chaincfg.Params {
	if c.Network == "mainnet" {
		return &chaincfg.MainNetParams
	}
	if c.Network == "testnet" {
		return &chaincfg.TestNet3Params
	}
	if c.Network == "regtest" {
		return &chaincfg.RegressionNetParams
	}
	if c.Network == "signet" {
		return &chaincfg.SigNetParams
	}
	panic("unknown network")
}

func (c Config) GetPrivateKeyAddr() (*btcec.PrivateKey, string, error) {
	if c.PrivateKey == "" {
		return nil, "", errors.New("PrivateKey is required")
	}
	pkBytes, err := hex.DecodeString(c.PrivateKey)
	if err != nil {
		return nil, "", err
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(pkBytes)
	if err != nil {
		return nil, "", err
	}
	tapKey := txscript.ComputeTaprootKeyNoScript(pubKey)
	addr, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(tapKey), c.GetNetwork(),
	)
	if err != nil {
		return nil, "", err
	}
	address := addr.EncodeAddress()
	return privKey, address, nil
}
func (c Config) GetRuneLogo() (mime string, data []byte) {
	if c.Etching != nil && c.Etching.Logo != "" {
		mime, err := getContentType(c.Etching.Logo)
		if err != nil {
			return "", nil
		}
		data, err := getFileBytes(c.Etching.Logo)
		if err != nil {
			return "", nil
		}
		return mime, data

	}
	return "", nil
}

func getContentType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)
	return contentType, nil
}
func getFileBytes(filePath string) ([]byte, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}
