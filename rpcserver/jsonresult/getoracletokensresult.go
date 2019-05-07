package jsonresult

type GetOracleTokensResult struct {
	OracleTokens []*OracleToken
}

type OracleToken struct {
	TokenID   string `json:"TokenID"`
	TokenName string `json:"TokenName"`
}
