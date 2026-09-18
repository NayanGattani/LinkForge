package shortid

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

// Init initializes the Snowflake node.
// In a distributed system, nodeID should be unique per machine/process (0-1023).
func Init(nodeID int64) error {
	var err error
	node, err = snowflake.NewNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to initialize snowflake node: %w", err)
	}
	return nil
}

// Generate creates a new unique 64-bit ID.
// Init must be called before calling Generate.
func Generate() uint64 {
	if node == nil {
		panic("shortid: Generate called before Init")
	}
	// snowflake.ID is essentially an int64, we can cast it safely as it won't be negative
	return uint64(node.Generate().Int64())
}

// GenerateBase62 generates a new ID and immediately encodes it to Base62.
func GenerateBase62() string {
	id := Generate()
	return Encode(id)
}
