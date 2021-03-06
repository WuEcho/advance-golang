// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package state provides a caching layer atop the Ethereum state trie.
package state

import (
	"errors"
	"fmt"
	"math/big"
	"pdx-chain-so/pkg/pdx-chain/common"
	"pdx-chain-so/pkg/pdx-chain/crypto"
	//"pdx-chain-so/pkg/pdx-chain/log"
	"pdx-chain-so/pkg/pdx-chain/rlp"
	"runtime"
	"sort"
	"sync"
)

type revision struct {
	id           int
	journalIndex int
}

var (
	// emptyState is the known hash of an empty state trie entry.
	emptyState = crypto.Keccak256Hash(nil)

	// emptyCode is the known hash of the empty EVM bytecode.
	emptyCode = crypto.Keccak256Hash(nil)

	// add by liangc : 对于常用的 code 进行 lru 缓存
	//codeCache, _ = lru.New(256)
)

// StateDBs within the ethereum protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	db   Database
	trie Trie

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects      map[common.Address]*stateObject
	stateObjectsDirty map[common.Address]struct{}

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// The refund counter, also used by state transitioning.
	refund uint64

	thash, bhash common.Hash
	txIndex      int
	logSize      uint
	preimages    map[common.Hash][]byte

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionId int

	lock sync.Mutex
}

// Create a new state from a given trie.
func New(root common.Hash, db Database) (*StateDB, error) {
	tr, err := db.OpenTrie(root)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:                db,
		trie:              tr,
		stateObjects:      make(map[common.Address]*stateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}, nil
}

func NewCopy(state *StateDB) *StateDB {
	return &StateDB{
		db:                state.db,
		trie:              state.trie,
		stateObjects:      make(map[common.Address]*stateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}
}

func (self *StateDB) GetSelectiveStateObjectRoot(addr common.Address) (root common.Hash, err error) {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		//self.setError(errors.New("selective state object not find"))
		return common.Hash{}, errors.New("selective state object not find")
	}
	root = stateObject.trie.Hash()
	//log.Debug("get selective state object root", "hash()", root.String(), "root", stateObject.data.Root.String())
	return root, nil

}

func (self *StateDB) TrieRoot() common.Hash {
	return self.trie.Hash()
}

// setError remembers the first non-nil error it is called with.
func (self *StateDB) setError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (self *StateDB) Error() error {
	return self.dbErr
}


// AddPreimage records a SHA3 preimage seen by the VM.
func (self *StateDB) AddPreimage(hash common.Hash, preimage []byte) {
	if _, ok := self.preimages[hash]; !ok {
		self.journal.append(addPreimageChange{hash: hash})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		self.preimages[hash] = pi
	}
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (self *StateDB) Preimages() map[common.Hash][]byte {
	return self.preimages
}

func (self *StateDB) AddRefund(gas uint64) {
	self.journal.append(refundChange{prev: self.refund})
	self.refund += gas
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (self *StateDB) Exist(addr common.Address) bool {
	return self.getStateObject(addr) != nil
}

// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (self *StateDB) Empty(addr common.Address) bool {
	so := self.getStateObject(addr)
	return so == nil || so.empty()
}

// Retrieve the balance from the given address or 0 if object not found
func (self *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return common.Big0
}

func (self *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0
}


//// add by liangc
//func setCodeCache(addr common.Address, obj codeobj) {
//	codeCache.Add(addr, obj)
//}
//
//// add by liangc
//func getCodeCache(addr common.Address) (codeobj, bool) {
//	if data, ok := codeCache.Get(addr); ok {
//		return data.(codeobj), true
//	}
//	return codeobj{}, false
//}
//
////del code cache
//func delCodeCache(addr common.Address) {
//	codeCache.Remove(addr)
//}

func (self *StateDB) GetCode(addr common.Address) []byte {
	//if codeobj, ok := getCodeCache(addr); ok {
	//	return codeobj.code
	//}
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		code := stateObject.Code(self.db)
		// add by liangc
		//if code != nil && len(code) > 0 {
		//	setCodeCache(addr, codeobj{code, stateObject.CodeHash()})
		//}
		return code
	}
	return nil
}

func (self *StateDB) GetCodeSize(addr common.Address) int {
	// add by liangc
	//if codeobj, ok := getCodeCache(addr); ok {
	//	return len(codeobj.code)
	//}
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return 0
	}
	//if stateObject.code != nil {
	//	// add by liangc
	//	setCodeCache(addr, codeobj{stateObject.code, stateObject.CodeHash()})
	//	return len(stateObject.code)
	//}
	code, err := self.db.ContractCode(stateObject.addrHash, common.BytesToHash(stateObject.CodeHash()))
	if err != nil {
		self.setError(err)
	}
	// add by liangc
	//if code != nil && len(code) > 0 {
	//	setCodeCache(addr, codeobj{code, stateObject.CodeHash()})
	//}
	return len(code)
}

func (self *StateDB) GetCodeHash(addr common.Address) common.Hash {
	// add by liangc
	//if codeobj, ok := getCodeCache(addr); ok {
	//	return common.BytesToHash(codeobj.hash)
	//}
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	// add by liangc
	//if stateObject.code != nil && len(stateObject.code) > 0 {
	//	setCodeCache(addr, codeobj{stateObject.code, stateObject.CodeHash()})
	//}
	return common.BytesToHash(stateObject.CodeHash())
}

func (self *StateDB) GetState(addr common.Address, bhash common.Hash) common.Hash {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetState(self.db, bhash)
	}
	return common.Hash{}
}

func (self *StateDB) GetPDXState(a common.Address, b common.Hash) []byte {
	stateObject := self.getStateObject(a)
	if stateObject != nil {
		return stateObject.GetPDXState(self.db, b)
	}
	return []byte{}
}

// Database retrieves the low level database supporting the lower level trie ops.
func (self *StateDB) Database() Database {
	return self.db
}

// StorageTrie returns the storage trie of an account.
// The return value is a copy and is nil for non-existent accounts.
func (self *StateDB) StorageTrie(addr common.Address) Trie {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return nil
	}
	cpy := stateObject.deepCopy(self)
	return cpy.updateTrie(self.db)
}

func (self *StateDB) HasSuicided(addr common.Address) bool {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (self *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (self *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
	}
}

func (self *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}

func (self *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

func (self *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
	}
}

func (self *StateDB) SetState(addr common.Address, key, value common.Hash) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(self.db, key, value)
	}
}

func (self *StateDB) SetPDXState(addr common.Address, key common.Hash, value []byte) {
	// add by liangc : 这个地方如果不加 SetState 的调用 SetPDXState 就不会生效，因为 SetPDXState 没有将 addr 加入 dirty >>>>
	// self.SetState(addr, common.Hash{1}, common.Hash{1})
	// add by liangc : 这个地方如果不加 SetState 的调用 SetPDXState 就不会生效，因为 SetPDXState 没有将 addr 加入 dirty <<<<
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetPDXState(self.db, key, value)
	}
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (self *StateDB) Suicide(addr common.Address) bool {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return false
	}
	self.journal.append(suicideChange{
		account:     &addr,
		prev:        stateObject.suicided,
		prevbalance: new(big.Int).Set(stateObject.Balance()),
	})
	stateObject.markSuicided()
	stateObject.data.Balance = new(big.Int)

	return true
}

//
// Setting, updating & deleting state object methods.
//

// updateStateObject writes the given object to the trie.
func (self *StateDB) updateStateObject(stateObject *stateObject) {
	addr := stateObject.Address()
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	self.setError(self.trie.TryUpdate(addr[:], data))
}

// deleteStateObject removes the given object from the state trie.
func (self *StateDB) deleteStateObject(stateObject *stateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	self.setError(self.trie.TryDelete(addr[:]))
}

// Retrieve a state object given by the address. Returns nil if not found.
func (self *StateDB) getStateObject(addr common.Address) (stateObject *stateObject) {
	// Prefer 'live' objects.
	if obj := self.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}

	// Load the object from the database.
	enc, err := self.trie.TryGet(addr[:])
	if len(enc) == 0 {
		self.setError(err)
		return nil
	}
	var data Account
	if err := rlp.DecodeBytes(enc, &data); err != nil {
		//log.Error("Failed to decode state object", "addr", addr, "err", err)
		return nil
	}
	// Insert into the live set.
	obj := newObject(self, addr, data)
	self.setStateObject(obj)
	return obj
}

func (self *StateDB) setStateObject(object *stateObject) {
	self.stateObjects[object.Address()] = object
}

// Retrieve a state object or create a new state object if nil.
func (self *StateDB) GetOrNewStateObject(addr common.Address) *stateObject {
	stateObject := self.getStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = self.createObject(addr)
	}
	stateObject.db.journal = self.journal
	return stateObject
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (self *StateDB) createObject(addr common.Address) (newobj, prev *stateObject) {
	prev = self.getStateObject(addr)
	newobj = newObject(self, addr, Account{})
	newobj.setNonce(0) // sets the object to dirty
	if prev == nil {
		self.journal.append(createObjectChange{account: &addr})
	} else {
		self.journal.append(resetObjectChange{prev: prev})
	}
	self.setStateObject(newobj)
	return newobj, prev
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (self *StateDB) CreateAccount(addr common.Address) {
	new, prev := self.createObject(addr)
	if prev != nil {
		new.setBalance(prev.data.Balance)
	}
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (self *StateDB) Copy() *StateDB {
	self.lock.Lock()
	defer self.lock.Unlock()

	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		db:                self.db,
		trie:              self.db.CopyTrie(self.trie),
		stateObjects:      make(map[common.Address]*stateObject, len(self.journal.dirties)),
		stateObjectsDirty: make(map[common.Address]struct{}, len(self.journal.dirties)),
		refund:            self.refund,
		logSize:           self.logSize,
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}
	// Copy the dirty states, logs, and preimages
	for addr := range self.journal.dirties {
		// As documented [here](https://pdx-chain/pull/16485#issuecomment-380438527),
		// and in the Finalise-method, there is a case where an object is in the journal but not
		// in the stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we need to check for
		// nil
		if object, exist := self.stateObjects[addr]; exist {
			state.stateObjects[addr] = object.deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}
	// Above, we don't copy the actual journal. This means that if the copy is copied, the
	// loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies
	for addr := range self.stateObjectsDirty {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = self.stateObjects[addr].deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}


	for hash, preimage := range self.preimages {
		state.preimages[hash] = preimage
	}
	return state
}

// Snapshot returns an identifier for the current revision of the state.
func (self *StateDB) Snapshot() int {
	id := self.nextRevisionId
	self.nextRevisionId++
	self.validRevisions = append(self.validRevisions, revision{id, self.journal.length()})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (self *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(self.validRevisions), func(i int) bool {
		return self.validRevisions[i].id >= revid
	})
	if idx == len(self.validRevisions) || self.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := self.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	self.journal.revert(self, snapshot)
	self.validRevisions = self.validRevisions[:idx]
}

// GetRefund returns the current value of the refund counter.
func (self *StateDB) GetRefund() uint64 {
	return self.refund
}

// Finalise finalises the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (s *StateDB) Finalise(deleteEmptyObjects bool) {
	for addr := range s.journal.dirties {
		stateObject, exist := s.stateObjects[addr]

		//if !exist {
		//	// ripeMD is 'touched' at block 1714175, in tx 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
		//	// That tx goes out of gas, and although the notion of 'touched' does not exist there, the
		//	// touch-event will still be recorded in the journal. Since ripeMD is a special snowflake,
		//	// it will persist in the journal even though the journal is reverted. In this special circumstance,
		//	// it may exist in `s.journal.dirties` but not in `s.stateObjects`.
		//	// Thus, we can safely ignore it here
		//	continue
		//}

		if !exist {
			//如果找不到就找一下
			stateObject = s.getStateObject(addr)
		}

		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			s.deleteStateObject(stateObject)
		} else {
			stateObject.updateRoot(s.db)
			s.updateStateObject(stateObject)
		}
		s.stateObjectsDirty[addr] = struct{}{}
	}
	// Invalidate journal because reverting across transactions is not allowed.
	s.clearJournalAndRefund()
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func PrintCallerName() string {
	pc, _, _, _ := runtime.Caller(3)
	return runtime.FuncForPC(pc).Name()
}
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	//t := time.Now()
	//defer func() {log.Debug("IntermediateRoot()", "elapsed", time.Since(t), "caller", PrintCallerName())}()

	s.Finalise(deleteEmptyObjects)
	h := s.trie.Hash()
	return h
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
func (self *StateDB) Prepare(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

func (s *StateDB) clearJournalAndRefund() {
	s.journal = newJournal()
	s.validRevisions = s.validRevisions[:0]
	s.refund = 0
}

// add by liangc : 预处理时需要获取全部临时状态并缓存
func (self *StateDB) StateObjects() map[common.Address]*stateObject {
	return self.stateObjects
}

func (self *StateDB) StateObjectsDirty() map[common.Address]struct{} {
	return self.stateObjectsDirty
}

func (self *StateDB) Setjournal(obj *stateObject) {
	obj.db.journal = self.journal
}

func (self *StateDB) Journal() map[common.Address]int {
	return self.journal.dirties
}

func (self *StateDB) SetDirties(dirties map[common.Address]int) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.journal.dirties = dirties
}
