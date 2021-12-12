package main

import (
	"os"
	"sqlSearch/dbConfig"
	"sqlSearch/sqlRest"
)

func main()  {
	path,err := os.Getwd()
	if err != nil {
		println("err:",err.Error())
	}
	conf := dbConfig.DBConfig{DBName: "file",DBPath: path,
		JournalModel: "WAL",CacheSize: "8000",Synchronous: "0",Mode:"rwc",TbName: "person"}

	println("conf",conf.DBPathInfo())
      println("conf info:",conf.Config())
	sqldb,err := sqlRest.CreateDB(conf)
	if err != nil {
		println("err:",err.Error())
	}

	err = sqlRest.CreateTbWithName(sqldb,conf)
	if err != nil {
		println("err:",err.Error())
	}
	err = sqlRest.InsertData(sqldb,conf,"echo")
	if err != nil {
		println("err:",err.Error())
	}

	v,err := sqlRest.QueryData(sqldb,conf,"echo")
	if err != nil {
		println("err:",err.Error())
	}

	println("v:",v)

	sqlRest.CloseDb(sqldb)

}