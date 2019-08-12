# Wire  - list transfered message objects of P2P in incognito chain
This package contains list of messages, which are used to transfer from peer to peer.

Each of message when send from peer to peer, 1st 24 bytes is header of message(with 1st 12 bytes is command type of message). That mean when creaste a message to send, we need add 24 bytes as header of message before send to other peers.

Every message have a max length to transfer. If peer receive a message which has length > max lenght of current version message, it should be rejected by peer inMessageHandler