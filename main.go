package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var nodeID string
var blockchain *Blockchain

func defaultHandler(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Takes team work to make the dream work.")
}

// HealthHandler is the HTTP handler that is expected to be used by a load
// balancer.  As such it simply returns HTTP-200.
func HealthHandler(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Healthy")
}

// FullChainHandler returns the blockchain as known by this node.
func FullChainHandler(ctx *gin.Context) {

	response := gin.H{
		"blockchain": blockchain.Chain,
		"network":    blockchain.Nodes,
	}
	ctx.JSON(http.StatusOK, response)
}

// MineHandler will run the Proof-of-Work routine, when the new proof
// is found a new block is created (aka it is considered "forged").
// A new transaction is created awarding coin to this node, and the new
// block (with the temporarily held transactions!) is added to the
// blockchain.
func MineHandler(ctx *gin.Context) {
	// TODO: farm out to another goroutine?

	// Run the proof of work algorithm to get the next proof.
	lastBlock := blockchain.LastBlock()
	lastProof := lastBlock.Proof
	proof := ProofOfWork(lastProof)

	// Receive a reward for finding the proof.
	// "0" signify that this node has mined a new coin.
	log.Printf("[node] %s\n", nodeID)
	blockchain.AddTransaction("0", nodeID, 1.0)

	// Forge the new Block by adding it to the chain
	block := blockchain.NewBlock(proof, Hash(lastBlock))

	response := gin.H{
		"message":      "New Block Forged",
		"index":        block.Index,
		"transactions": block.Transactions,
		"proof":        block.Proof,
		"parent_hash":  block.ParentHash,
	}
	ctx.JSON(http.StatusOK, response)
}

// TransactionHandler implicitly validates the POSTed data, creates a
// new transaction.  This transaction is not officially added to a
// block until the block has been mined and added to the blockchain.
func TransactionHandler(ctx *gin.Context) {

	transaction := Trans{}
	err := ctx.Bind(&transaction)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Malformed request")
	}
	idx := blockchain.AddTransaction(transaction.Sender,
		transaction.Recvr,
		transaction.Amount)

	response := gin.H{
		"message": fmt.Sprintf("Transaction will be added to Block %d", idx),
	}
	ctx.JSON(http.StatusCreated, response)
}
func NodeHandler(ctx *gin.Context) {

	node := Node{}
	err := ctx.Bind(&node)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Malformed request")
	}
	blockchain.RegisterNode(node)

	response := gin.H{
		"message": "Node registered",
		"nodes":   blockchain.Nodes,
	}
	ctx.JSON(http.StatusCreated, response)
}

func main() {

	port := ":8181"
	// Generate a globally unique address for this node, used when
	// creating new transactions
	_nodeID, err := NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	node := Node{ID: _nodeID, IPaddr: fmt.Sprintf("127.0.0.1:%s", port)}
	log.Printf("[genesis node] ID: %s\n", node.ID)

	// Instantiate the Blockchain, registering myself
	blockchain = NewBlockchain()
	blockchain.RegisterNode(node)

	router := gin.Default()
	router.Use(gin.Logger())
	router.GET("/chain", FullChainHandler)
	router.GET("/health", HealthHandler)
	router.GET("/mine", MineHandler)
	router.POST("/transactions/new", TransactionHandler)
	router.POST("/nodes/register", NodeHandler)

	fmt.Printf("Listening on %s...\n", port)
	err = http.ListenAndServe(port, router)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to listen on port(%s): %s", port, err.Error()))
	}
}
