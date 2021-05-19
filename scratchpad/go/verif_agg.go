package main

import (
	"encoding/json"
	"os"

	ffi "github.com/filecoin-project/filecoin-ffi"
	"github.com/filecoin-project/specs-actors/v5/actors/runtime/proof"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("agg")

func main() {
	logging.SetAllLoggers(logging.LevelInfo)
	f, err := os.Open("agg.json")
	if err != nil {
		panic(err)
	}

	d := json.NewDecoder(f)
	var aggregates []proof.AggregateSealVerifyProofAndInfos
	err = d.Decode(&aggregates)
	if err != nil {
		panic(err)
	}

	for i, agg := range aggregates {
		res, err := ffi.VerifyAggregateSeals(agg)
		if !res {
			log.Errorf("verifying aggregate at %d failed", i)
		}
		if err != nil {
			log.Errorf("verifying aggregate returned error: %+v", err)
		}
		log.Infof("agg %d done", i)
	}

}
