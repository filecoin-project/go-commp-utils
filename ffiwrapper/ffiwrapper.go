package ffiwrapper

import (
	"io"
	"os"
	"sync"

	ffi "github.com/filecoin-project/filecoin-ffi"
	"github.com/filecoin-project/go-commp-utils/zerocomm"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"
)

var log = logging.Logger("ffi-wrapper")

// ToReadableFile generates an os readable file from an io.Reader (converts via pipe & copy)
func ToReadableFile(r io.Reader, n int64) (*os.File, func() error, error) {
	f, ok := r.(*os.File)
	if ok {
		return f, func() error { return nil }, nil
	}

	var w *os.File

	f, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	var wait sync.Mutex
	var werr error

	wait.Lock()
	go func() {
		defer wait.Unlock()

		var copied int64
		copied, werr = io.CopyN(w, r, n)
		if werr != nil {
			log.Warnf("toReadableFile: copy error: %+v", werr)
		}

		err := w.Close()
		if werr == nil && err != nil {
			werr = err
			log.Warnf("toReadableFile: close error: %+v", err)
			return
		}
		if copied != n {
			log.Warnf("copied different amount than expected: %d != %d", copied, n)
			werr = xerrors.Errorf("copied different amount than expected: %d != %d", copied, n)
		}
	}()

	return f, func() error {
		wait.Lock()
		return werr
	}, nil
}

// GeneratePieceCIDFromFile generates an a piece cid from an io.Reader
func GeneratePieceCIDFromFile(proofType abi.RegisteredSealProof, piece io.Reader, pieceSize abi.UnpaddedPieceSize) (cid.Cid, error) {
	f, werr, err := ToReadableFile(piece, int64(pieceSize))
	if err != nil {
		return cid.Undef, err
	}

	pieceCID, err := ffi.GeneratePieceCIDFromFile(proofType, f, pieceSize)
	if err != nil {
		return cid.Undef, err
	}

	return pieceCID, werr()
}

func ZeroPadPieceCommitment(c cid.Cid, curSize abi.UnpaddedPieceSize, toSize abi.UnpaddedPieceSize) (cid.Cid, error) {
	cur := c
	for curSize < toSize {

		zc := zerocomm.ZeroPieceCommitment(curSize)

		p, err := ffi.GenerateUnsealedCID(abi.RegisteredSealProof_StackedDrg32GiBV1, []abi.PieceInfo{
			abi.PieceInfo{
				Size:     curSize.Padded(),
				PieceCID: cur,
			},
			abi.PieceInfo{
				Size:     curSize.Padded(),
				PieceCID: zc,
			},
		})
		if err != nil {
			return cid.Undef, err
		}

		cur = p
		curSize = curSize * 2
	}

	return cur, nil
}
