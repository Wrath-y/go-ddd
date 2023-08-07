package db

import (
	"context"
	"go-ddd/infrastructure/util/def"
	"go-ddd/infrastructure/util/logging"
	"log"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

var Orm *gorm.DB

func Setup() {
	Orm = newMysqlDB("default")
}

func newMysqlDB(store string) *gorm.DB {
	dbViper := viper.Sub("mysql." + store)
	if dbViper == nil {
		log.Fatal("mysql配置缺失", store)
	}

	address := dbViper.GetString("address")
	username := dbViper.GetString("username")
	password := dbViper.GetString("password")
	database := dbViper.GetString("database")
	maxIdleConns := dbViper.GetInt("max_idle_conns")
	maxOpenConns := dbViper.GetInt("max_open_conns")
	timeout := dbViper.GetString("timeout")
	if timeout == "" {
		timeout = "20"
	}
	// 慢sql阈值
	slowThreshold := dbViper.GetDuration("slow_threshold")
	if slowThreshold == 0 {
		slowThreshold = time.Second
	}

	dsn := username + ":" + password + "@tcp(" + address + ")/" + database +
		"?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&loc=Local&timeout=" + timeout + "s"
	orm, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: &gormLog{
			slowThreshold: slowThreshold,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, _ := orm.DB()
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	return orm
}

type gormLog struct {
	glog.Interface
	slowThreshold time.Duration
}

func (g *gormLog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	elapsed := time.Since(begin)
	if err != nil && err != gorm.ErrRecordNotFound {
		logging.FromContext(ctx).ErrorL("gorm.error", sql, err, logging.AttrOption{StartTime: &begin})
	} else if elapsed > g.slowThreshold {
		logging.FromContext(ctx).ErrorL("gorm.slow", sql, err, logging.AttrOption{StartTime: &begin})
	} else if viper.GetString("app.env") == def.EnvDevelopment { // 开发环境开启debug日志
		logging.FromContext(ctx).Warn("gorm.debug", sql, rows, logging.AttrOption{StartTime: &begin})
	}
}
