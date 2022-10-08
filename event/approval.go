package event

type ApprovalEventContent struct {
	Reason   string           `json:"dealReason"`
	Type     string           `json:"schema_type"`
	Response ApprovalResponse `json:"org.matrix.msc3381.poll.response"`
}

type ApprovalResponse struct {
	Answers []string `json:"answers"`
}
