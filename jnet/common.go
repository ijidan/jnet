package jnet

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"sync"
)

var connMap = map[string]*grpc.ClientConn{}
var lock sync.Mutex

//构建链接
func BuildConnection(address string) *grpc.ClientConn {
	lock.Lock()
	defer lock.Unlock()
	conn := connMap[address]
	if conn == nil {
		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			grpclog.Fatalln("gRPC connection error：" + err.Error())
		}
		connMap[address] = conn
	}
	return connMap[address]
}
