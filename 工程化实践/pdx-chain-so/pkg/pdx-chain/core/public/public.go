package public

import (
	"pdx-chain-so/pkg/pdx-chain/common"
	"pdx-chain-so/pkg/pdx-chain/core/state"
	"pdx-chain-so/pkg/pdx-chain/core/types"
)

var BC PublicBlockChain

type PublicBlockChain interface {
	GetBlockByNumber(number uint64) *types.Block
	GetCommitBlock(height uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
	State() (*state.StateDB, error) // 此处应该限制只能够查询state
}