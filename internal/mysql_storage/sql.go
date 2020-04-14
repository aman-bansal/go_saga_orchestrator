package mysql_storage

import (
	"database/sql"
	"fmt"
	"time"
)

type MysqlConfig struct {
	Host   string
	DbName string
	User   string
	Pass   string
}
type StorageClient interface {
}

type MysqlClient struct {
	db *sql.DB
}

const DefaultMaxOpenConn = 10
const DefaultMaxIdleConn = 10
const DefaultConnMaxLifeTime = time.Hour

func NewStorageClient(conf MysqlConfig) (StorageClient, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True", conf.User, conf.Pass, conf.Host, conf.DbName))
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(DefaultConnMaxLifeTime)
	db.SetMaxIdleConns(DefaultMaxIdleConn)
	db.SetMaxOpenConns(DefaultMaxOpenConn)

	return &MysqlClient{
		db: db,
	}, nil
}
