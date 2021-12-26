package so

import (
	"container/list"
	"pdx-chain-so/pkg/pdx-chain/common"
	"pdx-chain-so/pkg/pdx-chain/rlp"
)

type SOCallStub struct {
	handler *Handler
	args    [][]byte
	address  common.Address
}

func NewSoCallStub(handler *Handler,args [][]byte,address common.Address) *SOCallStub {
	return &SOCallStub{
		handler: handler,
		args: args,
		address: address,
	}
}

func (s *SOCallStub) GetState(key []byte) ([]byte,error) {
	mess := &CallSoSendMessage{
		inputs: [][]byte{key},
		callType: SoCall_GET_STATE,
		address: s.address,
	}

	res := s.handler.handle(mess)
	return res.res,res.err
}

func (s *SOCallStub) PutState(key []byte,value []byte) error {
	mess := &CallSoSendMessage{
		inputs: [][]byte{key,value},
		callType: SoCall_PUT_STATE,
		address: s.address,
	}
	res := s.handler.handle(mess)
	return res.err
}

func (s *SOCallStub) DelState(key []byte) error {
	mess := &CallSoSendMessage{
		inputs: [][]byte{key},
		callType: SoCall_DEL_STATE,
		address: s.address,
	}
	res := s.handler.handle(mess)
	return res.err
}

func (s *SOCallStub) GetArgs() [][]byte {
	return s.args
}

func (s *SOCallStub) GetStringArgs() []string {
	args := s.GetArgs()
	strargs := make([]string,0,len(args))
	for _,barg := range args {
		strargs = append(strargs,string(barg))
	}
	return strargs
}

func (s *SOCallStub) GetFunctionAndParameters() (function string,params [][]byte) {
	allArgs := s.GetArgs()
	function = ""
	params = [][]byte{}
	if len(allArgs) >= 1 {
		function = string(allArgs[0])
		params = allArgs[1:]
	}
	return
}

func (s *SOCallStub) GetHistoryForKey(key string,start,end uint64) (*list.Element,error) {
	mess := &CallSoSendMessage{
		inputs: [][]byte{[]byte(key),common.Uint64ToByte(start),common.Uint64ToByte(end)},
		callType: SoCall_GET_HISTORY,
		address: s.address,
	}
	res := s.handler.handle(mess)
	if len(res.res) > 0 {
		return nil,res.err
	}
	histList := new(list.List)
	err := rlp.DecodeBytes(res.res,histList)
	if err != nil {
		return nil, err
	}

	ele := histList.Front()
	return ele,nil
}