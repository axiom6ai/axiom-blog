package globalInit

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"

	"axiom-blog/config"
)

var (
	Db   *gorm.DB
	dsn  string
	conf = config.Conf
)

func DbInit() {
	dsn = genDsn()
	//程序启动打开数据库连接
	if err := initEngine(); err != nil {
		panic(err)
	}
}

// 将数据库连接信息连接成字符串
func genDsn() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		conf.Postgres.Host,
		conf.Postgres.User,
		conf.Postgres.Password,
		conf.Postgres.Database,
		conf.Postgres.Port)
}

func initEngine() error {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Duration(conf.Postgres.SlowThreshold) * time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.LogLevel(conf.Postgres.LogLevel),                  // 日志级别
			IgnoreRecordNotFoundError: true,                                                     // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,                                                    // 禁用彩色打印
		},
	)
	var err error
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		//默认关闭事务
		SkipDefaultTransaction: true,
		Logger:                 newLogger,
		NamingStrategy: schema.NamingStrategy{
			//表复数禁用
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalf("open db err:%v", err)
		return err
	}
	postgresDB, _ := Db.DB()
	if err = postgresDB.Ping(); err != nil {
		panic(fmt.Sprintf("error ping DB:%v", err))
	}

	//连接池设置
	postgresDB.SetMaxIdleConns(conf.Postgres.MaxIdle)
	postgresDB.SetMaxOpenConns(conf.Postgres.MaxConn)

	return nil
}

func Transaction() (tx *gorm.DB) {
	tx = Db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	return tx
}
