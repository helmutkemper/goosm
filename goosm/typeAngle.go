package goosm

import (
	"goosm/module/util"
)

// Angle
//
// English:
//
// # Object responsible for the angle treatment
//
// Português:
//
// Objeto responsável pelo tratamento de ângulo
type Angle struct {
	Degrees float64 //Angle
	Radians float64 //Angle
}

type AngleListStt struct {
	List []Angle
}

// SetDecimalDegrees
//
// English:
//
// # Set Angle value as decimal degrees
//
// Português:
//
// Defina o valor do ângulo como graus decimais
func (a *Angle) SetDecimalDegrees(degrees, primes, seconds float64) {
	a.Degrees = degrees + primes/60.0 + seconds/3600.0
	a.Radians = util.DegreesToRadians(degrees + primes/60.0 + seconds/3600.0)
}

// SetDegrees
//
// English:
//
// # Set Angle value as degrees
//
// Português:
//
// Determina o ângulo de graus
func (a *Angle) SetDegrees(angle float64) {
	a.Degrees = angle
	a.Radians = util.DegreesToRadians(angle)
}

// SetRadians
//
// English:
//
// # Set Angle value as radians
//
// Português:
//
// Determina o ângulo em radianos
func (a *Angle) SetRadians(angle float64) {
	a.Radians = angle
	a.Degrees = util.RadiansToDegrees(angle)
}

// AddDegrees
//
// English:
//
// # Adds the current angle to the given value, in degrees
//
// Português:
//
// Soma o ângulo atual com o valor determinado, em graus
func (a *Angle) AddDegrees(angle float64) {
	a.Degrees = a.Degrees + angle
	a.Radians = util.DegreesToRadians(a.Degrees)
}

// GetAsRadians
//
// English:
//
// # Returns the angle in radians
//
// Português:
//
// Retorna o ângulo em radianos
func (a *Angle) GetAsRadians() float64 {
	return a.Radians
}

// GetAsDegrees
//
// English:
//
// # Returns the angle in degrees
//
// Português:
//
// Retorna o ângulo em graus
func (a *Angle) GetAsDegrees() float64 {
	return a.Degrees
}
