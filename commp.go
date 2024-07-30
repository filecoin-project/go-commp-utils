package commp

import (
	"errors"
	"io"

	commcid "github.com/filecoin-project/go-fil-commcid"
	commphh "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// GeneratePieceCIDFromFile generates an a piece cid from an io.Reader
func GeneratePieceCIDFromFile(proofType abi.RegisteredSealProof, piece io.Reader, pieceSize abi.UnpaddedPieceSize) (cid.Cid, error) {
	var cc commphh.Calc
	if _, err := io.Copy(&cc, piece); err != nil {
		return cid.Undef, err
	}
	p, _, err := cc.Digest()
	if err != nil {
		return cid.Undef, err
	}
	return commcid.PieceCommitmentV1ToCID(p)
}

func ZeroPadPieceCommitment(c cid.Cid, curSize abi.UnpaddedPieceSize, toSize abi.UnpaddedPieceSize) (cid.Cid, error) {
	dmh, err := multihash.Decode(c.Hash())
	if err != nil {
		return cid.Undef, err
	}
	if len(dmh.Digest) != 32 {
		return cid.Undef, errors.New("invalid piece commitment, must be a 32 byte hash")
	}
	rawCommp, err := commphh.PadCommP(dmh.Digest, uint64(curSize.Padded()), uint64(toSize.Padded()))
	if err != nil {
		return cid.Undef, err
	}
	return commcid.DataCommitmentV1ToCID(rawCommp)
}
