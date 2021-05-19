package main

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"sort"

	logging "github.com/ipfs/go-log/v2"

	ffi "github.com/filecoin-project/filecoin-ffi"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/specs-actors/v5/actors/runtime/proof"
)

var log = logging.Logger("agg")

func main() {
	logging.SetAllLoggers(logging.LevelInfo)
	f, err := os.Open("proofs.json")
	if err != nil {
		panic(err)
	}
	d := json.NewDecoder(f)

	var infosBySize [][]proof.SealVerifyInfo
	{
		infos := make(map[uint64][]proof.SealVerifyInfo)
		var info []proof.SealVerifyInfo
		for {
			err = d.Decode(&info)
			if err != nil {
				break
			}
			mid := uint64(info[0].Miner)
			if info[0].SealProof != 8 {
				continue
			}
			infos[mid] = append(infos[mid], info...)
		}
		infosBySize = make([][]proof.SealVerifyInfo, 0, len(infos))
		for _, info := range infos {
			infosBySize = append(infosBySize, info)
		}
		sort.Slice(infosBySize, func(i, j int) bool {
			return len(infosBySize[i]) > len(infosBySize[j])
		})
	}

	fOut, err := os.OpenFile("agg.ndjson", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	e := json.NewEncoder(fOut)
	if err != nil {
		panic(err)
	}
	maxProofs := 14
	Npoints := 512
	eps := math.Pow(2, (float64(maxProofs)-3.32192809489)/float64(Npoints))
	maxProofs = 1 << maxProofs

	j := 0
	k := 0

	for N := maxProofs; N >= 10; {
		n := N / 10
		for ; len(infosBySize[j]) > n; j++ {
		}
		k = j
		for ; len(infosBySize[k]) == n && k+1 < len(infosBySize); k++ {
		}
		k--
		if k-j < 4 {
			j -= 4
			if j < 0 {
				j = 0
			}
			k += 4 - (k - j)
		}

		x := j + rand.Intn(k-j)
		info := infosBySize[x]
		//for _, inf := range info {
		//ok, err := ffi.VerifySeal(inf)
		//if !ok || err != nil {
		//fmt.Printf("verif %t error: %+v", ok, err)
		//return
		//}
		//}
		sealInfos := make([]proof.AggregateSealVerifyInfo, n)
		proofs := make([][]byte, n)
		for i := 0; i < n; i++ {
			in := info[i%len(info)]
			sealInfos[i] = proof.AggregateSealVerifyInfo{
				Number:                in.Number,
				Randomness:            in.Randomness,
				InteractiveRandomness: in.InteractiveRandomness,
				SealedCID:             in.SealedCID,
				UnsealedCID:           in.UnsealedCID,
			}
			proofs[i] = in.Proof
		}
		rand.Shuffle(len(proofs), func(i, j int) {
			sealInfos[i], sealInfos[j] = sealInfos[j], sealInfos[i]
			proofs[i], proofs[j] = proofs[j], proofs[i]
		})

		aggInfo := proof.AggregateSealVerifyProofAndInfos{
			Miner:          info[0].Miner,
			SealProof:      info[0].SealProof,
			AggregateProof: abi.RegisteredAggregationProof_SnarkPackV1,
			Infos:          sealInfos,
		}
		aggInfo.Proof, err = ffi.AggregateSealProofs(aggInfo, proofs)
		if err != nil {
			panic(err)
		}

		log.Infof("i: %d, j: %d, len(j): %d, k: %d, len(k): %d", n, j, len(infosBySize[j]),
			k, len(infosBySize[k]))
		//ok, err := ffi.VerifyAggregateSeals(aggInfo)
		//if !ok || err != nil {
		//fmt.Printf("verif %t error: %+v", ok, err)
		//continue
		//}
		e.Encode(aggInfo)

		N2 := int(float64(N)/eps + 0.5)
		if N != N2 {
			N = N2
		} else {
			N--
		}
	}

	fOut.Close()

}
