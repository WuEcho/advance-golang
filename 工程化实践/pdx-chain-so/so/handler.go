package so

import (
	"container/list"
	"errors"
	"pdx-chain-so/pkg/pdx-chain/common"
	"pdx-chain-so/pkg/pdx-chain/core/public"
	"pdx-chain-so/pkg/pdx-chain/core/state"
	"pdx-chain-so/pkg/pdx-chain/rlp"
)

var (
	SoCallError_Input_Error             = errors.New("Input is illegal")
	SoCallError_NoResult                = errors.New("No result by that given key")
	SoCallError_Key_Value_NotMatch      = errors.New("Key value not match")
	SoCallError_Start_FinishNum_Illegal = errors.New("Get History start and finish num is illegal")
	SoCallError_History_Encode_Error    = errors.New("Get History result encode error")
	SoCallError_History_Limit_Reached   = errors.New("Get History limit reached")
)

type messageType int

const (
	SoCall_GET_STATE   messageType = 1
	SoCall_PUT_STATE   messageType = 2
	Socall_PUT_STATES  messageType = 3
	SoCall_DEL_STATE   messageType = 4
	SoCall_GET_HISTORY messageType = 5
)

var MaxSize = 5 * 1024 * 1024 * 1024

type Handler struct {
	db  *state.MStateDB
	res chan *CallSoResMessage
}

func NewHandler(db *state.MStateDB) *Handler {
	return &Handler{
		db:  db,
		res: make(chan *CallSoResMessage),
	}
}

func (h *Handler) handle(message *CallSoSendMessage) (res *CallSoResMessage) {
	var resMessage *CallSoResMessage
	resMessage = validityInput(message.inputs[0])
	if resMessage != nil {
		return resMessage
	}
	key := common.BytesToHash(message.inputs[0])

	switch message.callType {

	case SoCall_GET_STATE:
		v := h.db.GetPDXState(message.address, key)
		if len(v) == 0 {
			resMessage = &CallSoResMessage{
				res: nil,
				err: SoCallError_NoResult,
			}
		}

		println("get state called")

	case SoCall_PUT_STATE:
		if len(message.inputs) != 2 {
			return &CallSoResMessage{
				res: nil,
				err: SoCallError_Key_Value_NotMatch,
			}
		}

		h.db.SetPDXState(message.address, key, message.inputs[1])

		resMessage = &CallSoResMessage{
			err: nil,
		}
		println("put state called")

	case Socall_PUT_STATES:
		if len(message.inputs)%2 != 0 {
			return &CallSoResMessage{
				res: nil,
				err: SoCallError_Key_Value_NotMatch,
			}
		}

		for i := 0; i < len(message.inputs)%2; i++ {
			k := message.inputs[i*2]
			tpk := common.BytesToHash(k)
			v := message.inputs[i*2+1]
			h.db.SetPDXState(message.address, tpk, v)
		}

		resMessage = &CallSoResMessage{
			err: nil,
		}
		println("put states called")

	case SoCall_DEL_STATE:
		h.db.SetPDXState(message.address, key, []byte{})

		resMessage = &CallSoResMessage{
			err: nil,
		}
		println("del state called")

	case SoCall_GET_HISTORY:
		//查询历史
		if len(message.inputs)%3 != 0 {
			return &CallSoResMessage{
				res: nil,
				err: SoCallError_Key_Value_NotMatch,
			}
		}

		start := message.inputs[1]
		finish := message.inputs[2]

		startNum := common.ByteToUint64(start)
		finishNum := common.ByteToUint64(finish)

		if startNum < 0 {
			return &CallSoResMessage{
				res: nil,
				err: SoCallError_Start_FinishNum_Illegal,
			}
		}

		if finishNum < 0 {
			return &CallSoResMessage{
				res: nil,
				err: SoCallError_Start_FinishNum_Illegal,
			}
		}

		if startNum > finishNum {
			return &CallSoResMessage{
				res: nil,
				err: SoCallError_Start_FinishNum_Illegal,
			}
		}

		var totalSize uint64
		recordList := new(list.List)
		for i := startNum; i < finishNum; i++ {
			block := public.BC.GetBlockByNumber(uint64(i))
			stateDb, err := public.BC.StateAt(block.Header().Root)
			if err != nil {
				break
			}
			v := stateDb.GetPDXState(message.address, key)
			if len(v) == 0 {
				continue
			}

			rEle := &RecordElement{
				value: v,
				num: block.NumberU64(),
			}
			recordList.PushFront(rEle)
			s := rEle.Size()
			totalSize += s
			if totalSize >= uint64(MaxSize) {
				println("err", "get history max size reached")
				//log.Error("get history max size reached")
				break
			}
		}
		data, err := rlp.EncodeToBytes(recordList)
		if err != nil {
			return &CallSoResMessage{
				res: nil,
				err: SoCallError_History_Encode_Error,
			}
		}

		resMessage = &CallSoResMessage{
			res: data,
			err: SoCallError_History_Limit_Reached,
		}
	}

	return resMessage
}

func validityInput(input []byte) *CallSoResMessage {
	if len(input) == 0 {
		return &CallSoResMessage{
			res: nil,
			err: SoCallError_Input_Error,
		}
	}
	return nil
}

type RecordElement struct {
	value   []byte
	num     uint64
}

func (r *RecordElement) GetValue() []byte {
	return r.value
}

func (r *RecordElement) GetNum() uint64 {
	return r.num
}

func (r *RecordElement) SetValue(v []byte) {
	r.value = v
}

func (r *RecordElement) SetNum(n uint64) {
	r.num = n
}

func (r *RecordElement) Size() uint64 {
	data,err := rlp.EncodeToBytes(r)
	if err != nil {
		return 0
	}
	return uint64(len(data))
}
