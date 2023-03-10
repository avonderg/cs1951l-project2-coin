package miner

import (
	"Coin/pkg/block"
	"bytes"
	"context"
	"fmt"
	"math"
	"time"
)

// Mine When asked to mine, the miner selects the transactions
// with the highest priority to add to the mining pool.
func (m *Miner) Mine() *block.Block {
	//TODO
	if m.TxPool.PriorityMet() { // just return if false instead
		m.Mining.Store(true)             // set mining to true
		m.MiningPool = m.NewMiningPool() // create new pool

		// create new coinbase tx
		cbTx := m.GenerateCoinbaseTransaction(m.MiningPool)
		// make list of all transactions in mining pool in addition to cbTx
		// pass in cbTx into append
		txs := []*block.Transaction{cbTx}
		// need to add the other transactions in the miner pool

		block := block.New(m.PreviousHash, txs, m.DifficultyTarget)

		// calculate nonce and check if it is true or not
	}
	return nil
}

// CalculateNonce finds a winning nonce for a block. It uses context to
// know whether it should quit before it finds a nonce (if another block
// was found). ASICSs are optimized for this task.
func (m *Miner) CalculateNonce(ctx context.Context, b *block.Block) bool {
	nonce := uint32(0)
	target := m.DifficultyTarget

	for nonce < m.Config.NonceLimit {
		// check if another found
		select {
		case <-ctx.Done():
			return false // quit if another block was found
		//
		default:
			nonce += 1 // decrease the nonce by what factor?
			b.Header.Nonce = nonce

			hash := []byte(b.Hash()) // convert hash to byte array

			// does it meet the difficulty target?
			if bytes.Compare(hash, target) < 0 {
				return true // nonce is less than the difficulty target, exit the loop
			}
		}
	}
	return true
}

// GenerateCoinbaseTransaction generates a coinbase
// transaction based off the transactions in the mining pool.
// It does this by combining the fee reward to the minting reward,
// and sending that sum to itself.
func (m *Miner) GenerateCoinbaseTransaction(txs []*block.Transaction) *block.Transaction {
	//TODO

	inpSum, _ := m.getInputSums(txs)
	var outSum []uint32
	var fee uint32

	for i, tx := range txs {
		outSum[i] = tx.SumOutputs()
	}
	for i, sum := range inpSum {
		fee += sum - outSum[i] // aggregate
	}
	reward := m.CalculateMintingReward() + fee // add fee reward to minting reward
	new_tx := &block.Transaction{
		Version:  1,
		Inputs:   []*block.TransactionInput{{}},
		Outputs:  []*block.TransactionOutput{{Amount: reward, LockingScript: m.Id.GetPublicKeyString()}},
		LockTime: 0,
	}
	return new_tx
}

// getInputSums returns the sums of the inputs of a slice of transactions,
// as well as an error if the function fails. This function sends a request to
// its GetInputsSum channel, which the node picks up. The node then handles
// the request, returning the sum of the inputs in the InputsSum channel.
// This function times out after 1 second.
func (m *Miner) getInputSums(txs []*block.Transaction) ([]uint32, error) {
	// time out after 1 second
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// ask the node to sum the inputs for our transactions
	m.GetInputSums <- txs
	// wait until we get a response from the node in our SumInputs channel
	for {
		select {
		case <-ctx.Done():
			// Oops! We ran out of time
			return []uint32{0}, fmt.Errorf("[miner.sumInputs] Error: timed out")
		case sums := <-m.InputSums:
			// Yay! We got a response from our node.
			return sums, nil
		}
	}
}

// CalculateMintingReward calculates
// the minting reward the miner should receive based
// on the current chain length.
func (m *Miner) CalculateMintingReward() uint32 {
	c := m.Config
	chainLength := m.ChainLength.Load()
	if chainLength >= c.SubsidyHalvingRate*c.MaxHalvings {
		return 0
	}
	halvings := chainLength / c.SubsidyHalvingRate
	rwd := c.InitialSubsidy
	rwd /= uint32(math.Pow(2, float64(halvings)))
	return rwd
}
