package main

// gorm gen configure

import (
	"errors"
	"fmt"

	"review-service/internal/conf"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"gorm.io/gen"
)

func configData() *conf.Data {
	flagconf := "../../configs/config.yaml"

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	return bc.Data
}

// NewDB 创建数据库连接
func connectDB(c *conf.Data) *gorm.DB {
	if c == nil {
		panic(errors.New("GEN: connectDB fail, need cfg"))
	}
	fmt.Printf("c.Database.Driver: %+v\n", c.Database.Driver)
	switch c.Database.Driver {
	case "mysql":
		dsn := c.Database.Source
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		return db
	default:
		panic("unsupported database driver: " + c.Database.Driver)
	}
}

func autoSync(db *gorm.DB) {
	// 指定生成代码的具体相对目录(相对当前文件)，默认为：./query
	// 默认生成需要使用WithContext之后才可以查询的代码，但可以通过设置gen.WithoutContext禁用该模式
	g := gen.NewGenerator(gen.Config{
		// 默认会在 OutPath 目录生成CRUD代码，并且同目录下生成 model 包
		// 所以OutPath最终package不能设置为model，在有数据库表同步的情况下会产生冲突
		// 若一定要使用可以通过ModelPkgPath单独指定model package的名称
		OutPath: "../../internal/data/query",
		/* ModelPkgPath: "dal/model"*/

		// gen.WithoutContext：禁用WithContext模式
		// gen.WithDefaultQuery：生成一个全局Query对象Q
		// gen.WithQueryInterface：生成Query接口
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable: true,
	})

	// 通常复用项目中已有的SQL连接配置db(*gorm.DB)
	// 非必需，但如果需要复用连接时的gorm.Config或需要连接数据库同步表信息则必须设置
	g.UseDB(db)

	// 从连接的数据库为所有表生成Model结构体和CRUD代码
	// 也可以手动指定需要生成代码的数据表
	g.ApplyBasic(g.GenerateAllTable()...)

	// 通过ApplyInterface添加为表添加自定义方法
	// g.ApplyInterface(func(model.Querier) {}, g.GenerateModel("review_reply_info"))
	// g.GenerateModel("review_reply_info", gen.WithMethod(model.CommonModel{}))

	// 执行并生成代码
	g.Execute()
}

func operateDB(db *gorm.DB) {
	// query.SetDefault(db)

	// operateReviewReplyInfo := query.ReviewReplyInfo
	// ret, err := operateReviewReplyInfo.WithContext(context.Background()).Where(operateReviewReplyInfo.ID.Eq(1)).First()
	// if err != nil {
	// 	fmt.Printf("err: %+v\n", err)
	// }
	// fmt.Printf("reviewReplyInfo: %+v\n", ret.ExtJSONMap())

	// 通过自定义方法查询
	// ret2, err := operateReviewReplyInfo.WithContext(context.Background()).GetByVersion(0)
	// if err != nil {
	// 	fmt.Printf("err: %+v\n", err)
	// }
	// for _, v := range ret2 {
	// 	fmt.Printf("v.ID: %+v, version: %+v\n", v.ID, v.Version)
	// }
}

func main() {
	conf := configData()
	db := connectDB(conf)
	autoSync(db)
	operateDB(db)
}
