package database

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"log"
	"sync"
	"time"
)

type Model struct {
	CreateTime time.Time	`db:"create_time" json:"createTime"`
	UpdateTime time.Time	`db:"update_time" json:"updateTime"`
	UpdateUser string		`db:"update_user" json:"updateUser"`
	DeleteTime int64		`db:"delete_time" json:"-"`
}

type IMysqlManager interface {
	Open()
	GetDbByUserName(userName string) (db *sqlx.DB, err error)
	GetAllDbs() (dbs []*sqlx.DB)
	GenerateId() int64
}

type MysqlConfig struct{
	WorkerId 	int64			`json:"worker_id" validate:"min=0"`
	Type 		string 			`json:"type" validate:"required,oneof=mysql"`
	User 		string			`json:"user" validate:"required,gte=1"`
	Password 	string			`json:"password" validate:"required,gte=1"`
	Hosts 		map[int]string	`json:"hosts" validate:"required,gte=1,dive,keys,min=0,endkeys,tcp_addr"`
	Name 		string			`json:"name" validate:"required,gte=1"`
	MaxIdle 	int				`json:"max_idle" validate:"required,min=1"`
	MaxOpen 	int				`json:"max_open" validate:"required,min=1"`
	MaxLifetime time.Duration	`json:"max_lifetime" validate:"required,gte=1"`
}

var (
	_mysqlConfig MysqlConfig
	_mysqlManagerOnce sync.Once
	_mysqlManager     IMysqlManager
)

/**
 * 根据配置接口对象初始化
 */
func SetupMysql(mysqlConfig MysqlConfig)  {
	_mysqlConfig = mysqlConfig
}

func Mysql() IMysqlManager {
	_mysqlManagerOnce.Do(func() {
		_mysqlManager = NewMysqlManager(_mysqlConfig)
	})
	return _mysqlManager
}

func NewMysqlManager(mysqlConfig MysqlConfig) IMysqlManager {
	idWorker, err := foundation.NewWorker(mysqlConfig.WorkerId)
	if err != nil {
		panic(err)
	}
	return &mysqlManagerImpl{
		config:          mysqlConfig,
		dbMap:           make(map[int]*sqlx.DB),
		dataCenterCount: 0,
		idWorker: 		 idWorker,
	}
}

type mysqlManagerImpl struct {
	opened bool
	config MysqlConfig
	/** 数据中心id 关联 db Map **/
	dbMap map[int]*sqlx.DB
	/** 数据中心数量 **/
	dataCenterCount int
	idWorker foundation.SnowId
}

/**
 * 数据库open
 * db: DB 对象
 */
func (this *mysqlManagerImpl) Open() {
	if this.opened {
		return
	}
	if err := validate.ValidateParameter(this.config); err != nil {
		log.Fatalf("mysql config validate failed:%s", err.Error())
		return
	}
	for id, host := range this.config.Hosts {
		var dbLink = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
			this.config.User, this.config.Password, host, this.config.Name)
		db, err := sqlx.Open(this.config.Type, dbLink)
		if err != nil {
			panic(err)
		}
		db.SetMaxIdleConns(this.config.MaxIdle)
		db.SetMaxOpenConns(this.config.MaxOpen)
		db.SetConnMaxLifetime(this.config.MaxLifetime)
		this.dbMap[id] = db
		this.dataCenterCount += 1
	}
	this.opened = true
}

/**
 * 根据用户名基因确定数据库对象
 */
func (this *mysqlManagerImpl) GetDbByUserName(userName string) (db *sqlx.DB, err error) {
	dna, err := Dna(userName)
	if err != nil {
		return nil, err
	}
	dataCenterId := dna % this.dataCenterCount
	return this.dbMap[dataCenterId], nil
}

/**
 * 获取所有数据库对象
 */
func (this *mysqlManagerImpl) GetAllDbs() (dbs []*sqlx.DB) {
	for _, v := range this.dbMap {
		dbs = append(dbs, v)
	}
	return dbs
}

/**
 * 分布式id生成
 */
func (this *mysqlManagerImpl) GenerateId() int64 {
	return this.idWorker.GetId()
}

//var (
//	tokenBucketOnce sync.Once
//	tokenBucket		*ratelimit.Bucket
//)
//
//func getTokenBucket() *ratelimit.Bucket {
//	tokenBucketOnce.Do(func() {
//		tokenBucket	= ratelimit.NewBucket(config.Server.RateLimitFillInterval, config.Server.RateLimitCapacity)
//	})
//	return tokenBucket
//}

