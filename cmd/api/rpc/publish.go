/*
 * @Author: a76yyyy q981331502@163.com
 * @Date: 2022-06-12 22:40:31
 * @LastEditors: a76yyyy q981331502@163.com
 * @LastEditTime: 2022-06-18 23:44:05
 * @FilePath: /tiktok/cmd/api/rpc/publish.go
 * @Description: Publish RPC 客户端初始化及 相关 RPC 通信操作定义
 */

package rpc

import (
	"context"
	"fmt"
	"time"

	etcd "github.com/a76yyyy/registry-etcd"
	"github.com/a76yyyy/tiktok/kitex_gen/publish"
	"github.com/a76yyyy/tiktok/kitex_gen/publish/publishsrv"
	"github.com/a76yyyy/tiktok/pkg/errno"
	"github.com/a76yyyy/tiktok/pkg/middleware"
	"github.com/a76yyyy/tiktok/pkg/ttviper"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/retry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

var publishClient publishsrv.Client

// Publish RPC 客户端初始化
func initPublishRpc(Config *ttviper.Config) {
	EtcdAddress := fmt.Sprintf("%s:%d", Config.Viper.GetString("Etcd.Address"), Config.Viper.GetInt("Etcd.Port"))
	r, err := etcd.NewEtcdResolver([]string{EtcdAddress})
	if err != nil {
		panic(err)
	}
	ServiceName := Config.Viper.GetString("Server.Name")

	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(ServiceName),
		provider.WithExportEndpoint("localhost:4317"),
		provider.WithInsecure(),
	)
	defer p.Shutdown(context.Background())

	c, err := publishsrv.NewClient(
		ServiceName,
		client.WithMiddleware(middleware.CommonMiddleware),
		client.WithInstanceMW(middleware.ClientMiddleware),
		client.WithMuxConnection(1),                       // mux
		client.WithRPCTimeout(30*time.Second),             // rpc timeout
		client.WithConnectTimeout(30000*time.Millisecond), // conn timeout
		client.WithFailureRetry(retry.NewFailurePolicy()), // retry
		client.WithSuite(tracing.NewClientSuite()),        // tracer
		client.WithResolver(r),                            // resolver
		// Please keep the same as provider.WithServiceName
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: ServiceName}),
	)
	if err != nil {
		panic(err)
	}
	publishClient = c
}

// 传递 发布视频操作 的上下文, 并获取 RPC Server 端的响应.
func PublishAction(ctx context.Context, req *publish.DouyinPublishActionRequest) (resp *publish.DouyinPublishActionResponse, err error) {
	resp, err = publishClient.PublishAction(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 0 {
		return nil, errno.NewErrNo(int(resp.StatusCode), *resp.StatusMsg)
	}
	return resp, nil
}

// 传递 获取用户发布视频列表操作 的上下文, 并获取 RPC Server 端的响应.
func PublishList(ctx context.Context, req *publish.DouyinPublishListRequest) (resp *publish.DouyinPublishListResponse, err error) {
	resp, err = publishClient.PublishList(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 0 {
		return nil, errno.NewErrNo(int(resp.StatusCode), *resp.StatusMsg)
	}
	return resp, nil
}
