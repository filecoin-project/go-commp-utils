package pieceio

import (
	"context"
	"io"
	"os"

	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-car"
	"github.com/ipld/go-ipld-prime"

	"github.com/filecoin-project/go-commp-utils/writer"
	"github.com/filecoin-project/go-multistore"
	"github.com/filecoin-project/go-state-types/abi"
)

var log = logging.Logger("pieceio")

type PreparedCar interface {
	Size() uint64
	Dump(w io.Writer) error
}

type CarIO interface {
	// WriteCar writes a given payload to a CAR file and into the passed IO stream
	WriteCar(ctx context.Context, bs ReadStore, payloadCid cid.Cid, node ipld.Node, w io.Writer, userOnNewCarBlocks ...car.OnNewCarBlockFunc) error

	// PrepareCar prepares a car so that its total size can be calculated without writing it to a file.
	// It can then be written with PreparedCar.Dump
	PrepareCar(ctx context.Context, bs ReadStore, payloadCid cid.Cid, node ipld.Node, userOnNewCarBlocks ...car.OnNewCarBlockFunc) (PreparedCar, error)

	// LoadCar loads blocks into the a store from a given CAR file
	LoadCar(bs WriteStore, r io.Reader) (cid.Cid, error)
}

type pieceIO struct {
	carIO      CarIO
	bs         blockstore.Blockstore
	multiStore MultiStore
}

type MultiStore interface {
	Get(i multistore.StoreID) (*multistore.Store, error)
}

func NewPieceIO(carIO CarIO, bs blockstore.Blockstore, multiStore MultiStore) PieceIO {
	return &pieceIO{carIO, bs, multiStore}
}

func (pio *pieceIO) GeneratePieceReader(payloadCid cid.Cid, selector ipld.Node, storeID *multistore.StoreID, userOnNewCarBlocks ...car.OnNewCarBlockFunc) (io.ReadCloser, uint64, error, <-chan error) {
	bstore, err := pio.bstore(storeID)
	if err != nil {
		return nil, 0, err, nil
	}
	preparedCar, err := pio.carIO.PrepareCar(context.Background(), bstore, payloadCid, selector, userOnNewCarBlocks...)
	if err != nil {
		return nil, 0, err, nil
	}
	pieceSize := uint64(preparedCar.Size())
	r, w, err := os.Pipe()
	if err != nil {
		return nil, 0, err, nil
	}
	writeErr := make(chan error, 1)
	go func() {
		werr := preparedCar.Dump(w)
		err := w.Close()
		if werr == nil && err != nil {
			werr = err
		}
		writeErr <- werr
	}()
	return r, pieceSize, nil, writeErr
}

func (pio *pieceIO) GeneratePieceCommitment(rt abi.RegisteredSealProof, payloadCid cid.Cid, selector ipld.Node, storeID *multistore.StoreID, userOnNewCarBlocks ...car.OnNewCarBlockFunc) (cid.Cid, abi.UnpaddedPieceSize, error) {
	bstore, err := pio.bstore(storeID)
	if err != nil {
		return cid.Undef, 0, err
	}
	preparedCar, err := pio.carIO.PrepareCar(context.Background(), bstore, payloadCid, selector, userOnNewCarBlocks...)
	if err != nil {
		return cid.Undef, 0, err
	}

	commpWriter := &writer.Writer{}
	err = preparedCar.Dump(commpWriter)
	if err != nil {
		return cid.Undef, 0, err
	}
	dataCIDSize, err := commpWriter.Sum()
	if err != nil {
		return cid.Undef, 0, err
	}
	return dataCIDSize.PieceCID, dataCIDSize.PieceSize.Unpadded(), nil
}

func (pio *pieceIO) ReadPiece(storeID *multistore.StoreID, r io.Reader) (cid.Cid, error) {
	bstore, err := pio.bstore(storeID)
	if err != nil {
		return cid.Undef, err
	}
	return pio.carIO.LoadCar(bstore, r)
}

func (pio *pieceIO) bstore(storeID *multistore.StoreID) (blockstore.Blockstore, error) {
	if storeID == nil {
		return pio.bs, nil
	}
	store, err := pio.multiStore.Get(*storeID)
	if err != nil {
		return nil, err
	}
	return store.Bstore, nil
}
