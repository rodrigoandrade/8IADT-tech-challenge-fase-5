package domain

type AIResponse struct {
	Message            string                 `json:"message"`
	Intent             string                 `json:"intent,omitempty"`
	LeadStatus         string                 `json:"lead_status,omitempty"`
	QualificationScore *float64               `json:"qualification_score,omitempty"`
	NextAction         string                 `json:"next_action,omitempty"`
	Reasoning          string                 `json:"reasoning,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	TransactionType    string                 `json:"transaction_type,omitempty"`
	Qualification      map[string]interface{} `json:"qualification,omitempty"`
}
