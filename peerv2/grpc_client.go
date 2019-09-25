package peerv2

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	p2pgrpc "github.com/paralin/go-libp2p-grpc"
	"google.golang.org/grpc"
	"log"
)

type GRPCService_Client struct {
	p2pgrpc *p2pgrpc.GRPCProtocol
}

func (self *GRPCService_Client) ProxyRegister(ctx context.Context, peerID peer.ID, pubkey string) (string, error) {
	grpcConn, err := self.p2pgrpc.Dial(ctx, peerID, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return "", err
	}
	client := NewProxyRegisterServiceClient(grpcConn)
	reply, err := client.ProxyRegister(ctx, &ProxyRegisterMsg{Pubkey: "11234567890"})
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	return reply.Result, nil
}
