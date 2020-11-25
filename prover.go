package rsmt2d

import (
	"crypto/sha256"
	"hash"
	"io"

	"github.com/NebulousLabs/merkletree"
)

type ProofType int

// what proof types does rsmt2d *need*?
const (
	Inclusion ProofType = iota
	Exclusion
	Range
	NameSpace
)

// Proof describes the data needed to verify a rsmt2d compatable merkle tree
// notes: I'm not really sure that returning single require struct is more
// convienient than returning four seperate objects
type Proof struct {
	Type   ProofType // is this even needed?
	Root   []byte
	Set    [][]byte
	Index  uint64
	Leaves uint64
}

// Prover wraps the required methods to create merkle proofs for portions of an
// ExtendedDataSqaure
type Prover interface {
	Prove(ind uint, leaves [][]byte) (Proof, error)
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
	tree.SetIndex(uint64(ind))
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

//

//

//

//////////////////
// Old... these aren't different from the original impl
/////////////////
// Tree describes the methods expected for a rsmt2d compatible merkle tree
type Tree interface {
	// Using this method instead of Push might not be the best idea, because
	// then the implementation isn't going to panic on an error, but I figured
	// I'd include it
	io.Writer
	// Push([]byte) error // an error might be called for
	// Push([]byte)
	Prover
	Root() []byte
}

// TreeCreator (alternatively TreeFactory per @musalbas) issues a new merkle
// tree that can be used by rsmt2d to hash a row or column of erasured data
type TreeCreator interface {
	New() Tree
}

// Prover (rename to Prover) wraps the Prove method that issues a merkle
// proof
type TreeProver interface {
	Prove(x, y uint) (Proof, error)
}

// LooseProver would force the tree implementation to parse incoming pointers to
// proofs. I'm not the biggest fan of this one, but I thought I'd include it.
type LooseProver interface {
	// take in a pointer to an arbitrary proof datastructure
	// fill with proof data
	Prove(int, interface{}) error
	// Verify some proof
	Verify(interface{}) (bool, error)
}

type ByteProver interface {
	Prove([]byte) (Proof, error)
	Verify([]byte) (bool, error)
}
