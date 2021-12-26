package params

import (
	"crypto/ecdsa"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"pdx-chain-so/pkg/pdx-chain/common"
	"regexp"
	"sync"
	"time"
)

const (
	NormalBlock = 1
	CommitBlock = 2
	MaxLimit    = 400000000 //4亿limit最大msg大小是10MB p2p包广播限制就是10MB
	Account     = "chaindata"
)

type Org struct {
	Name   string   `json:"name"`
	NodeCa []string `json:"node_ca"` //node ca file
	UserCa []string `json:"user_ca"`
}

type ConsortiumConfObj struct {
	Name string `json:"name"`
	Orgs []Org  `json:"orgs"`
	//Cert string `json:"cert"`//current cert chain
	UserAuth bool `json:"user_auth"`
	DappAuth bool `json:"dapp_auth"`
}

var (
	//账户state的合约簇id
	Config               *cli.Context
	Consortium           bool
	ConsortiumConf 	   *ConsortiumConfObj
	ConsortiumDir        string
	DataDir              string
	Cert                 string //local node cert
	CertRootId           string
	RootIdCaMap          map[string]string
	UserCertPublicKeyMap map[string]struct{}
	OrgNameMapKeyIds     map[string][]string // org name -> []keyId
	OrgNameMapUserPublicKeys     map[string][]string // org name -> []publicKey
	OrgNameMapNodePublicKeys     map[string][]string // org name -> []publicKey
	Perf                 bool
	CCEnable             bool // 是否可以调用cc
	StartTime            time.Time
	MinerPrivateKey      *ecdsa.PrivateKey
	Syncing              *int32
	DistantOfCfd         uint64
	Gasless              bool //default
	NetWorkId            int
	Etherbase            common.Address
	NumMasters           int32
	BlockDelay           int32 // in millisecond
	//block confirmation window
	Cnfw                *big.Int
	Cfd                 *big.Int
	BecameQuorumLimt    int32
	ConsensusQuorumLimt int32
	//Majority            int
	PerQuorum         bool //每个commit更新委员会
	UIP1              *big.Int
	UIP7              *big.Int
	Sm2Crypto         bool
	GasLimit          uint64 //gaslimit
	GasStrategyRefuse bool   //default no
	MaxMsgSize        uint64 //限制最大的msg大小 计算公式 gaslimit/40=msg最大尺寸 (40是0字节占用的gas大小)
	ASyncExecute      bool   //异步执行
	ChainId           *big.Int
	MyId              string   //当前节点p256得的标致
	PeerBlacklist     sync.Map // peer 黑名单列表 nodeAddr为key
	KeyIdBlacklist	  sync.Map // keyId黑名单列表 keyId作为key ！note:只是存在全局变量里，不需要再存state中，因为该ca已经被删除，再次链接证书链验证会不通过
	PeerConnList      sync.Map // 接入节点证书的公钥对应的addr->peerId
	PeerKeyIdList     sync.Map // 接入节点证书的ca keyId->[]node addr
)

var (
	SubKeyidReg = regexp.MustCompile(`X509v3 Subject Key Identifier: \n(.*)`)
	CertReg = regexp.MustCompile(`-{5}BEGIN CERTIFICATE-{5}\s+([^-]+)-{5}END CERTIFICATE-{5}`)
	AuthKeyidReg = regexp.MustCompile(`keyid:(.*)`)
)