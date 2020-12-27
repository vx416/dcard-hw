package restful

import "github.com/vx416/dcard-work/pkg/server"

func V1Routes(serv *server.Server, handler *Handler) {
	serv.GET("/", handler.ReqStatsEndpoint)
	serv.GET("/guardian_animal", handler.GetGuardianAnimalEndpoint)
}
