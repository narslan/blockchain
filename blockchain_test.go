package ethos_test

import (
	"testing"

	"github.com/narslan/ethos"
)

func TestNewBlockChain(t *testing.T) {

	bc := ethos.NewBlockChain()

	if len(bc.Chain) != 1 {
		t.Fatalf("length of blockchain should be 1 not %d", len(bc.Chain))
	}

	t.Run("genesis block", func(t *testing.T) {
		if bc.Chain[0].Idx != 1 {
			t.Fatalf("the first id of genesis block should be 1 not %d", bc.Chain[0].Idx)
		}

		if bc.Chain[0].PreviousHash != "1" {
			t.Fatalf("the previous hash of genesis block should be %v not %v", "1", bc.Chain[0].PreviousHash)
		}
	})

	t.Run("proof assignment", func(t *testing.T) {
		bc.NewBlock(1000)

		if bc.Chain[1].Proof != 1000 {
			t.Fatalf("proof of block should be %d not %d", 1000, len(bc.Chain))
		}
	})

}

func TestNewTx(t *testing.T) {

	bc := ethos.NewBlockChain()

	if len(bc.Chain) != 1 {
		t.Fatalf("length of blockchain should be 1 not %d", len(bc.Chain))
	}

	idx := bc.NewTx("you", "me", 100)

	if idx != 1 {
		t.Fatalf("index of the block should be %d found %d", 1, len(bc.Chain))
	}
}
