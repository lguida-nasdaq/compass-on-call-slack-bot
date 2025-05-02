package api

type Schedule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Timezone    string `json:"timezone"`
	Enabled     bool   `json:"enabled"`
	TeamID      string `json:"teamId"`
}

type OnCallParticipant struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type OnCallResponse struct {
	OnCallParticipants []OnCallParticipant `json:"onCallParticipants"`
}
