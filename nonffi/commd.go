package nonffi

import (
	"errors"
	"fmt"
	"math/bits"

	"github.com/filecoin-project/go-commp-utils/zerocomm"
	commcid "github.com/filecoin-project/go-fil-commcid"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	sha256simd "github.com/minio/sha256-simd"
)

func GenerateUnsealedCID(proofType abi.RegisteredSealProof, pieces []abi.PieceInfo) (cid.Cid, error) {
	spi, found := abi.SealProofInfos[proofType]
	if !found {
		return cid.Undef, fmt.Errorf("unknown seal proof type %d", proofType)
	}
	if len(pieces) == 0 {
		return cid.Undef, errors.New("no pieces provided")
	}

	maxSize := abi.PaddedPieceSize(spi.SectorSize)

	// sancheck everything
	for i, p := range pieces {
		if p.Size < 128 {
			return cid.Undef, fmt.Errorf("invalid Size of PieceInfo %d: value %d is too small", i, p.Size)
		}
		if pieces[i].Size > maxSize {
			return cid.Undef, fmt.Errorf("invalid Size of PieceInfo %d: value %d is larger than sector size of SealProofType %d", i, p.Size, proofType)
		}
		if bits.OnesCount64(uint64(p.Size)) != 1 {
			return cid.Undef, fmt.Errorf("invalid Size of PieceInfo %d: value %d is not a power of 2", i, p.Size)
		}
		if _, err := commcid.CIDToPieceCommitmentV1(p.PieceCID); err != nil {
			return cid.Undef, fmt.Errorf("invalid PieceCid for PieceInfo %d: %w", i, err)
		}
	}

	// reimplement https://github.com/filecoin-project/rust-fil-proofs/blob/380d6437c2/filecoin-proofs/src/pieces.rs#L85-L145
	stack := append(
		make([]abi.PieceInfo, 0, 32),
		pieces[0],
	)

	for i := 1; i < len(pieces); i++ {

		for stack[len(stack)-1].Size < pieces[i].Size {
			lastSize := stack[len(stack)-1].Size

			stack = reduceStack(
				append(
					stack,
					abi.PieceInfo{
						Size:     lastSize,
						PieceCID: zerocomm.ZeroPieceCommitment(lastSize.Unpadded()),
					},
				),
			)
		}

		stack = reduceStack(
			append(
				stack,
				pieces[i],
			),
		)
	}

	for len(stack) > 1 {
		lastSize := stack[len(stack)-1].Size
		stack = reduceStack(
			append(
				stack,
				abi.PieceInfo{
					Size:     lastSize,
					PieceCID: zerocomm.ZeroPieceCommitment(lastSize.Unpadded()),
				},
			),
		)
	}

	if stack[0].Size > maxSize {
		return cid.Undef, fmt.Errorf("provided pieces sum up to %d bytes, which is larger than sector size of SealProofType %d", stack[0].Size, proofType)
	}

	return stack[0].PieceCID, nil
}

var s256 = sha256simd.New()

func reduceStack(s []abi.PieceInfo) []abi.PieceInfo {
	for {
		if len(s) < 2 || s[len(s)-2].Size != s[len(s)-1].Size {
			break
		}

		l, _ := commcid.CIDToPieceCommitmentV1(s[len(s)-2].PieceCID)
		r, _ := commcid.CIDToPieceCommitmentV1(s[len(s)-1].PieceCID)
		s256.Reset()
		s256.Write(l)
		s256.Write(r)
		d := s256.Sum(make([]byte, 0, 32))
		d[31] &= 0b00111111
		newPiece, _ := commcid.PieceCommitmentV1ToCID(d)

		s[len(s)-2] = abi.PieceInfo{
			Size:     2 * s[len(s)-2].Size,
			PieceCID: newPiece,
		}

		s = s[:len(s)-1]
	}

	return s
}
