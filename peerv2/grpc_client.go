package peerv2

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p-core/peer"
	p2pgrpc "github.com/paralin/go-libp2p-grpc"
	"google.golang.org/grpc"
)

type GRPCService_Client struct {
	p2pgrpc *p2pgrpc.GRPCProtocol
}

func (self *GRPCService_Client) ProxyRegister(
	ctx context.Context,
	peerID peer.ID,
	pubkey string,
	messages []string,
) ([]string, error) {
	grpcConn, err := self.p2pgrpc.Dial(ctx, peerID, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	client := NewProxyRegisterServiceClient(grpcConn)
	reply, err := client.ProxyRegister(ctx, &ProxyRegisterMsg{CommitteePublicKey: pubkey, WantedMessages: messages})
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return reply.Topics, nil
}
