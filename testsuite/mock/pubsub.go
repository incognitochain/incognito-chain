package mock

import "github.com/incognitochain/incognito-chain/pubsub"

type Pubsub struct{}

func (ps *Pubsub) PublishMessage(message *pubsub.Message) {
	return
}
