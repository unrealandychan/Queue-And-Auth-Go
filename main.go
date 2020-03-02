package main

import (
		"./router"
			"github.com/gin-gonic/gin"

	"log"
	"github.com/apex/gateway"
)

func routerEngine() *gin.Engine{
	h := router.Handler{}
	r := gin.Default()
	r.POST("/event",h.NewEventHTTP)
	r.POST("/token" , h.IssueTokenHTTP)
	r.POST("/check",h.ValidateTokenAndCheckQueueHTTP)
	r.Run()
	return r
}


func main() {
	addr := ":"+"3000"
	log.Fatal(gateway.ListenAndServe(addr,routerEngine()))

	}
