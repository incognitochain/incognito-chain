package rpccaller

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCBaseRes struct {
	Id       int       `json:"id"`
	RPCError *RPCError `json:"error"`
}

