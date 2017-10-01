package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

type Node struct {
	ID     string `json:"node_id"`
	IPaddr string `json:"ip" binding:"required"`
}

// Trans hold a single transaction
type Trans struct {
	Sender string  `json:"sender" binding:"required"`
	Recvr  string  `json:"receiver" binding:"required"`
	Amount float64 `json:"amount" binding:"required"`
}

// Block holds the data for a block
type Block struct {
	Index        int       `json:"index"`
	Timestamp    time.Time `json:"timestamp"`
	Transactions []Trans   `json:"transactions"`
	Proof        int       `json:"proof"`
	ParentHash   string    `json:"parent_hash"`
}

type Blockchain struct {
	tempTrans []Trans  // explicitly not pointers
	Chain     []*Block `json:"chain"`
	Nodes     []Node   `json:"nodes"`
}

func NewBlockchain() *Blockchain {

	trans := []Trans{}
	chain := []*Block{}
	nodes := []Node{}
	bc := &Blockchain{tempTrans: trans, Chain: chain, Nodes: nodes}

	// Create the genesis block
	bc.NewBlock(100, "1")
	return bc
}

// NewBlock creates a new Block in the Blockchain.
func (bc *Blockchain) NewBlock(proof int, hash string) *Block {

	//previous_hash: Hash of previous Block
	//proof: The proof given by the Proof of Work algorithm

	block := &Block{
		Index:        len(bc.Chain) + 1,
		Timestamp:    time.Now(),
		Transactions: copyTrans(bc.tempTrans), // TODO: deep copy?
		Proof:        proof,
		ParentHash:   hash,
	}
	// clear
	bc.tempTrans = nil
	bc.Chain = append(bc.Chain, block)
	return block
}

// LastBlock is a convenience function
func (bc *Blockchain) LastBlock() *Block {
	return bc.Chain[len(bc.Chain)-1]
}

// AddTransaction creates a new Transaction to be added into the next
// mined Block.  The index of the block that will hold this transaction
// will be returned.
func (bc *Blockchain) AddTransaction(sender, receiver string, amount float64) int {
	trans := Trans{
		Sender: sender,
		Recvr:  receiver,
		Amount: amount,
	}
	bc.tempTrans = append(bc.tempTrans, trans)

	return (bc.LastBlock().Index) + 1
}

// RegisterNode registers a new node by IP address with the network.
func (bc *Blockchain) RegisterNode(node Node) {
	id, _ := NewUUID()
	node.ID = id
	bc.Nodes = append(bc.Nodes, node)
}

// ProofOfWork implements a very simple Proof of Work algorithm:
//	Find a number p' such that hash(pp') contains leading 4 zeroes,
//	where p is the previous proof and p' is the new proof.
func ProofOfWork(lastProof int) int {
	proof := 0
	for !ValidProof(lastProof, proof) {
		proof++
	}
	return proof
}

// ValidProof validates a proof as part of a Proof of Work algorithm
// implementation.
func ValidProof(lastProof, proof int) bool {

	guess := fmt.Sprintf("%s%s", strconv.Itoa(lastProof), strconv.Itoa(proof))

	// double sha-256 just for fun...
	h1 := sha256.New()
	h2 := sha256.New()
	h1.Write([]byte(guess))
	h2.Write(h1.Sum(nil))
	r := hex.EncodeToString(h2.Sum(nil))
	return r[len(r)-4:] == "0000"
}

// Hash creates a SHA-256 hash of the given block.
func Hash(block *Block) string {
	// The json package always orders keys when marshalling which is
	// critical else we'll have inconsistent hashes.
	result, _ := json.Marshal(block)
	h := sha256.New()
	h.Write(result)
	r := hex.EncodeToString(h.Sum(nil))
	return r
}

// NewUUID generates a random UUID according to RFC 4122
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
func copyTrans(src []Trans) []Trans {
	// The built-in `copy` func alone will not do the job here as:
	// 	"The number of elements copied is the minimum of len(src) and len(dst)."
	tmp := make([]Trans, len(src))
	copy(tmp, src)
	return tmp
}
