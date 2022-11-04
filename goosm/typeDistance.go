package goosm

import "math"

// Distance
//
// English:
//
// Português:
type Distance struct {
	Meters float64 // distance
}

type DistanceList struct {
	List []Distance
}

// GetMeters
//
// English:
//
// # Returns the distance in meters
//
// Português:
//
// Retorna a distância em metros
func (d *Distance) GetMeters() float64 {
	return d.Meters
}

// AddMeters
//
// English:
//
// # Adds a new value to the current distance, in meters
//
// Português:
//
// Adiciona um novo valor a distância atual, em metros
func (d *Distance) AddMeters(m float64) {
	d.Meters += m
}

// SetMeters
//
// English:
//
// # Determine the distance in meters
//
// Português:
//
// Determina a distância em metros
func (d *Distance) SetMeters(m float64) {
	d.Meters = m
}

// SetMetersIfGreaterThan
//
// English:
//
// # Determines a new distance, if it is greater than the current distance, in meters
//
// Português:
//
// Determina uma nova distância, se ela for maior do que a distância atual, em metros
func (d *Distance) SetMetersIfGreaterThan(m float64) {
	test := math.Max(d.Meters, m)
	d.Meters = test
}
