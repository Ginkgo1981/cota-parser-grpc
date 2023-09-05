package service

import (
	"encoding/hex"
	"strings"

	pb "github.com/nervina-labs/cota-entries-parser/api/cotaparser/v1"
	"github.com/nervina-labs/cota-entries-parser/pkg/molecule"
)

type CotaParser struct{}

func NewCotaParser() *CotaParser {
	return &CotaParser{}
}

func (p CotaParser) Parse(witnessStr string, version string) (*pb.GetCotaEntriesResponse, error) {

	bytes, err := hex.DecodeString(strings.TrimPrefix(witnessStr, "0x"))
	if err != nil {
		return nil, err
	}

	witnessArgs := molecule.WitnessArgsFromSliceUnchecked(bytes)
	if witnessArgs.InputType().IsNone() {
		return &pb.GetCotaEntriesResponse{}, nil
	}

	inputType, err := witnessArgs.InputType().IntoBytes()
	if err != nil {
		return nil, erTT
	}

	actionType := inputType.RawData()[0]
	switch actionType {
	case 1:
		return p.GenerateDefineResp(inputType)

	default:
		// TODO add error type
		return nil, errors.New("invalid action type")


}

func (p CotaParser) GenerateDefineResp(inputType *molecule.Bytes) (*pb.GetCotaEntriesResponse, error) {
	entries := molecule.DefineCotaNFTEntriesFromSliceUnchecked(inputType.RawData()[1:])
	define, err := p.generatedDefine(entries)
	if err != nil {
		return nil, err
	}
	return &pb.GetCotaEntriesResponse{
		Entry: &pb.CotaEntry{
			Entry: &pb.CotaEntry_Define{
				Define: define,
			},
		},
	}, nil
}
