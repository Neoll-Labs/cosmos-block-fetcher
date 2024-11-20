package types

type BlockMetadata struct {
	Height uint64 `json:"height"`
	NumTxs int    `json:"num_txs"`
}

type BlockResponse struct {
	Result struct {
		Block struct {
			Header struct {
				Height  string `json:"height"`
				ChainID string `json:"chain_id"`
				Time    string `json:"time"`
			} `json:"header"`
			Data struct {
				Txs []interface{} `json:"txs"`
			} `json:"data"`
		} `json:"block"`
	} `json:"result"`
}

type StatusResponse struct {
	Result struct {
		NodeInfo struct {
			Network string `json:"network"`
		} `json:"node_info"`
	} `json:"result"`
}

type Output struct {
	ChainID string          `json:"chain_id"`
	Blocks  []BlockMetadata `json:"blocks"`
}
