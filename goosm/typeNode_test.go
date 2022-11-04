package goosm

import "fmt"

func ExampleNode_Init() {
	point := Node{}
	point.Init(100, -48.4589921, -27.4275954, &map[string]string{"name": "Place Palace Hotel"})

	fmt.Printf("%v", point)

	// Output:
	// {"type":"FeatureCollection","features":[{"type":"Feature","id":"100","properties":{"id":"100","name":"Place Palace Hotel"},"geometry":{"type":"Point","coordinates":[-48.4589922,-27.4275954,0]}}]}
}

func ExampleNode_BoundingBox() {
	var err error
	var box Box

	point := Node{}
	point.Init(100, -48.4589921, -27.4275954, &map[string]string{"name": "Place Palace Hotel"})
	box, err = point.BoundingBox(100)
	if err != nil {
		fmt.Printf("test fail: %v", err)
	}

	fmt.Printf("%v", box.MakeGeoJSonFeature())

	// Output:
	// {"type":"Feature","id":"0","properties":{"id":"0"},"geometry":{"type":"Polygon","bbox":[-48.4582761,-27.4269598,-48.4582761,-27.4269598],"coordinates":[[[-48.4582761,-27.4269598,0],[-48.4597084,-27.4269598,0],[-48.4597084,-27.4282311,0],[-48.4582761,-27.4282311,0],[-48.4582761,-27.4269598,0],[-48.4582761,-27.4269598,0]]]}}
}

func ExampleNode_DirectionBetweenTwoPoints() {
	pointA := Node{}
	pointA.Init(2, -48.4515528, -27.4268720, &map[string]string{"name": "Rua José Mussi"})

	pointB := Node{}
	pointB.Init(1, -48.4583771, -27.4276728, &map[string]string{"name": "Avenida Da Nações"})

	fmt.Printf("%v", pointA.DirectionBetweenTwoPoints(pointB))

	// Output:
	// 262.4672688249775
}

func ExampleNode_DistanceBetweenTwoPoints() {
	pointA := Node{}
	pointA.Init(2, -48.4515528, -27.4268720, &map[string]string{"name": "Rua José Mussi"})

	pointB := Node{}
	pointB.Init(1, -48.4583771, -27.4276728, &map[string]string{"name": "Avenida Da Nações"})

	fmt.Printf("%v", pointA.DistanceBetweenTwoPoints(pointB))

	// Output:
	// 679.6735414854305
}

func ExampleNode_DestinationPoint() {
	var err error

	pointA := Node{}
	pointA.Init(2, -48.4515528, -27.4268720, &map[string]string{"name": "Rua José Mussi"})

	pointB := Node{}

	pointB, err = pointA.DestinationPoint(679.6735414854305, 262.4672688249775)
	if err != nil {
		fmt.Printf("test fail: %v", err)
	}

	fmt.Printf("%v", pointB)

	// Output:
	// {"type":"FeatureCollection","features":[{"type":"Feature","id":"0","properties":{"id":"0"},"geometry":{"type":"Point","coordinates":[-48.4583772,-27.4276729,0]}}]}
}

func ExampleNode_MakeGeoJSonFeature() {
	point := Node{}
	point.Init(100, -48.4589921, -27.4275954, &map[string]string{"name": "Place Palace Hotel"})

	fmt.Printf("%v", point.MakeGeoJSonFeature())

	// Output:
	// {"type":"FeatureCollection","features":[{"type":"Feature","id":"100","properties":{"id":"100","name":"Place Palace Hotel"},"geometry":{"type":"Point","coordinates":[-48.4589922,-27.4275954,0]}}]}
}
