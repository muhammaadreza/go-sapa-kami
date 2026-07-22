package model

type Dashboard struct {
	KPIs     KPIs              `json:"kpis"`
	Charts   Charts            `json:"charts"`
	Messages []MessageResponse `json:"latest_messages"`
}

type KPIs struct {
	Messages    KPI `json:"messages"`
	Innovations KPI `json:"innovations"`
	Critisms    KPI `json:"criticisms"`
	ActiveUsers KPI `json:"active_users"`
}

type KPI struct {
	Total 	   int64	`json:"total"`
	Percentage int64    `json:"percentage"`
	Sparkline  []int64  `json:"sparkline"`
}

type Charts struct {
	Weekly  []Chart `json:"weekly"`
	Monthly []Chart `json:"monthly"`
}

type Chart struct {
	Date 		string `json:"date" format:"2006-01-02"`
	Innovations int64  `json:"innovations"`
	Criticisms  int64  `json:"criticisms"`
}
