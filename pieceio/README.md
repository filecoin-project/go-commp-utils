# pieceio

The pieceio module is a collection of structs for generating piece commitments (a.k.a. CommP) and 
storing pieces for storage market deals. It is used by the 
[`storagemarket`](../storagemarket) module.

## Installation
```bash
go get github.com/filecoin-project/go-commp-utils/pieceio
```

## PieceIO
`PieceIO` is used by [`storagemarket`](../storagemarket) client for proposing deals. 

**To initialize a PieceIO:**
```go
package pieceio

func NewPieceIO(carIO CarIO, bs blockstore.Blockstore) PieceIO
```
**Parameters**
* `carIO` is a [CarIO](#CarIO) from this module
* `bs` is an IPFS blockstore for storing and retrieving data for deals. See
 [github.com/ipfs/go-ipfs-blockstore](github.com/ipfs/go-ipfs-blockstore).

## CarIO
CarIO is a utility module that wraps [github.com/ipld/go-car](https://github.com/ipld/go-car) for use by storagemarket.

**To initialize a CarIO:**
```go
package cario

func NewCarIO() pieceio.CarIO
```

Please the [tests](pieceio_test.go) for more information about expected behavior.