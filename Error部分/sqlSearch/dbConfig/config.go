package dbConfig

import "encoding/json"

type DBConfig struct {
	DBName         string  `json:"db_name"`
	JournalModel   string  `json:"journal_model"`
	CacheSize      string  `json:"cache_size"`
	Synchronous    string  `json:"synchronous"`
	Mode           string  `json:"mode"`
	DBPath         string  `json:"path"`
	TbName         string  `json:"tb_name"`
}

func (c *DBConfig) Encode() ([]byte,error)  {
	return json.Marshal(c)
}

func (c *DBConfig) Decode(data []byte) error {
	return json.Unmarshal(data,c)
}

func (c *DBConfig) Config() string  {
	return "file:"+ c.DBPath+"/"+c.DBName+".db?"+"_journal_mode="+c.JournalModel+"&"+
		"_cache_size="+c.CacheSize+"&"+"_synchronous"+c.Synchronous+"&"+"mode="+c.Mode
}

func (c *DBConfig) DBPathInfo() string {
	return c.DBPath+"/"+c.DBName+".db"
}