package data

import (
	"errors"
	"fmt"
	"strings"

	"review-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"review-service/internal/data/query"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewDB, NewData, NewReviewRepo)

// Data .
type Data struct {
	query *query.Query
	log   *log.Helper
}

// NewData .
func NewData(db *gorm.DB, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	// 为生成的GEN设置数据库连接对象
	query.SetDefault(db)

	return &Data{query: query.Q, log: log.NewHelper(logger)}, cleanup, nil
}

func NewDB(c *conf.Data) (*gorm.DB, error) {
	if c == nil {
		panic(errors.New("GEN: connectDB fail, need cfg"))
	}
	fmt.Printf("c.Database.Driver: %+v\n", c.Database.Driver)
	switch strings.ToLower(c.Database.Driver) {
	case "mysql":
		dsn := c.Database.Source
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Printf("mysql connect err: %+v", err)
			return nil, err
		}
		return db, nil
	default:
		return nil, errors.New("unsupported database driver: " + c.Database.Driver)
	}
}
