package main

//import "Driver-go/elevio"

//var queue = make([]elevio.ButtonEvent, 0)
var queue = make([]int, 0)

func addToQueue(f int) {
	queue = append(queue, f)
}