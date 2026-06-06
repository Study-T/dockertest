package main

import (
	"flag"

	"ns-tracking-go/app/internal/config"
	routes "ns-tracking-go/app/internal/handler"
	"ns-tracking-go/app/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/app.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	server := rest.MustNewServer(c.RestConf)
	routes.RegisterHandlers(server, ctx)

	defer func() {
		server.Stop()
		ctx.Close()
	}()

	server.Start()
}
