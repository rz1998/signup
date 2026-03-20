package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"signup/config"
	"signup/handler"
	"signup/svc"
)

var configFile = flag.String("f", "etc/api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// Create service context
	serverCtx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatalf("Failed to create service context: %v", err)
	}
	defer serverCtx.Close()

	// Create HTTP server
	server := rest.MustNewServer(c.RestConf,
		rest.WithCors("*"),
	)
	defer server.Stop()

	// Register handlers
	handler.RegisterHandlers(server, serverCtx)

	// Start server
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
