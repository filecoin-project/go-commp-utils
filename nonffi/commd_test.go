package nonffi

import (
	"testing"

	"github.com/filecoin-project/go-commp-utils/zerocomm"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
)

func TestGenerateUnsealedCID(t *testing.T) {

	/*
		Testing live sector data with the help of a fellow SP

		~$ lotus-miner sectors status --log 139074
			SectorID:	139074
			Status:		Proving
			CIDcommD:	baga6ea4seaqiw3gbmstmexb7sqwkc5r23o3i7zcyx5kr76pfobpykes3af62kca
			...
			Precommit:	bafy2bzacec3dyxgqfbjekvnbin6uhcel7adis576346bi3tahp64bhijeiymy
			Commit:		bafy2bzacecafq4ksrjzlhjagxkrrpycmfpjo5ch62s3tbq7gr5rop75fuqhwk
			Deals:		[3755444 0 0 3755443 3755442 3755608 3755679 3755680 0 3755754 3755803 3755883 0 3755882 0 0 0]
	*/

	expCommD := cidMustParse("baga6ea4seaqiw3gbmstmexb7sqwkc5r23o3i7zcyx5kr76pfobpykes3af62kca")

	commD, _ := GenerateUnsealedCID(
		abi.RegisteredSealProof_StackedDrg32GiBV1_1, // 32G sector SP
		[]abi.PieceInfo{
			{PieceCID: cidMustParse("baga6ea4seaqknzm22isnhsxt2s4dnw45kfywmhenngqq3nc7jvecakoca6ksyhy"), Size: 256 << 20},  // https://filfox.info/en/deal/3755444
			{PieceCID: cidMustParse("baga6ea4seaqnq6o5wuewdpviyoafno4rdpqnokz6ghvg2iyeyfbqxgcwdlj2egi"), Size: 1024 << 20}, // https://filfox.info/en/deal/3755443
			{PieceCID: cidMustParse("baga6ea4seaqpixk4ifbkzato3huzycj6ty6gllqwanhdpsvxikawyl5bg2h44mq"), Size: 512 << 20},  // https://filfox.info/en/deal/3755442
			{PieceCID: cidMustParse("baga6ea4seaqaxwe5dy6nt3ko5tngtmzvpqxqikw5mdwfjqgaxfwtzenc6bgzajq"), Size: 512 << 20},  // https://filfox.info/en/deal/3755608
			{PieceCID: cidMustParse("baga6ea4seaqpy33nbesa4d6ot2ygeuy43y4t7amc4izt52mlotqenwcmn2kyaai"), Size: 1024 << 20}, // https://filfox.info/en/deal/3755679
			{PieceCID: cidMustParse("baga6ea4seaqphvv4x2s2v7ykgc3ugs2kkltbdeg7icxstklkrgqvv72m2v3i2aa"), Size: 256 << 20},  // https://filfox.info/en/deal/3755680
			{PieceCID: cidMustParse("baga6ea4seaqf5u55znk6jwhdsrhe37emzhmehiyvjxpsww274f6fiy3h4yctady"), Size: 512 << 20},  // https://filfox.info/en/deal/3755754
			{PieceCID: cidMustParse("baga6ea4seaqa3qbabsbmvk5er6rhsjzt74beplzgulthamm22jue4zgqcuszofi"), Size: 1024 << 20}, // https://filfox.info/en/deal/3755803
			{PieceCID: cidMustParse("baga6ea4seaqiekvf623muj6jpxg6vsqaikyw3r4ob5u7363z7zcaixqvfqsc2ji"), Size: 256 << 20},  // https://filfox.info/en/deal/3755883
			{PieceCID: cidMustParse("baga6ea4seaqhsewv65z2d4m5o4vo65vl5o6z4bcegdvgnusvlt7rao44gro36pi"), Size: 512 << 20},  // https://filfox.info/en/deal/3755882

			// GenerateUnsealedCID does not "fill a sector", do it here to match the SP provided sector commD
			{PieceCID: zerocomm.ZeroPieceCommitment(abi.PaddedPieceSize(8 << 30).Unpadded()), Size: 8 << 30},
			{PieceCID: zerocomm.ZeroPieceCommitment(abi.PaddedPieceSize(16 << 30).Unpadded()), Size: 16 << 30},
		},
	)

	if commD != expCommD {
		t.Fatalf("calculated commd for sector f01392893:139074 as %s, expected %s", commD, expCommD)
	}
}

// Delete below when this ships in a stable release https://github.com/ipfs/go-cid/pull/139
func cidMustParse(v interface{}) cid.Cid {
	c, err := cid.Parse(v)
	if err != nil {
		panic(err)
	}
	return c
}
