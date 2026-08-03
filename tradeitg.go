package main

import (
	"flag"
	"fmt"

	"github.com/starslipay/trade_itg/internal/config"
	"github.com/starslipay/trade_itg/internal/metrics"
	"github.com/starslipay/trade_itg/internal/server"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/trade_itg_pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/tradeitg.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		trade_itg_pb.RegisterTradeItgServer(grpcServer, server.NewTradeItgServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(metrics.UnaryMetricInterceptor)
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
