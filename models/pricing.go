package models

type PriceRequest struct {
	DistanceKm      float64 `json:"distance_km"`
	DurationMinutes float64 `json:"duration_minutes"`
	VehicleType     string  `json:"vehicle_type"`
}

type PriceResponse struct {
	VehicleType   string  `json:"vehicle_type"`
	BaseFare      float64 `json:"base_fare"`
	DistanceCost  float64 `json:"distance_cost"`
	DurationCost  float64 `json:"duration_cost"`
	TotalFare     float64 `json:"total_fare"`
	Currency      string  `json:"currency"`
}