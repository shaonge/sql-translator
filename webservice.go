package main

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"io"
	"log"
	"net/http"
	"strconv"
)

func StartWebService() error {
	if StartGetService() != nil {
		log.Fatal("get服务启动失败！")
		return fmt.Errorf("get服务启动失败！")
	}

	if http.ListenAndServe(":8080", nil) != nil {
		log.Fatal("服务启动失败！")
		return fmt.Errorf("服务启动失败！")
	}

	return nil
}

func get(request *restful.Request, response *restful.Response) {
	switch request.PathParameter("entity") {
	case "user":
	default:
		response.WriteErrorString(404, "")
		return
	}

	id, err := strconv.Atoi(request.PathParameter("id"))
	if err != nil {
		response.WriteErrorString(404, "")
		return
	}

	user, err := GetUser(int32(id))
	io.WriteString(response, user.String())
}

func StartGetService() error {
	ws := new(restful.WebService)
	ws.Route(ws.GET("/{entity}/{id}").To(get))
	restful.Add(ws)
	return nil
}
