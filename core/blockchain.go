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
	"bytes"
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/yeeco/gyee/common"
	"github.com/yeeco/gyee/log"
	"github.com/yeeco/gyee/persistent"
)

var (
	ErrBlockChainNoStorage  = errors.New("must provide block chain storage")
	ErrBlockChainIDMismatch = errors.New("chainID mismatch")
)

// BlockChain is a Data Manager that
//   created with a Storage, for chain trie/data storage
//   created with a Genesis block
//   handles tx / block lookup within the chain
//   check on  block arrival, receive block on signatures confirmation
//   notify sub routines to stop, while wait for them to stop
type BlockChain struct {
	chainID ChainID
	storage persistent.Storage
	genesis *Block

	lastBlockHash   common.Hash
	lastBlockHeight uint64

	//blockPool *BlockPool
	//txPool    *TransactionPool

	lock    sync.RWMutex
	running bool
	quitCh  chan struct{}
	wg      sync.WaitGroup
}

func NewBlockChainWithCore(core *Core) (*BlockChain, error) {
	return NewBlockChain(ChainID(core.config.Chain.ChainID), core.storage)
}

func NewBlockChain(chainID ChainID, storage persistent.Storage) (*BlockChain, error) {
	log.Info("Create New Blockchain")

	// check storage
	if storage == nil {
		return nil, ErrBlockChainNoStorage
	}

	if err := prepareStorage(storage, chainID); err != nil {
		return nil, err
	}

	bc := &BlockChain{
		chainID: chainID,
		storage: storage,
		quitCh:  make(chan struct{}),
	}

	bc.genesis = bc.GetBlockByNumber(0)
	if bc.genesis == nil {
		genesis, err := LoadGenesis(chainID)
		if err != nil {
			return nil, err
		}
		bc.genesis, err = genesis.genBlock(storage)
		if err != nil {
			return nil, err
		}
		if err = bc.AddBlock(bc.genesis); err != nil {
			return nil, err
		}
	}

	return bc, nil
}

func (bc *BlockChain) Start() error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	if bc.running {
		return errors.New("block chain already started")
	}

	log.Info("BlockChain Start...")

	go bc.loop()

	return nil
}

func (bc *BlockChain) Stop() {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	log.Info("BlockChain Stop...")
	close(bc.quitCh)
	bc.wg.Wait()
}

func (bc *BlockChain) Wait() {
	bc.lock.RLock()
	if !bc.running {
		bc.lock.RUnlock()
		return
	}
	stop := bc.quitCh
	bc.lock.RUnlock()

	// wait for close
	<-stop
}

// add a checked block to block chain
func (bc *BlockChain) AddBlock(b *Block) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	batch := bc.storage.NewBatch()

	// add block txs to storage, key "tx"+tx.hash
	if err := b.transactions.addToBatch(batch); err != nil {
		return err
	}

	// add block header to storage
	pbHeader, err := b.header.toSignedProto()
	if err != nil {
		return err
	}
	hashHeader, err := putHeader(batch, pbHeader)
	if err != nil {
		return err
	}

	// add block body to storage
	if body := b.getBody(); body != nil {
		if err := putBlockBody(batch, hashHeader, body); err != nil {
			return err
		}
	}

	// block mapping
	if err := putBlockHash2Num(batch, hashHeader, b.header.Number); err != nil {
		return err
	}
	if err := putBlockNum2Hash(batch, b.header.Number, hashHeader); err != nil {
		return err
	}

	// batch writing to storage
	if err := batch.Write(); err != nil {
		return err
	}

	bc.lastBlockHash.SetBytes(hashHeader[:])
	bc.lastBlockHeight = b.header.Number

	return nil
}

func (bc *BlockChain) GetBlockByNumber(number uint64) *Block {
	hash := getBlockNum2Hash(bc.storage, number)
	if bytes.Equal(hash[:], EmptyHash[:]) {
		return nil
	}
	return bc.GetBlockByHash(hash)
}

func (bc *BlockChain) GetBlockByHash(hash common.Hash) *Block {
	signedHeader, err := getHeader(bc.storage, hash)
	if err != nil {
		return nil
	}
	body, err := getBlockBody(bc.storage, hash)
	if err != nil {
		return nil
	}
	header := new(BlockHeader)
	if err := rlp.DecodeBytes(signedHeader.Header, header); err != nil {
		return nil
	}
	return &Block{
		header:    header,
		signature: signedHeader,
		body:      body,
	}
}

func (bc *BlockChain) loop() {
	log.Info("BlockChain loop...")
	bc.wg.Add(1)
	defer bc.wg.Done()

	for {
		select {
		case <-bc.quitCh:
			log.Info("BlockChain loop end.")
			return
		}
	}
}

func (bc *BlockChain) CurrentBlockHeight() uint64 {
	return 0
}

func (bc *BlockChain) GetValidators() map[string]uint {
	//从state取
	//测试先取一个固定的
	return map[string]uint{
		"aaaa": 1,
		"bbbb": 2,
		"cccc": 3,
		"dddd": 4,
	}
}

//非验证节点，是否需要启txPool?
