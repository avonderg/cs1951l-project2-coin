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
	var coins []*CoinInfo
	// change : (how much you spent) - (amount + fee)
	sum := uint32(0)

	for out, coin := range w.CoinCollection {
		sum += coin.TransactionOutput.Amount
		signature, _ := coin.TransactionOutput.MakeSignature(w.Id)
		inp := &block.TransactionInput{coin.ReferenceTransactionHash, coin.OutputIndex, signature}
		inputs = append(inputs, inp)
		// enough to satisfy amount + fee? it exceeds by a little and what you exceed it by is what you are returning
		coins = append(coins, coin)
		delete(w.CoinCollection, out) // do you remove it from the coin collection??????
	}
	change := sum - (amount + fee)

	return change, inputs, coins
}

// generateTransactionOutputs generates the transaction outputs required to create a transaction.
func (w *Wallet) generateTransactionOutputs(
	amount uint32,
	receiverPK []byte,
	change uint32,
) []*block.TransactionOutput {
	//TODO: optional, but we recommend using a helper like this
	var outputs []*block.TransactionOutput
	output_receiver := &block.TransactionOutput{amount, string(receiverPK)}
	outputs = append(outputs, output_receiver)
	output_change := &block.TransactionOutput{change, string(w.Id.GetPublicKeyBytes())}
	outputs = append(outputs, output_change)
	return outputs
}

// RequestTransaction allows the wallet to send a transaction to the node,
// which will propagate the transaction along the P2P network.
func (w *Wallet) RequestTransaction(amount uint32, fee uint32, recipientPK []byte) *block.Transaction {
	//TODO
	if !(w.Balance >= amount+fee) {
		return nil
	}
	//pk := w.Id.GetPublicKeyString()
	change, inputs, coins := w.generateTransactionInputs(amount, fee)
	outputs := w.generateTransactionOutputs(amount, recipientPK, change)

	w.UnseenSpentCoins[coins[0].ReferenceTransactionHash] = coins
	transac := &block.Transaction{Version: 0, Inputs: inputs, Outputs: outputs, LockTime: 0} // unsure ab locktime and version

	w.Balance = change + fee
	return transac
}

// HandleBlock handles the transactions of a new block. It:
// (1) sees if any of the inputs are ones that we've spent
// (2) sees if any of the incoming outputs on the block are ours
// (3) updates our unconfirmed coins, since we've just gotten
// another confirmation!
func (w *Wallet) HandleBlock(txs []*block.Transaction) {
	//TODO
	for _, tx := range txs {
		w.checkInputs(tx)
		w.checkOutputs(tx.Outputs, tx)
		w.updateCoin()
		// third helper that increments by 1 and chceks if it exceeds the limit and delete
	}
}

// look at other fuctions dealing with putting txs into a block (# txs in a block incorrect)!! otherwise its good

// step (1): sees if any of the inputs are ones that we've spent
func (w *Wallet) checkInputs(tx *block.Transaction) {
	//TODO
	//for _, input := range inps {
	//	hash := input.ReferenceTransactionHash
	//	if _, ok := w.UnseenSpentCoins[hash]; ok { // if spent
	//		coinInfo := w.UnseenSpentCoins[hash]
	//
	//		delete(w.UnseenSpentCoins, hash)
	//		for _, coin := range coinInfo {
	//			w.UnconfirmedSpentCoins[coin] = 1
	//		}
	//	}
	//}
	hash := tx.Hash()
	if _, ok := w.UnseenSpentCoins[hash]; ok { // if spent
		coinInfo := w.UnseenSpentCoins[hash]

		delete(w.UnseenSpentCoins, hash)

		for _, coin := range coinInfo {
			w.UnconfirmedSpentCoins[coin] = 0
		}
	}
}

// step (2): sees if any of the incoming outputs on the block are ours
func (w *Wallet) checkOutputs(outs []*block.TransactionOutput, tx *block.Transaction) {
	for i, out := range outs {
		// create coin
		if out.LockingScript == w.Id.GetPublicKeyString() {
			coin := &CoinInfo{tx.Hash(), uint32(i), out}

			// check if coin is ours
			w.UnconfirmedReceivedCoins[coin] = 1
		}
	}
}

// step (3):  updates our unconfirmed coins, since we've just gotten another confirmation!
func (w *Wallet) updateCoin() {
	// unconfirmed receveved < safe block amount (wallet.confirmed..)
	// loop through all coinfos and corresponding # confirmations , if confirmations >= safe block amt
	// then delete the unconfirmed coins from that field
	// otherwise if they haven't reached it then increment the # of confirmations

	for coin, _ := range w.UnconfirmedSpentCoins {
		w.UnconfirmedSpentCoins[coin] += 1
		if (w.UnconfirmedSpentCoins[coin]) >= w.Config.SafeBlockAmount { // safe block amount????
			delete(w.UnconfirmedSpentCoins, coin)
			delete(w.CoinCollection, coin.TransactionOutput)
		}
	}
	for coin, _ := range w.UnconfirmedReceivedCoins {
		w.UnconfirmedReceivedCoins[coin] += 1
		if (w.UnconfirmedReceivedCoins[coin]) >= w.Config.SafeBlockAmount {
			w.Balance += coin.TransactionOutput.Amount // add it to balance
			w.CoinCollection[coin.TransactionOutput] = coin
			delete(w.UnconfirmedReceivedCoins, coin)
		}
	}
}

// HandleFork handles a fork, updating the wallet's relevant fields.
func (w *Wallet) HandleFork(blocks []*block.Block, undoBlocks []*chainwriter.UndoBlock) {
	//TODO: for extra credit!
}
