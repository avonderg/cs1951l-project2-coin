package wallet

import (
	"Coin/pkg/block"
	"Coin/pkg/blockchain/chainwriter"
	"Coin/pkg/id"
)

// CoinInfo holds the information about a TransactionOutput
// necessary for making a TransactionInput.
// ReferenceTransactionHash is the hash of the transaction that the
// output is from.
// OutputIndex is the index into the Outputs array of the
// Transaction that the TransactionOutput is from.
// TransactionOutput is the actual TransactionOutput
type CoinInfo struct {
	ReferenceTransactionHash string
	OutputIndex              uint32
	TransactionOutput        *block.TransactionOutput
}

// Wallet handles keeping track of the owner's coins
//
// # CoinCollection is the owner of this wallet's set of coins
//
// UnseenSpentCoins is a mapping of transaction hashes (which are strings)
// to a slice of coinInfos. It's used for keeping track of coins that we've
// used in a transaction but haven't yet seen in a block.
//
// UnconfirmedSpentCoins is a mapping of Coins to number of confirmations
// (which are integers). We can't confirm that a Coin has been spent until
// we've seen enough POW on top the block containing our sent transaction.
//
// UnconfirmedReceivedCoins is a mapping of CoinInfos to number of confirmations
// (which are integers). We can't confirm we've received a Coin until
// we've seen enough POW on top the block containing our received transaction.
type Wallet struct {
	Config              *Config
	Id                  id.ID
	TransactionRequests chan *block.Transaction
	Address             string
	Balance             uint32

	// All coins
	CoinCollection map[*block.TransactionOutput]*CoinInfo

	// Not yet seen
	UnseenSpentCoins map[string][]*CoinInfo

	// Seen but not confirmed
	UnconfirmedSpentCoins    map[*CoinInfo]uint32
	UnconfirmedReceivedCoins map[*CoinInfo]uint32
}

// SetAddress sets the address
// of the node in the wallet.
func (w *Wallet) SetAddress(a string) {
	w.Address = a
}

// New creates a wallet object
func New(config *Config, id id.ID) *Wallet {
	if !config.HasWallet {
		return nil
	}
	return &Wallet{
		Config:                   config,
		Id:                       id,
		TransactionRequests:      make(chan *block.Transaction),
		Balance:                  0,
		CoinCollection:           make(map[*block.TransactionOutput]*CoinInfo),
		UnseenSpentCoins:         make(map[string][]*CoinInfo),
		UnconfirmedSpentCoins:    make(map[*CoinInfo]uint32),
		UnconfirmedReceivedCoins: make(map[*CoinInfo]uint32),
	}
}

// generateTransactionInputs creates the transaction inputs required to make a transaction.
// In addition to the inputs, it returns the amount of change the wallet holder should
// return to themselves, and the coinInfos used
func (w *Wallet) generateTransactionInputs(amount uint32, fee uint32) (uint32, []*block.TransactionInput, []*CoinInfo) {
	//TODO: optional, but we recommend using a helper like this
	var inputs []*block.TransactionInput

	return 0, nil, nil
}

// generateTransactionOutputs generates the transaction outputs required to create a transaction.
func (w *Wallet) generateTransactionOutputs(
	amount uint32,
	receiverPK []byte,
	change uint32,
) []*block.TransactionOutput {
	//TODO: optional, but we recommend using a helper like this
	var outputs []*block.TransactionOutput
	output := &block.TransactionOutput{amount, string(receiverPK)}
	outputs = append(outputs, output)
	w.Balance += change
	return outputs
}

// RequestTransaction allows the wallet to send a transaction to the node,
// which will propagate the transaction along the P2P network.
func (w *Wallet) RequestTransaction(amount uint32, fee uint32, recipientPK []byte) *block.Transaction {
	//TODO
	return nil
}

// HandleBlock handles the transactions of a new block. It:
// (1) sees if any of the inputs are ones that we've spent
// (2) sees if any of the incoming outputs on the block are ours
// (3) updates our unconfirmed coins, since we've just gotten
// another confirmation!
func (w *Wallet) HandleBlock(txs []*block.Transaction) {
	//TODO
	for _, tx := range txs {
		w.checkInputs(tx.Inputs)
		w.checkOutputs(tx.Outputs, tx.Inputs)
	}
}

// step (1): sees if any of the inputs are ones that we've spent
func (w *Wallet) checkInputs(inps []*block.TransactionInput) {
	//TODO
	for _, input := range inps {
		hash := input.ReferenceTransactionHash
		if _, ok := w.UnseenSpentCoins[hash]; ok { // if spent
			coinInfo := w.UnseenSpentCoins[hash]
			count := 0
			for _, coin := range coinInfo {
				w.UnconfirmedSpentCoins[coin] = uint32(count + 1) // is count correct
				delete(w.UnseenSpentCoins, hash)                  // delete from map
			}
		}
	}
}

// step (2): sees if any of the incoming outputs on the block are ours
func (w *Wallet) checkOutputs(outs []*block.TransactionOutput, inps []*block.TransactionInput) {
	for i, out := range outs {
		// create coin
		if out.LockingScript == w.Id.GetPublicKeyString() {
			coin := &CoinInfo{inps[i].ReferenceTransactionHash, uint32(i), out}

			// check if coin is ours
			w.UnconfirmedReceivedCoins[coin] = 0 // no clue what to assign it to
		}
	}
}

// step (3):  updates our unconfirmed coins, since we've just gotten another confirmation!
func (w *Wallet) updateCoin() {
	// unconfirmed receveved < safe block amount (wallet.confirmed..)
	// loop through all coinfos and corresponding # confirmations , if confirmations >= safe block amt
	// then delete the unconfirmed coins from that field
	// otherwise if they haven't reached it then increment the # of confirmations

	for _, coin := range w.CoinCollection {
		if _, ok := w.UnconfirmedSpentCoins[coin]; ok { // if its unconfirmed spent
			if (w.UnconfirmedSpentCoins[coin]) >= w.Config.SafeBlockAmount { // safe block amount????
				delete(w.UnconfirmedSpentCoins, coin)
			} else {
				w.UnconfirmedSpentCoins[coin] += 1
			}
		}
		if _, ok := w.UnconfirmedReceivedCoins[coin]; ok { // if its unconfirmed received
			if (w.UnconfirmedReceivedCoins[coin]) >= w.Config.SafeBlockAmount {
				w.Balance += w.UnconfirmedReceivedCoins[coin] // add it to balance
				delete(w.UnconfirmedReceivedCoins, coin)
			} else {
				w.UnconfirmedReceivedCoins[coin] += 1
			}
		}
	}
}

// HandleFork handles a fork, updating the wallet's relevant fields.
func (w *Wallet) HandleFork(blocks []*block.Block, undoBlocks []*chainwriter.UndoBlock) {
	//TODO: for extra credit!
}
