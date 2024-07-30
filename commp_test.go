package commp_test

import (
	"bytes"
	"testing"

	"github.com/filecoin-project/go-commp-utils"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
)

func TestZeroPadding(t *testing.T) {
	psize := abi.PaddedPieceSize(128).Unpadded()
	data := make([]byte, psize)

	content := []byte("i am the biggest cat, what do you think about that")
	copy(data, content)

	pcid, err := commp.GeneratePieceCIDFromFile(abi.RegisteredSealProof_StackedDrg32GiBV1_1, bytes.NewReader(data), psize)
	if err != nil {
		t.Fatal(err)
	}

	// value calculated using separate tooling
	expCid, err := cid.Decode("baga6ea4seaqozp3abki6vgdf7ztbipcycmxfyt2o64cpuyvdkczsjxsg7bqmioi")
	if err != nil {
		t.Fatal(err)
	}

	if pcid != expCid {
		t.Fatalf("expected %s, got %s", expCid, pcid)
	}

	padded, err := commp.ZeroPadPieceCommitment(pcid, psize, psize*4)
	if err != nil {
		t.Fatal(err)
	}

	data2 := make([]byte, psize*4)
	copy(data2, content)

	padExpCid, err := commp.GeneratePieceCIDFromFile(abi.RegisteredSealProof_StackedDrg32GiBV1_1, bytes.NewReader(data2), psize*4)
	if err != nil {
		t.Fatal(err)
	}

	if padded != padExpCid {
		t.Fatalf("wrong padding, expected %s, got %s", padExpCid, padded)
	}
}
