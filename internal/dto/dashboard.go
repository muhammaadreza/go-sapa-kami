package dto

// Dashboard Response
type Dashboard struct {
	KPIs        KPIs                 `json:"kpis"`
	Charts      Charts               `json:"charts"`
	Aspirations []AspirationResponse `json:"latest_aspirations"`
}

// KPIs Nested
type KPIs struct {
	Aspirations KPI `json:"aspirations"`
	Innovations KPI `json:"innovations"`
	Critisms    KPI `json:"criticisms"`
	ActiveUsers KPI `json:"active_users"`
}

// KPI Nested
type KPI struct {
	Total      int64   `json:"total"`
	Open       int64   `json:"open"`
	Percentage int64   `json:"percentage"`
	Sparkline  []int64 `json:"sparkline"`
}

// Charts Nested
type Charts struct {
	Weekly  []Chart `json:"weekly"`
	Monthly []Chart `json:"monthly"`
}

// Chart Nested
type Chart struct {
	Date        string `json:"date" format:"2006-01-02"`
	Innovations int64  `json:"innovations"`
	Criticisms  int64  `json:"criticisms"`
}
