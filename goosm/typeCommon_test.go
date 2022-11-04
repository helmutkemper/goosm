package goosm

import (
	"fmt"
)

func ExampleCommon_boundingBox() {
	var box Box
	var common = Common{}

	box = common.boundingBox([2]float64{-48.4589921, -27.4275954}, 100)
	fmt.Printf("%v", box.MakeGeoJSonFeature())

	// Output:
	// {"type":"Feature","id":"0","properties":{"id":"0"},"geometry":{"type":"Polygon","bbox":[-48.458276,-27.4269598,-48.458276,-27.4269598],"coordinates":[[[-48.458276,-27.4269598,0],[-48.4597083,-27.4269598,0],[-48.4597083,-27.4282311,0],[-48.458276,-27.4282311,0],[-48.458276,-27.4269598,0],[-48.458276,-27.4269598,0]]]}}
}

func ExampleCommon_directionBetweenTwoPoints() {
	var common = Common{}
	fmt.Printf("%v", common.directionBetweenTwoPoints([2]float64{-48.4515528, -27.4268720}, [2]float64{-48.4583771, -27.4276728}))

	// Output:
	// 262.4672688249775
}

func ExampleCommon_distanceBetweenTwoPoints() {
	var common = Common{}
	fmt.Printf("%v", common.distanceBetweenTwoPoints([2]float64{-48.4515528, -27.4268720}, [2]float64{-48.4583771, -27.4276728}))

	// Output:
	// 679.6735414854305
}

func ExampleCommon_destinationPoint() {
	var common = Common{}
	var pointA = Node{}
	var coordinate [2]float64

	coordinate = common.destinationPoint([2]float64{-48.4515528, -27.4268720}, 679.6735414854305, 262.4672688249775)
	pointA.Init(0, coordinate[0], coordinate[1], nil)
	fmt.Printf("%v", pointA)

	// Output:
	// {"type":"FeatureCollection","features":[{"type":"Feature","id":"0","properties":{"id":"0"},"geometry":{"type":"Point","coordinates":[-48.4583772,-27.4276729,0]}}]}
}

//func ExampleCommon_MakeGeoJSonFeature() {
//	point := Node{}
//	point.Init(100, -48.4589921, -27.4275954, &map[string]string{"name": "Place Palace Hotel"})
//
//	fmt.Printf("%v", point.MakeGeoJSonFeature())
//
//	// Output:
//	// {"type":"FeatureCollection","features":[{"type":"Feature","id":"100","properties":{"id":"100","name":"Place Palace Hotel"},"geometry":{"type":"Point","coordinates":[-48.4589922,-27.4275954,0]}}]}
//}
