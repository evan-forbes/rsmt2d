package rsmt2d

import (
	"crypto/sha256"
	"hash"

	"github.com/NebulousLabs/merkletree"
)

// Proof describes the data needed to verify a rsmt2d compatable merkle tree
// notes: I'm not really sure that returning single require struct is more
// convienient than returning four separate objects
type Proof struct {
	Root   []byte
	Set    [][]byte
	Index  uint64
	Leaves uint64
}

// Prover wraps the required methods to create merkle proofs for portions of an
// ExtendedDataSqaure
type Prover interface {
	Prove(ind uint, leaves [][]byte) (Proof, error)
	Root(ind uint, leaves [][]byte) []byte
}

type defaultProver struct {
	hasher func() hash.Hash
}

func NewDefaultProver() Prover {
	return defaultProver{hasher: sha256.New}
}

// Prove fullfills the Prover interface to make the merkle proofs expected by rsmt2d
func (p defaultProver) Prove(ind uint, leaves [][]byte) (proof Proof, err error) {
	tree := merkletree.New(p.hasher())
	err = tree.SetIndex(uint64(ind))
	if err != nil {
		return proof, err
	}
	for _, leaf := range leaves {
		tree.Push(leaf)
	}
	merkleRoot, proofSet, proofIndex, numLeaves := tree.Prove()
	return Proof{
		Type:   Inclusion,
		Root:   merkleRoot,
		Set:    proofSet,
		Index:  proofIndex,
		Leaves: numLeaves,
	}, nil
}

// Root fullfills the Prover interface to make merkle roots expected by a datasquare
func (p defaultProver) Root(ind uint, leaves [][]byte) []byte {
	tree := merkletree.New(p.hasher())
	err := tree.SetIndex(uint64(ind))
	// we should never have to worry about this considering we're making an empty tree
	if err != nil {
		panic(err)
	}
	for _, leaf := range leaves {
		tree.Push(leaf)
	}
	return tree.Root()
}
