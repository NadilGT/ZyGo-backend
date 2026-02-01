package models

type RouteResponse struct {
	Code      string     `json:"code"`
	Routes    []Route    `json:"routes"`
	Waypoints []Waypoint `json:"waypoints"`
}

type Route struct {
	Legs       []Leg     `json:"legs"`
	WeightName string    `json:"weight_name"`
	Geometry   Geometry  `json:"geometry"`
	Weight     float64   `json:"weight"`
	Duration   float64   `json:"duration"`
	Distance   float64   `json:"distance"`
}

type Leg struct {
	Steps    []Step  `json:"steps"`
	Weight   float64 `json:"weight"`
	Summary  string  `json:"summary"`
	Duration float64 `json:"duration"`
	Distance float64 `json:"distance"`
}

type Step struct {
	// Empty in this response, but can be extended if needed
}

type Geometry struct {
	Coordinates [][]float64 `json:"coordinates"`
	Type        string      `json:"type"`
}

type Waypoint struct {
	Hint     string    `json:"hint"`
	Location []float64 `json:"location"`
	Name     string    `json:"name"`
	Distance float64   `json:"distance"`
}
