package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/ethclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	// Create your own private key and fund it. You can find a test avalanche faucet.
	key1, _   = // add private key here 
	addr1     = crypto.PubkeyToAddress(key1.PublicKey)
	toAddress = common.HexToAddress("0x766c47CFAB38A6A3813551b9e31BDAca455Bf262")
	clientURI = "https://avalanche-fuji-c-chain.publicnode.com"
	chainID   = new(big.Int).SetInt64(43113) // fuji
)

func main() {
	fmt.Println("Starting to send transaction")
	ctx := context.Background()
	client, err := ethclient.Dial(clientURI)
	if err != nil {
		fmt.Println("error with ethClient Dial: %w", err)
		return
	}
	currentNonce, err := client.NonceAt(ctx, addr1, nil)
	if err != nil {
		fmt.Println("error in getting nonce: %w", err)
		return
	}

	signedTx, err := createAndSignTx(ctx, &toAddress, currentNonce)
	if err != nil {
		fmt.Println("error in tx creation/signing: %w", err)
		return
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		fmt.Println("error with sending transaction: %w", err)
		return
	}

	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		select {
		case <-time.After(2 * time.Second):
		case <-sendCtx.Done():
			fmt.Println("failed to await tx %s nonce %d: %w", signedTx.Hash(), currentNonce, ctx.Err())
		}

		newNonce, err := client.NonceAt(ctx, addr1, nil)
		if err != nil {
			fmt.Println("error in getting nonce: %w", err)
		}

		fmt.Println("checking if transaction got accepted")
		if currentNonce < newNonce {
			fmt.Println("successful transaction")
			return
		}
	}
}

func createAndSignTx(ctx context.Context, toAddress *common.Address, nonce uint64) (*types.Transaction, error) {
	signer := types.LatestSignerForChainID(chainID)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID, // hardcode fuji testnet for now
		Nonce:     nonce,
		To:        toAddress,
		Gas:       21000,                   //hardcode gas transaction for now
		GasFeeCap: big.NewInt(25000000000), //hardcode base fee for now
		GasTipCap: big.NewInt(0),
		Data:      []byte{},
		Value:     big.NewInt(0),
	})

	signedTx, err := types.SignTx(tx, signer, key1)
	if err != nil {
		return nil, errors.New("failed to sign transaction")
	}

	return signedTx, nil
}
