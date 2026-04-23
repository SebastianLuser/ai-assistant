package domain

type PreviewResult struct {
	Tool        string         `json:"tool"`
	WouldDo     string         `json:"would_do"`
	Params      map[string]any `json:"params"`
	Reversible  bool           `json:"reversible"`
	ConfirmHint string         `json:"confirm_hint"`
}
