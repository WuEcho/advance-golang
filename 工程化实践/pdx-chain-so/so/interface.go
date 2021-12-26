package so

import "container/list"

type Call interface {
	Run(stub interface{}) ([]byte,error)
}

type StubInterface interface {
	//GetArgs returns the arguments intended for the so Run as an array of byte arrays
	GetArgs() [][]byte
	//GetStringArgs returns the arguments intended for the so Run as a string array.
	GetStringArgs() []string
	//GetFunctionAndParameters returns the first argument as the function name
	//and the rest of arguments as parameters in a string array.
	GetFunctionAndParameters() (string,[][]byte)
	//GetHistoryForKey returns a history of key values across time.
	//For each historic key update,the historic value and associated block num.
	GetHistoryForKey(key string,start,end uint64) (*list.Element,error)
	//GetState returns the value of the specified `key` from the
	//ledger.Note that GetState doesn`t read data from the writest,which
	//has not been committed to the state.
	GetState(key []byte) ([]byte,error)
	//PutState puts the specified `key` and `value` into the state.simple keys
	//must not be an empty string and must not start with a null character
	PutState(key []byte,value []byte) error
	//DelState records the specified `key` to be deleted in the state.
	DelState(key []byte) error
	//GetStateByRange(startKey,endKey string)
}

