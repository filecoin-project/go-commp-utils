package nonffi

import (
	"math/bits"
	"runtime"
	"sync"

	"github.com/filecoin-project/go-commp-utils/zerocomm"
	"github.com/filecoin-project/go-fil-commcid"
	"github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"
)

type DataCIDSize struct {
	PayloadSize int64
	PieceSize   abi.PaddedPieceSize
	PieceCID    cid.Cid
}

const commPBufPad = abi.PaddedPieceSize(8 << 20)
const CommPBuf = abi.UnpaddedPieceSize(commPBufPad - (commPBufPad / 128)) // can't use .Unpadded() for const

type ciderr struct {
	c   cid.Cid
	err error
}

type Writer struct {
	len    int64
	buf    [CommPBuf]byte
	leaves []chan ciderr

	tbufs    [][CommPBuf]byte
	throttle chan int
}

func (w *Writer) Write(p []byte) (int, error) {
	if w.throttle == nil {
		w.throttle = make(chan int, runtime.NumCPU())
		for i := 0; i < cap(w.throttle); i++ {
			w.throttle <- i
		}
	}
	if w.tbufs == nil {
		w.tbufs = make([][CommPBuf]byte, cap(w.throttle))
	}

	n := len(p)
	for len(p) > 0 {
		buffered := int(w.len % int64(len(w.buf)))
		toBuffer := len(w.buf) - buffered
		if toBuffer > len(p) {
			toBuffer = len(p)
		}

		copied := copy(w.buf[buffered:], p[:toBuffer])
		p = p[copied:]
		w.len += int64(copied)

		if copied > 0 && w.len%int64(len(w.buf)) == 0 {
			leaf := make(chan ciderr, 1)
			bufIdx := <-w.throttle
			copy(w.tbufs[bufIdx][:], w.buf[:])

			go func() {
				cc := borrowCommpCalc()
				defer func() {
					w.throttle <- bufIdx
					forfeitCommpCalc(cc)
				}()
				if _, err := cc.Write(w.tbufs[bufIdx][:]); err != nil {
					leaf <- ciderr{err: err}
					return
				}
				switch d, _, err := cc.Digest(); {
				case err != nil:
					leaf <- ciderr{err: err}
				default:
					l, err := commcid.PieceCommitmentV1ToCID(d)
					leaf <- ciderr{c: l, err: err}
				}
			}()

			w.leaves = append(w.leaves, leaf)
		}
	}
	return n, nil
}

func (w *Writer) Sum() (DataCIDSize, error) {
	// process last non-zero leaf if exists
	lastLen := w.len % int64(len(w.buf))
	rawLen := w.len

	leaves := make([]cid.Cid, len(w.leaves))
	for i, leaf := range w.leaves {
		r := <-leaf
		if r.err != nil {
			return DataCIDSize{}, xerrors.Errorf("processing leaf %d: %w", i, r.err)
		}
		leaves[i] = r.c
	}

	// process remaining bit of data
	if lastLen != 0 {
		if len(leaves) != 0 {
			copy(w.buf[lastLen:], make([]byte, int(int64(CommPBuf)-lastLen)))
			lastLen = int64(CommPBuf)
		}

		var pb []byte
		var pps uint64
		{
			// Return borrowed cc as soon as it is no longer
			// needed instead of using deffer.
			cc := borrowCommpCalc()
			if _, err := cc.Write(w.buf[:lastLen]); err != nil {
				forfeitCommpCalc(cc)
				return DataCIDSize{}, err
			}
			var err error
			if pb, pps, err = cc.Digest(); err != nil {
				forfeitCommpCalc(cc)
				return DataCIDSize{}, err
			}
			forfeitCommpCalc(cc)
		}
		p, err := commcid.PieceCommitmentV1ToCID(pb)
		if err != nil {
			return DataCIDSize{}, err
		}
		pieceSize := abi.PaddedPieceSize(pps)
		if pieceSize.Unpadded() < CommPBuf { // special case for pieces smaller than 16MiB
			return DataCIDSize{
				PayloadSize: w.len,
				PieceSize:   pieceSize,
				PieceCID:    p,
			}, nil
		}

		leaves = append(leaves, p)
	}

	// pad with zero pieces to power-of-two size
	fillerLeaves := (1 << (bits.Len(uint(len(leaves) - 1)))) - len(leaves)
	for i := 0; i < fillerLeaves; i++ {
		leaves = append(leaves, zerocomm.ZeroPieceCommitment(CommPBuf))
	}

	if len(leaves) == 1 {
		return DataCIDSize{
			PayloadSize: rawLen,
			PieceSize:   abi.PaddedPieceSize(len(leaves)) * commPBufPad,
			PieceCID:    leaves[0],
		}, nil
	}

	pieces := make([]abi.PieceInfo, len(leaves))
	for i, leaf := range leaves {
		pieces[i] = abi.PieceInfo{
			Size:     commPBufPad,
			PieceCID: leaf,
		}
	}

	p, err := GenerateUnsealedCID(abi.RegisteredSealProof_StackedDrg64GiBV1, pieces)
	if err != nil {
		return DataCIDSize{}, xerrors.Errorf("generating unsealed CID: %w", err)
	}

	return DataCIDSize{
		PayloadSize: rawLen,
		PieceSize:   abi.PaddedPieceSize(len(leaves)) * commPBufPad,
		PieceCID:    p,
	}, nil
}

var commpCalcPool = sync.Pool{
	New: func() interface{} { return new(commp.Calc) },
}

func borrowCommpCalc() *commp.Calc {
	return commpCalcPool.Get().(*commp.Calc)
}

func forfeitCommpCalc(cc *commp.Calc) {
	cc.Reset()
	commpCalcPool.Put(cc)
}
