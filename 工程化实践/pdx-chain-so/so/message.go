package so

import "pdx-chain-so/pkg/pdx-chain/common"

type CallSoSendMessage struct {
	inputs       [][]byte
	callType     messageType
	address      common.Address
}

type CallSoResMessage struct {
	res         []byte
	err         error
}