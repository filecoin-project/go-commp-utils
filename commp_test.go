package commp_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/filecoin-project/go-commp-utils/v2"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

func TestZeroPadding(t *testing.T) {
	psize := abi.PaddedPieceSize(128).Unpadded()
	data := make([]byte, psize)

	content := []byte("i am the biggest cat, what do you think about that")
	copy(data, content)

	pcid, err := commp.GeneratePieceCIDFromFile(abi.RegisteredSealProof_StackedDrg32GiBV1_1, bytes.NewReader(data), psize)
	require.NoError(t, err)

	// value calculated using separate tooling
	expCid, err := cid.Decode("baga6ea4seaqozp3abki6vgdf7ztbipcycmxfyt2o64cpuyvdkczsjxsg7bqmioi")
	require.NoError(t, err)
	require.Equal(t, expCid, pcid)

	padded, err := commp.ZeroPadPieceCommitment(pcid, psize, psize*4)
	require.NoError(t, err)

	data2 := make([]byte, psize*4)
	copy(data2, content)

	padExpCid, err := commp.GeneratePieceCIDFromFile(abi.RegisteredSealProof_StackedDrg32GiBV1_1, bytes.NewReader(data2), psize*4)
	require.NoError(t, err)
	require.Equal(t, padded, padExpCid)
}

func TestMismatchedPieceSize(t *testing.T) {
	psize := abi.PaddedPieceSize(128).Unpadded()
	data := []byte("i am the smallest cat") // less than psize
	_, err := commp.GeneratePieceCIDFromFile(abi.RegisteredSealProof_StackedDrg32GiBV1_1, bytes.NewReader(data), psize)
	require.ErrorIs(t, err, io.EOF)
}
