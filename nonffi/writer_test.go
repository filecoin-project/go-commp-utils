package nonffi

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"testing"

	commcid "github.com/filecoin-project/go-fil-commcid"
	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/stretchr/testify/require"
)

func TestWriterZero(t *testing.T) {
	for i, s := range []struct {
		writes []int
		expect abi.PaddedPieceSize
	}{
		{writes: []int{200}, expect: 256},
		{writes: []int{200, 200}, expect: 512},
		{writes: []int{int(CommPBuf)}, expect: commPBufPad},
		{writes: []int{int(CommPBuf) * 2}, expect: 2 * commPBufPad},
		{writes: []int{int(CommPBuf), int(CommPBuf), int(CommPBuf)}, expect: 4 * commPBufPad},
		{writes: []int{int(CommPBuf), int(CommPBuf), int(CommPBuf), int(CommPBuf), int(CommPBuf), int(CommPBuf), int(CommPBuf), int(CommPBuf), int(CommPBuf)}, expect: 16 * commPBufPad},

		{writes: []int{200, int(CommPBuf)}, expect: 2 * commPBufPad},
	} {
		s := s
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			w := &Writer{}
			var rawSum int64
			for _, write := range s.writes {
				rawSum += int64(write)
				_, err := w.Write(make([]byte, write))
				require.NoError(t, err)
			}

			p, err := w.Sum()
			require.NoError(t, err)
			require.Equal(t, rawSum, p.PayloadSize)
			require.Equal(t, s.expect, p.PieceSize)
		})
	}
}

func TestWriterData(t *testing.T) {
	dataLen := float64(CommPBuf) * 6.78
	data, _ := io.ReadAll(io.LimitReader(rand.Reader, int64(dataLen)))

	cc := new(commp.Calc)
	_, err := cc.Write(data)
	require.NoError(t, err)
	pb, _, err := cc.Digest()
	require.NoError(t, err)
	exp, err := commcid.PieceCommitmentV1ToCID(pb)
	require.NoError(t, err)

	w := &Writer{}
	_, err = io.Copy(w, bytes.NewReader(data))
	require.NoError(t, err)

	res, err := w.Sum()
	require.NoError(t, err)

	require.Equal(t, exp.String(), res.PieceCID.String())
}

func BenchmarkWriterZero(b *testing.B) {
	buf := make([]byte, int(CommPBuf)*b.N)
	b.SetBytes(int64(CommPBuf))
	b.ResetTimer()

	w := &Writer{}

	_, err := w.Write(buf)
	require.NoError(b, err)
	o, err := w.Sum()

	b.StopTimer()

	require.NoError(b, err)
	require.Equal(b, int64(CommPBuf)*int64(b.N), o.PayloadSize)
}
