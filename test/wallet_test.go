package test

import (
	"Coin/pkg/blockchain"
	"Coin/pkg/utils"
	"testing"
)

func TestMine1(t *testing.T) {
	utils.SetDebug(true)

	nodes := NewCluster(5)
	var chains []*blockchain.BlockChain
	for _, node := range nodes {
		chains = append(chains, node.BlockChain)
	}
	defer CleanUp(chains)
	//genBlock := blockchain.GenesisBlock(blockchain.DefaultConfig())
	//chain := blockchain.New(blockchain.DefaultConfig())
	StartCluster(nodes)
	user := nodes[1]
	FillWalletWithCoins(user.Wallet, 10, 250)
	if user.Wallet.Balance != 250*10 {
		t.Errorf("user has wrong number of coins in its wallet")
	}
	//StartMiners(nodes)
	//defer CleanUp()
}
