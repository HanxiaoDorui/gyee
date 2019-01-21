// Copyright (C) 2019 gyee authors
//
// This file is part of the gyee library.
//
// The gyee library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gyee library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the gyee library.  If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
	"github.com/yeeco/gyee/common"
	"github.com/yeeco/gyee/core/pb"
	sha3 "github.com/yeeco/gyee/crypto/hash"
	"github.com/yeeco/gyee/persistent"
)

const (
	KeyChainID = "ChainID"

	KeyPrefixTx     = "tx"
	KeyPrefixHeader = "bh"
	KeyPrefixBody   = "bb"
)

func prepareStorage(storage persistent.Storage, id ChainID) error {
	key := keyChainID()
	encChainID, err := storage.Get(key)
	if err != nil {
		if err != persistent.ErrKeyNotFound {
			return err
		}
		// not found
		encChainID := make([]byte, 4)
		binary.BigEndian.PutUint32(encChainID, uint32(id))
		return storage.Put(key, encChainID)
	}
	decoded := binary.BigEndian.Uint32(encChainID)
	if ChainID(decoded) != id {
		return ErrBlockChainIDMismatch
	}
	return nil
}

func getHeader(getter persistent.Getter, hash common.Hash) (*corepb.SignedBlockHeader, error) {
	enc, err := getter.Get(keyHeader(hash))
	if err != nil {
		if err == persistent.ErrKeyNotFound {
			return nil, nil
		}
		return nil, err
	}
	pb := new(corepb.SignedBlockHeader)
	if err := proto.Unmarshal(enc, pb); err != nil {
		return nil, err
	}
	return pb, nil
}

func putHeader(putter persistent.Putter, header *corepb.SignedBlockHeader) (common.Hash, error) {
	enc, err := proto.Marshal(header)
	if err != nil {
		return EmptyHash, err
	}
	hash := common.BytesToHash(sha3.Sha3256(enc))
	if err := putter.Put(keyHeader(hash), enc); err != nil {
		return EmptyHash, err
	}
	return hash, nil
}

func getBlockBody(getter persistent.Getter, hash common.Hash) (*corepb.BlockBody, error) {
	enc, err := getter.Get(keyBlockBody(hash))
	if err != nil {
		if err == persistent.ErrKeyNotFound {
			return nil, nil
		}
		return nil, err
	}
	pb := new(corepb.BlockBody)
	if err := proto.Unmarshal(enc, pb); err != nil {
		return nil, err
	}
	return pb, nil
}

func putBlockBody(putter persistent.Putter, hash common.Hash, body *corepb.BlockBody) error {
	enc, err := proto.Marshal(body)
	if err != nil {
		return err
	}
	if err := putter.Put(keyBlockBody(hash), enc); err != nil {
		return err
	}
	return nil
}

func keyChainID() []byte {
	return []byte(KeyChainID)
}

func keyHeader(hash common.Hash) []byte {
	return append([]byte(KeyPrefixHeader), hash[:]...)
}

func keyBlockBody(hash common.Hash) []byte {
	return append([]byte(KeyPrefixBody), hash[:]...)
}
