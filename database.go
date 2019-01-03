package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

type Field struct {
	Name       string `yaml:"field"`
	ForeignKey string `yaml:"foreign_key,omitempty"`
}

type Table struct {
	Name     string    `yaml:"table"`
	Fields   []Field   `yaml:"fields"`
	database *DataBase `yaml:"omitempty"`
}

func (t Table) Database() *DataBase {
	return t.database
}

type DataBase struct {
	Name     string   `yaml:"database"`
	Type     string   `yaml:"type"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Host     string   `yaml:"host"`
	Port     string   `yaml:"port"`
	Tables   []Table  `yaml:"tables,omitempty"`
	dbConn   *gorm.DB `yaml:"omitempty"`
}

func (d DataBase) DBConn() *gorm.DB {
	return d.dbConn
}

type Config struct {
	Databases []DataBase `yaml:"databases"`
}

var (
	Conf            Config
	Tables          map[string]*Table
	DefaultDatabase *DataBase
	RootIsOnline    bool
	Mutex           sync.Mutex
)

// 读取配置文件，初始化服务
func ConfigInit(configFilePath string) {
	b, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal("读取配置文件错误")
	}

	if yaml.Unmarshal(b, &Conf) != nil {
		log.Fatal("配置文件有误")
	}

	if len(Conf.Databases) == 0 {
		log.Fatal("配置文件中没有指定数据库")
	}

	Tables = make(map[string]*Table)
	DefaultDatabase = &Conf.Databases[0]

	for i := range Conf.Databases {
		Conf.Databases[i].dbConn = getDBConn(Conf.Databases[i].Type, &Conf.Databases[i])
	}

	for i := range Conf.Databases {
		for j := range Conf.Databases[i].Tables {
			t := &Conf.Databases[i].Tables[j]
			t.database = &Conf.Databases[i]
			Tables[t.Name] = t
		}
	}
	log.Println("配置文件读取成功，初始化完成")
}

// 将服务当前状态更新到配置文件中，因为可能有新表未在配置文件中
func SaveConfig() error {
	b, err := yaml.Marshal(&Conf)
	if err != nil {
		log.Println("配置保存失败")
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("backup/%d.config.yaml", time.Now().Unix()), b, 0644)
	if err != nil {
		log.Println("写备份配置文件失败")
		return err
	}

	err = ioutil.WriteFile("config.yaml", b, 0644)
	if err != nil {
		log.Println("写配置文件失败，下次服务启动请注意检查配置文件与备份文件的一致性")
	}

	return nil
}

func getDBConn(databaseType string, database *DataBase) *gorm.DB {
	switch databaseType {
	case "mysql":
		return connectMySQL(database)
	case "postgres":
		return connectPostgres(database)
	default:
		log.Fatal("不支持该数据库")
	}
	return nil
}

func connectMySQL(d *DataBase) *gorm.DB {
	db, err := gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", d.Username, d.Password, d.Host, d.Port, d.Name))
	if err != nil {
		log.Fatal("连接mysql失败")
	}
	db.LogMode(true)
	log.Printf("成功连接mysql数据库%s\n", d.Name)

	return db
}

func connectPostgres(d *DataBase) *gorm.DB {
	db, err := gorm.Open("postgres", fmt.Sprintf(
		"host=%s user=%s dbname=%s sslmode=disable password=%s", d.Host, d.Username, d.Name, d.Password))
	if err != nil {
		log.Fatal("连接postgres失败")
	}
	db.LogMode(true)
	log.Printf("成功连接postgres数据库%s\n", d.Name)

	return db
}
