package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type JsonRequest struct {
	Cmd     string              `json:"cmd"`
	Table   string              `json:"table,omitempty"`
	Data    []map[string]string `json:"data,omitempty"`
	request *restful.Request
}

type JsonResponse struct {
	Data     []map[string]string `json:"data,omitempty"`
	response *restful.Response
}

func StartWebService() {
	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(welcome))
	ws.Route(ws.POST("/{entity}").To(post))
	restful.Add(ws)

	if http.ListenAndServe(":8080", nil) != nil {
		log.Fatal("web服务启动失败")
	}
}

func printlnRequest(req *JsonRequest) {
	log.Printf("url: %s   cmd: %s\n", req.request.Request.URL, req.Cmd)
}

func welcome(request *restful.Request, response *restful.Response) {
	_, err := response.Write([]byte("Here is DataHelper, welcome!"))
	if err != nil {
		_ = response.WriteErrorString(603, "server error")
		return
	}
}

func post(request *restful.Request, response *restful.Response) {
	log.Println("新请求到达")
	req := &JsonRequest{request: request}
	res := &JsonResponse{response: response}
	res.response.AddHeader("Access-Control-Allow-Origin", "*")
	res.response.AddHeader("Access-Control-Allow-Methods", "POST")
	res.response.AddHeader("Access-Control-Allow-Headers", "x-requested-with,content-type")
	res.response.Header().Set("Content-Type", "application/json")
	doRequest(req, res)
	writeResponse(req, res)
}

func writeError(res *JsonResponse, statusCode int, errorMsg string) {
	res.response.Header().Set("Content-Type", errorMsg)
	_ = res.response.WriteErrorString(statusCode, "")
	log.Println("存在错误，请求处理失败")
}

func writeResponse(req *JsonRequest, res *JsonResponse) {
	log.Println("准备写入response")

	body, err := json.Marshal(res)
	if err != nil {
		writeError(res, 605, fmt.Sprintf("server error: %s", err))
		return
	}

	_, err = res.response.Write(body)
	if err != nil {
		writeError(res, 605, fmt.Sprintf("server error: %s", err))
		return
	}

	log.Println("response写入成功")
}

func checkTableExist(table string) bool {
	if _, ok := Tables[table]; !ok {
		return false
	}
	return true
}

func doRequest(req *JsonRequest, res *JsonResponse) {
	log.Println("准备解析请求数据")
	body, err := ioutil.ReadAll(req.request.Request.Body)
	if err != nil {
		writeError(res, 603, fmt.Sprintf("server error: %s", err))
		return
	}
	err = json.Unmarshal(body, req)
	if err != nil {
		writeError(res, 603, fmt.Sprintf("request data error: %s", err))
		return
	}
	log.Println("解析请求数据成功")
	printlnRequest(req)
	req.Table = req.request.PathParameter("entity")
	switch req.Cmd {
	case "authenticate":
		doAuthenticate(req, res)
	case "show":
		doShow(req, res)
	case "create":
		doCreate(req, res)
	case "drop":
		doDrop(req, res)
	case "describe":
		doDescribe(req, res)
	case "insert":
		doInsert(req, res)
	case "select":
		doSelect(req, res)
	case "update":
		doUpdate(req, res)
	case "delete":
		doDelete(req, res)
	default:
		writeError(res, 601, "invalid command")
		return
	}
}

func doShow(req *JsonRequest, res *JsonResponse) {
	switch req.request.PathParameter("entity") {
	case "tables":
		allTables(req, res)
	case "data-types":
		allDataTypes(req, res)
	default:
		writeError(res, 608, "url error: unrecognized url")
		return
	}
}

func allTables(req *JsonRequest, res *JsonResponse) {
	//if len(req.Data) > 0 {
	//	table := req.Data[0]["table"]
	//	if !checkTableExist(table) {
	//		writeError(res, 607, fmt.Sprintf("table %s is not existed", table))
	//		return
	//	}
	//	database := Tables[table].Database().Name
	//	for k, v := range Tables {
	//		if v.database.Name == database {
	//			table := make(map[string]string)
	//			table["name"] = k
	//			res.Data = append(res.Data, table)
	//		}
	//	}
	//} else {
	for k := range Tables {
		table := make(map[string]string)
		table["name"] = k
		res.Data = append(res.Data, table)
	}
	//}
}

func allDataTypes(req *JsonRequest, res *JsonResponse) {
	dataTypes := []string{"INT", "VARCHAR(255)"}
	for i := range dataTypes {
		res.Data = append(res.Data, map[string]string{"type": dataTypes[i]})
	}
}

func doAuthenticate(req *JsonRequest, res *JsonResponse) {
	if req.request.PathParameter("entity") != "authentication" {
		writeError(res, 608, "url error: url should be /authentication")
		return
	}
	//Mutex.Lock()
	if !(!RootIsOnline && len(req.Data) != 0 && req.Data[0]["username"] == "root" && req.Data[0]["password"] == "123456") {
		writeError(res, 606, "authenticate failed")
		return
	}
	//RootIsOnline = true
	//Mutex.Unlock()
}

func doCreate(req *JsonRequest, res *JsonResponse) {
	if checkTableExist(req.Table) {
		writeError(res, 607, fmt.Sprintf("table %s is existed", req.Table))
		return
	}
	if len(req.Data) == 0 {
		writeError(res, 603, "request data error: no field")
		return
	}

	d := DefaultDatabase

	s2 := ""
	if len(req.Data) > 1 {
		for f, t := range req.Data[1] {
			if t != "" {
				s2 += fmt.Sprintf("foreign key (%s) references %s(id) on delete set null on update cascade,", f, t)
				d = Tables[t].Database()
			}
		}
	}

	s1 := fmt.Sprintf("create table %s(", req.Table)
	switch d.Type {
	case "mysql":
		s1 += "id int primary key auto_increment,"
	case "postgres":
		s1 += "id serial primary key,"
	default:
		panic("哈哈哈哈哈哈哈")
	}

	for f, t := range req.Data[0] {
		if f != "id" {
			s1 += fmt.Sprintf("%s %s,", f, t)
		}
	}

	statement := s1 + s2
	statement = strings.Trim(statement, ",")
	statement += ")"
	err := d.DBConn().Exec(statement).Error
	if err != nil {
		writeError(res, 604, fmt.Sprintf("access database failed: %s", err))
		return
	}

	var table Table
	table.Name = req.Table
	table.database = d
	table.Fields = make([]Field, 0)

	for k := range req.Data[0] {
		f := Field{k, req.Data[1][k]}
		table.Fields = append(table.Fields, f)
	}

	Tables[req.Table] = &table
	d.Tables = append(d.Tables, table)

	err = SaveConfig()
	if err != nil {
		writeError(res, 605, fmt.Sprintf("server error: %s", err))
		return
	}
}

func doDrop(req *JsonRequest, res *JsonResponse) {
	if !checkTableExist(req.Table) {
		writeError(res, 607, fmt.Sprintf("table %s is not existed", req.Table))
		return
	}
	DB := Tables[req.Table].Database().DBConn()
	err := DB.Exec(fmt.Sprintf("drop table %s", req.Table)).Error
	if err != nil {
		writeError(res, 604, fmt.Sprintf("access database failed: %s", err))
		return
	}

	m := make(map[string]*Table)
	tables := make([]Table, 0)
	d := Tables[req.Table].Database()
	for i := range d.Tables {
		if d.Tables[i].Name != req.Table {
			m[d.Tables[i].Name] = &d.Tables[i]
		}
	}
	for _, v := range m {
		tables = append(tables, *v)
	}
	d.Tables = tables

	delete(Tables, req.Table)

	err = SaveConfig()
	if err != nil {
		writeError(res, 605, fmt.Sprintf("server error: %s", err))
		return
	}
}

func doDescribe(req *JsonRequest, res *JsonResponse) {
	if !checkTableExist(req.Table) {
		writeError(res, 607, fmt.Sprintf("table %s is not existed", req.Table))
		return
	}
	table := Tables[req.Table]
	for _, f := range table.Fields {
		field := make(map[string]string)
		field["field"] = f.Name
		field["foreign_key"] = f.ForeignKey
		res.Data = append(res.Data, field)
	}
}

func doInsert(req *JsonRequest, res *JsonResponse) {
	if !checkTableExist(req.Table) {
		writeError(res, 607, fmt.Sprintf("table %s is not existed", req.Table))
		return
	}
	DB := Tables[req.Table].Database().DBConn()
	DB = DB.Begin()

	for _, m := range req.Data {
		if m["id"] != "" {
			writeError(res, 603, "request data error: id should be null")
			DB.Rollback()
			return
		}
		delete(m, "id")
		fields := make([]string, 0, len(m))
		values := make([]string, 0, len(m))
		for k, v := range m {
			fields = append(fields, k)
			values = append(values, v)
		}
		fs := ""
		for _, f := range fields {
			fs += fmt.Sprintf(",%s", f)
		}
		fs = strings.Trim(fs, ",")

		vs := ""
		for _, v := range values {
			vs += fmt.Sprintf(",'%s'", v)
		}
		vs = strings.Trim(vs, ",")

		statement := fmt.Sprintf("insert into %s(%s) values(%s)", req.Table, fs, vs)
		err := DB.Exec(statement).Error
		if err != nil {
			writeError(res, 604, fmt.Sprintf("access database failed: %s", err))
			DB.Rollback()
			return
		}

		var row *sql.Row
		switch Tables[req.Table].Database().Type {
		case "mysql":
			row = DB.Table(req.Table).Select("last_insert_id()").Row()
		case "postgres":
			row = DB.Table(req.Table).Select("max(id)").Row()
		default:
			panic("哈哈哈哈哈哈哈")
		}

		var id string
		err = row.Scan(&id)
		if err != nil {
			writeError(res, 604, fmt.Sprintf("server error: %s", err))
			DB.Rollback()
			return
		}
		m["id"] = id
	}

	res.Data = req.Data
	DB.Commit()
}

func doSelect(req *JsonRequest, res *JsonResponse) {
	if !checkTableExist(req.Table) {
		writeError(res, 607, fmt.Sprintf("table %s is not existed", req.Table))
		return
	}
	DB := Tables[req.Table].Database().DBConn()

	fields := make([]string, 0)
	for k := range req.Data[0] {
		fields = append(fields, k)
	}

	var rows *sql.Rows
	//defer rows.Close()
	var err error

	if req.Data[0]["id"] == "" {
		rows, err = DB.Table(req.Table).Select(fields).Rows()
	} else {
		rows, err = DB.Table(req.Table).Where("id = ?", req.Data[0]["id"]).Select(fields).Rows()
	}

	if err != nil {
		writeError(res, 604, fmt.Sprintf("access database failed: %s", err))
		return
	}

	values := make([]sql.NullString, len(fields))
	p := make([]interface{}, len(fields))
	for i := range values {
		p[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(p...)
		if err != nil {
			writeError(res, 605, fmt.Sprintf("server error: %s", err))
			return
		}

		obj := make(map[string]string)

		for i, f := range fields {
			obj[f] = func(s sql.NullString) string {
				if s.Valid {
					return s.String
				}
				return "-1"
			}(values[i])
		}

		res.Data = append(res.Data, obj)
	}
}

func doUpdate(req *JsonRequest, res *JsonResponse) {
	if !checkTableExist(req.Table) {
		writeError(res, 607, fmt.Sprintf("table %s is not existed", req.Table))
		return
	}
	DB := Tables[req.Table].Database().DBConn()
	DB = DB.Begin()

	for _, m := range req.Data {
		if m["id"] == "" {
			writeError(res, 603, "request data error: id can't be null")
			DB.Rollback()
			return
		} else {
			id := m["id"]
			err := DB.Table(req.Table).Where("id = ?", id).Update(m).Error
			if err != nil {
				writeError(res, 604, fmt.Sprintf("access database failed: %s", err))
				DB.Rollback()
				return
			}
		}
	}

	res.Data = req.Data

	DB.Commit()
}

func doDelete(req *JsonRequest, res *JsonResponse) {
	if !checkTableExist(req.Table) {
		writeError(res, 607, fmt.Sprintf("table %s is not existed", req.Table))
		return
	}
	DB := Tables[req.Table].Database().DBConn()
	DB = DB.Begin()

	for _, m := range req.Data {
		if m["id"] == "" {
			writeError(res, 603, "request data error: id can't be null")
			DB.Rollback()
			return
		} else {
			id := m["id"]
			delete(m, "id")
			err := DB.Table(req.Table).Where("id = ?", id).Delete(nil).Error
			if err != nil {
				writeError(res, 604, fmt.Sprintf("access database failed: %s", err))
				DB.Rollback()
				return
			}
		}
	}
	DB.Commit()
}
