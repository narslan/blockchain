package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/narslan/ethos"
)

type TransactionRequest struct {
	Sender    string  `json:"sender,omitempty"`
	Recipient string  `json:"recipient,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
}

//Chain dumps the state of blockchain.
func (s *Server) Chain(c *gin.Context) {
	c.JSON(http.StatusOK, s.BlockChain)
}

//NewTx creates a new transaction .
func (s *Server) NewTx(c *gin.Context) {

	var txReq TransactionRequest
	if err := c.ShouldBindJSON(&txReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idx := s.BlockChain.NewTx(txReq.Sender, txReq.Recipient, txReq.Amount)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Transaction will be added block: [%d]", idx)})
}

type MineResponse struct {
	Message      string              `json:"message"`
	Index        int                 `json:"index"`
	Transactions []ethos.Transaction `json:"transactions"`
	Proof        int64               `json:"proof"`
	PreviousHash string              `json:"previous_hash"`
}

func (s *Server) Mine(c *gin.Context) {

	//run the proof of work algorithm to get the next proof
	//lastBlock := s.BlockChain.Chain[len(s.BlockChain.Chain)-1]
	proof := s.BlockChain.ProofOfWork()
	// We must receive a reward for finding the proof.
	// The sender is "0" to signify that this node has mined a new coin.
	s.BlockChain.NewTx(
		"0",
		"me", //TODO: this should be a node identifier
		1,
	)
	// We forge a new block by adding a newblock.
	b := s.BlockChain.NewBlock(proof)

	response := MineResponse{
		Message:      "New Block Forged",
		Index:        b.Idx,
		Transactions: b.Tx,
		Proof:        proof,
		PreviousHash: b.PreviousHash,
	}
	c.JSON(http.StatusOK, response)
}

// response = {
// 	'message': "New Block Forged",
// 	'index': block['index'],
// 	'transactions': block['transactions'],
// 	'proof': block['proof'],
// 	'previous_hash': block['previous_hash'],
// }
