package service

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	pb "github.com/ginkgo1981/entries-parser/api/cotaparser/v1"
	"github.com/ginkgo1981/entries-parser/pkg/molecule"
)

func NewCotaParser() *CotaParser {
	return &CotaParser{}
}

type CotaParser struct{}

func (p CotaParser) Parse(witnessStr string, version string) (*pb.GetCotaEntriesResponse, error) {
	log.Println("witness: ", witnessStr)
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
		return nil, err
	}

	actionType := inputType.RawData()[0]
	switch actionType {
	case 1:
		return p.GenerateDefineResp(inputType)
	case 2:
		return p.GenerateMintResp(inputType, version)
	case 3:
	case 4:
	case 5:
	case 6:
	case 7:
	case 8:
	default:
		// TODO add error type
		return nil, errors.New("invalid action type")
	}

	return nil, err
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

func (p CotaParser) GenerateMintResp(inputType *molecule.Bytes, version string) (*pb.GetCotaEntriesResponse, error) {
	switch version {
	case "0":
		entries := molecule.MintCotaNFTEntriesFromSliceUnchecked(inputType.RawData()[1:])
		mint, err := p.generatedV0Mint(entries)
		if err != nil {
			return nil, err
		}
		return &pb.GetCotaEntriesResponse{
			Entry: &pb.CotaEntry{
				Entry: &pb.CotaEntry_Mint{
					Mint: mint,
				},
			},
		}, nil
	case "1":
		entries := molecule.MintCotaNFTV1EntriesFromSliceUnchecked(inputType.RawData()[1:])
		mint, err := p.generatedV1Mint(entries)
		if err != nil {
			return nil, err
		}
		return &pb.GetCotaEntriesResponse{
			Entry: &pb.CotaEntry{
				Entry: &pb.CotaEntry_Mint{
					Mint: mint,
				},
			},
		}, nil
	default:
		// TODO add error type
		return nil, errors.New("invalid cota version")
	}
}

func (p CotaParser) generatedV0Mint(entries *molecule.MintCotaNFTEntries) (*pb.Mint, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("generatedV0Mint panic: %v\n", err)
		}
	}()
	defineKeys := &pb.DefineCotaNFTKeyVec{}
	for i := uint(0); i < entries.DefineKeys().Len(); i++ {
		key := entries.DefineKeys().Get(i)
		defineKeys.Keys = append(defineKeys.Keys, &pb.DefineCotaNFTId{
			SmtType: "0x" + hex.EncodeToString(key.SmtType().RawData()),
			CotaId:  "0x" + hex.EncodeToString(key.CotaId().RawData()),
		})
	}

	defineOldValues := &pb.DefineCotaNFTValueVec{}
	for i := uint(0); i < entries.DefineOldValues().Len(); i++ {
		value := entries.DefineOldValues().Get(i)
		defineOldValues.Values = append(defineOldValues.Values, &pb.DefineCotaNFTValue{
			Total:     binary.BigEndian.Uint32(value.Total().RawData()),
			Issued:    binary.BigEndian.Uint32(value.Issued().RawData()),
			Configure: uint32(value.Configure().AsSlice()[0]),
		})
	}

	defineValues := &pb.DefineCotaNFTValueVec{}
	for i := uint(0); i < entries.DefineNewValues().Len(); i++ {
		value := entries.DefineNewValues().Get(i)
		defineValues.Values = append(defineValues.Values, &pb.DefineCotaNFTValue{
			Total:     binary.BigEndian.Uint32(value.Total().RawData()),
			Issued:    binary.BigEndian.Uint32(value.Issued().RawData()),
			Configure: uint32(value.Configure().AsSlice()[0]),
		})
	}

	withdrawalKeys := &pb.WithdrawalCotaNFTKeyVec{}
	for i := uint(0); i < entries.WithdrawalKeys().Len(); i++ {
		key := entries.WithdrawalKeys().Get(i)
		withdrawalKeys.CotaNFTIds = append(withdrawalKeys.CotaNFTIds, &pb.CotaNFTId{
			SmtType: "0x" + hex.EncodeToString(key.SmtType().RawData()),
			CotaId:  "0x" + hex.EncodeToString(key.CotaId().RawData()),
			Index:   binary.BigEndian.Uint32(key.Index().RawData()),
		})
	}

	withdrawalValues := &pb.WithdrawalCotaNFTValueVec{}
	for i := uint(0); i < entries.WithdrawalValues().Len(); i++ {
		value := entries.WithdrawalValues().Get(i)
		toLock := molecule.ScriptFromSliceUnchecked(value.ToLock().RawData())
		hashType, err := p.hashType(toLock.HashType().AsSlice()[0])
		if err != nil {
			return nil, err
		}
		withdrawalValues.Values = append(withdrawalValues.Values, &pb.WithdrawalCotaNFTValue{
			NftInfo: &pb.CotaNFTInfo{
				Configure:      uint32(value.NftInfo().Configure().AsSlice()[0]),
				State:          uint32(value.NftInfo().State().AsSlice()[0]),
				Characteristic: "0x" + hex.EncodeToString(value.NftInfo().Characteristic().RawData()),
			},
			ToLock: &pb.Script{
				CodeHash: "0x" + hex.EncodeToString(toLock.CodeHash().RawData()),
				HashType: hashType,
				Args:     "0x" + hex.EncodeToString(toLock.Args().RawData()),
			},
			OutPoint: "0x" + hex.EncodeToString(value.OutPoint().RawData()),
		})
	}

	proof := "0x" + hex.EncodeToString(entries.Proof().RawData())
	action := string(entries.Action().RawData())
	action = strings.Map(fixUtf, action)

	return &pb.Mint{
		DefineKeys:      defineKeys,
		DefineOldValues: defineOldValues,
		DefineNewValues: defineValues,
		WithdrawalKeys: &pb.Mint_V0Keys{
			V0Keys: withdrawalKeys,
		},
		WithdrawValues: &pb.Mint_V0Values{
			V0Values: withdrawalValues,
		},
		Proof:  proof,
		Action: action,
	}, nil
}

func (p CotaParser) generatedV1Mint(entries *molecule.MintCotaNFTV1Entries) (*pb.Mint, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("generatedV1Mint panic: %v\n", err)
		}
	}()
	defineKeys := &pb.DefineCotaNFTKeyVec{}
	for i := uint(0); i < entries.DefineKeys().Len(); i++ {
		key := entries.DefineKeys().Get(i)
		defineKeys.Keys = append(defineKeys.Keys, &pb.DefineCotaNFTId{
			SmtType: "0x" + hex.EncodeToString(key.SmtType().RawData()),
			CotaId:  "0x" + hex.EncodeToString(key.CotaId().RawData()),
		})
	}

	defineOldValues := &pb.DefineCotaNFTValueVec{}
	for i := uint(0); i < entries.DefineOldValues().Len(); i++ {
		value := entries.DefineOldValues().Get(i)
		defineOldValues.Values = append(defineOldValues.Values, &pb.DefineCotaNFTValue{
			Total:     binary.BigEndian.Uint32(value.Total().RawData()),
			Issued:    binary.BigEndian.Uint32(value.Issued().RawData()),
			Configure: uint32(value.Configure().AsSlice()[0]),
		})
	}

	defineValues := &pb.DefineCotaNFTValueVec{}
	for i := uint(0); i < entries.DefineNewValues().Len(); i++ {
		value := entries.DefineNewValues().Get(i)
		defineValues.Values = append(defineValues.Values, &pb.DefineCotaNFTValue{
			Total:     binary.BigEndian.Uint32(value.Total().RawData()),
			Issued:    binary.BigEndian.Uint32(value.Issued().RawData()),
			Configure: uint32(value.Configure().AsSlice()[0]),
		})
	}

	withdrawalKeys := &pb.WithdrawalCotaNFTKeyV1Vec{}
	for i := uint(0); i < entries.WithdrawalKeys().Len(); i++ {
		key := entries.WithdrawalKeys().Get(i)
		withdrawalKeys.Keys = append(withdrawalKeys.Keys, &pb.WithdrawalCotaNFTKeyV1{
			NftIds: &pb.CotaNFTId{
				SmtType: "0x" + hex.EncodeToString(key.NftId().SmtType().RawData()),
				CotaId:  "0x" + hex.EncodeToString(key.NftId().CotaId().RawData()),
				Index:   binary.BigEndian.Uint32(key.NftId().Index().RawData()),
			},
			OutPoint: "0x" + hex.EncodeToString(key.OutPoint().RawData()),
		})
	}

	withdrawalValues := &pb.WithdrawalCotaNFTValueV1Vec{}
	for i := uint(0); i < entries.WithdrawalValues().Len(); i++ {
		value := entries.WithdrawalValues().Get(i)
		toLock := molecule.ScriptFromSliceUnchecked(value.ToLock().RawData())
		hashType, err := p.hashType(toLock.HashType().AsSlice()[0])
		if err != nil {
			return nil, err
		}
		withdrawalValues.Values = append(withdrawalValues.Values, &pb.WithdrawalCotaNFTValueV1{
			NftInfo: &pb.CotaNFTInfo{
				Configure:      uint32(value.NftInfo().Configure().AsSlice()[0]),
				State:          uint32(value.NftInfo().State().AsSlice()[0]),
				Characteristic: "0x" + hex.EncodeToString(value.NftInfo().Characteristic().RawData()),
			},
			ToLock: &pb.Script{
				CodeHash: "0x" + hex.EncodeToString(toLock.CodeHash().RawData()),
				HashType: hashType,
				Args:     "0x" + hex.EncodeToString(toLock.Args().RawData()),
			},
		})
	}

	proof := "0x" + hex.EncodeToString(entries.Proof().RawData())
	action := string(entries.Action().RawData())
	action = strings.Map(fixUtf, action)

	return &pb.Mint{
		DefineKeys:      defineKeys,
		DefineOldValues: defineOldValues,
		DefineNewValues: defineValues,
		WithdrawalKeys: &pb.Mint_V1Keys{
			V1Keys: withdrawalKeys,
		},
		WithdrawValues: &pb.Mint_V1Values{
			V1Values: withdrawalValues,
		},
		Proof:  proof,
		Action: action,
	}, nil
}

func (p CotaParser) hashType(hashType uint8) (string, error) {
	switch hashType {
	case 0:
		return "data", nil
	case 1:
		return "type", nil
	case 2:
		return "data1", nil
	default:
		return "", errors.New("invalid hash type")
	}
}

func (p CotaParser) generatedDefine(entries *molecule.DefineCotaNFTEntries) (*pb.Define, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("generatedDefine panic: %v\n", err)
		}
	}()
	if entries.DefineKeys().Len() == 0 {
		return nil, errors.New("no define entries")
	}
	keys := &pb.DefineCotaNFTKeyVec{}
	values := &pb.DefineCotaNFTValueVec{}
	for i := uint(0); i < entries.DefineKeys().Len(); i++ {
		key := entries.DefineKeys().Get(i)
		keys.Keys = append(keys.Keys, &pb.DefineCotaNFTId{
			SmtType: "0x" + hex.EncodeToString(key.SmtType().RawData()),
			CotaId:  "0x" + hex.EncodeToString(key.CotaId().RawData()),
		})
		value := entries.DefineValues().Get(i)
		values.Values = append(values.Values, &pb.DefineCotaNFTValue{
			Total:     binary.BigEndian.Uint32(value.Total().RawData()),
			Issued:    binary.BigEndian.Uint32(value.Issued().RawData()),
			Configure: uint32(value.Configure().AsSlice()[0]),
		})
	}
	proof := "0x" + hex.EncodeToString(entries.Proof().RawData())
	action := string(entries.Action().RawData())
	action = strings.Map(fixUtf, action)

	return &pb.Define{
		DefineKeys:   keys,
		DefineValues: values,
		Proof:        proof,
		Action:       action,
	}, nil
}

func fixUtf(r rune) rune {
	if r == utf8.RuneError {
		return -1
	}
	return r
}
