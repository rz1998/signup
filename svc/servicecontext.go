package svc

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/rest"

	"signup/config"
)

type ServiceContext struct {
	Config     config.Config
	HttpServer *rest.Server
	DB         *sql.DB
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	ctx := &ServiceContext{
		Config:     c,
		HttpServer: nil,
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("警告: 数据库连接失败: %v (将使用内存数据)", err)
	} else {
		db.SetMaxOpenConns(c.Database.MaxOpenConns)
		db.SetMaxIdleConns(c.Database.MaxIdleConns)
		if err := db.Ping(); err != nil {
			log.Printf("警告: 数据库Ping失败: %v (将使用内存数据)", err)
			db.Close()
		} else {
			ctx.DB = db
			log.Println("数据库连接成功")
		}
	}

	return ctx, nil
}

func (s *ServiceContext) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
}
