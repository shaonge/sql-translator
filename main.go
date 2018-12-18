package main

import (
	"log"
)

func main() {
	if ConnectMySQL() != nil {
		log.Fatal("数据库连接错误！")
		return
	}
	DB.LogMode(true)
	defer DB.Close()

	DB.AutoMigrate(&User{})

	if StartWebService() != nil {
		log.Fatal("web服务启动失败！")
		return
	}
}
