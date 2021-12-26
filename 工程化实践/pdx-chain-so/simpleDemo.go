package main

import (
	"errors"
	"pdx-chain-so/so"
)

var Simple simple

type simple struct {}

func (s *simple) Run(stub interface{}) ([]byte,error)  {
	v,ok := stub.(so.StubInterface)
	if !ok {
		return nil, nil
	}
	function,inputs := v.GetFunctionAndParameters()
	if function == "queryPersonInfo" {
		return s.queryPersonInfo(v,inputs)
	}else if function == "savePersonInfo" {
		return s.savePersonInfo(v,inputs)
	}
	return []byte{},nil
}

func (s *simple) queryPersonInfo(stub so.StubInterface,args [][]byte) ([]byte,error) {
	if len(args) != 1 {
		return []byte{},errors.New("查询输入有误")
	}

	key := args[0]
	v,err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	return v,nil
}

func (s *simple) savePersonInfo(stub so.StubInterface,args [][]byte) ([]byte,error) {
	if len(args) != 2 {
		return nil,errors.New("输入有误")
	}
	key := args[0]
	v := args[1]
	err := stub.PutState(key,v)
	if err != nil {
		return nil, err
	}
	return v,nil
}

func main()  {
	
}