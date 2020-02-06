package pieceio

import (
	"io"

	"github.com/filecoin-project/go-fil-markets/filestore"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
)

type WriteStore interface {
	Put(blocks.Block) error
}

type ReadStore interface {
	Get(cid.Cid) (blocks.Block, error)
}

// PieceIO converts between payloads and pieces
type PieceIO interface {
	GeneratePieceCommitment(payloadCid cid.Cid, selector ipld.Node) ([]byte, uint64, error)
	ReadPiece(r io.Reader) (cid.Cid, error)
}

type PieceIOWithStore interface {
	PieceIO
	GeneratePieceCommitmentToFile(payloadCid cid.Cid, selector ipld.Node) ([]byte, filestore.Path, uint64, error)
}
