package ethos

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

//BlockChain is responsible for managing the chain.
type BlockChain struct {
	Chain     []*Block
	CurrentTx []Transaction
	mu        sync.Mutex
}

//Block is the most fundemental structure of blockchain.
type Block struct {
	Idx          int
	TimeStamp    int64
	Tx           []Transaction
	Proof        int64  //The proof given by the Proof of Work algorithm
	PreviousHash string //Hash of previous Block
}

//Transaction holds the information of a transaction.
type Transaction struct {
	Sender    string
	Recipient string
	Amount    float64
}

func NewBlockChain() *BlockChain {
	//create a new blockchain
	bc := &BlockChain{
		Chain: make([]*Block, 0),
	}

	//Genesis block has a default proof of 100
	bc.NewBlock(100)
	return bc

}

//NewBlock accepts a previous hash and a proof which in return produces a block.
func (bc *BlockChain) NewBlock(proof int64) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	//If blockchain has no block create hash "1" as a special case for genesis
	//block.
	var phash string
	if len(bc.Chain) == 0 {
		phash = "1"
	} else {
		phash = asSha256(bc.Chain[len(bc.Chain)-1])
	}
	//Create a new block.
	b := &Block{
		Idx:          len(bc.Chain) + 1,
		TimeStamp:    time.Now().Unix(),
		Tx:           make([]Transaction, 0),
		PreviousHash: phash,
		Proof:        proof,
	}
	copy(b.Tx, bc.CurrentTx)
	//Reset current transaction.
	bc.CurrentTx = make([]Transaction, 0)
	//Add block to the end of the chain.
	bc.Chain = append(bc.Chain, b)
	return b
}

//Creates a new transaction to go into the next mined Block.
func (bc *BlockChain) NewTx(sender, recipient string, amount float64) int {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	tx := Transaction{
		sender,
		recipient,
		amount,
	}
	bc.CurrentTx = append(bc.CurrentTx, tx)
	//The index of the Block that will hold this transaction
	return len(bc.Chain)

}

// Simple Proof of Work Algorithm:
//          - Find a number p' such that hash(pp') contains leading 4 zeroes, where p is the previous p'
//          - p is the previous proof, and p' is the new proof
func (bc *BlockChain) ProofOfWork() int64 {

	proof := int64(0)
	lastProof := bc.Chain[len(bc.Chain)-1].Proof
	lastHash := asSha256(bc.Chain[len(bc.Chain)-1])

	for {

		if validateProof(lastHash, proof, lastProof) {
			break
		} else {
			proof = proof + 1
		}

	}

	return proof

}

//Validates the Proof: Does hash(last_proof, proof) contain 4 leading zeroes?
func validateProof(lastHash string, lastProof, proof int64) bool {

	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d%d%s", lastProof, proof, lastHash)))
	guess := fmt.Sprintf("%x", h.Sum(nil))
	return guess[:4] == "0000"

}

func asSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}
