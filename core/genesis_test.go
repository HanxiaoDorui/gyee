/*
 *  Copyright (C) 2017 gyee authors
 *
 *  This file is part of the gyee library.
 *
 *  The gyee library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The gyee library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with the gyee library.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package core

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func getGenesis(t *testing.T, cid ChainID) *Genesis {
	g, err := LoadGenesis(cid)
	if err != nil {
		t.Fatalf("failed to decode genesis toml: %v", err)
	}
	if count := len(g.Consensus.Tetris.Validators); count <= 0 {
		t.Fatalf("wrong validator count %v", count)
	}
	return g
}

func getGenesisBlock(t *testing.T, cid ChainID) (*Genesis, *Block) {
	genesis := getGenesis(t, cid)
	block, err := genesis.genBlock(nil)
	if err != nil {
		t.Fatalf("failed to build genesis block %v", err)
	}
	hash, err := block.header.Hash()
	if err != nil {
		t.Fatalf("failed to calculate block hash %v", err)
	}
	hashHex := hex.EncodeToString(hash)
	if hashHex != genesis.Hash {
		// ON TEST FAIL:
		//   - if genesis block content is INTENTIONALLY modified, with confidence:
		//     modify genesis toml accordingly
		//   - if not, some encoding / hashing change is not backward compatible and MUST BE FIXED
		t.Fatalf("genesis hash mismatch: need %v got %v", genesis.Hash, hashHex)
	}
	return genesis, block
}

func TestLoadGenesis(t *testing.T) {
	genesis := getGenesis(t, MainNetID)

	fmt.Printf("%v\n", genesis)
}

func TestMainNetGenesis(t *testing.T) {
	genesis, block := getGenesisBlock(t, MainNetID)

	if count := len(genesis.Consensus.Tetris.Validators); count != 4 {
		t.Errorf("wrong validator count %v", count)
	}

	fmt.Printf("block: %v\n", block)
}

func TestTestNetGenesis(t *testing.T) {
	genesis, block := getGenesisBlock(t, TestNetID)

	if count := len(genesis.Consensus.Tetris.Validators); count != 4 {
		t.Errorf("wrong validator count %v", count)
	}

	fmt.Printf("block: %v\n", block)
}
