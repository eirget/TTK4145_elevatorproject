package main

import (
	"Driver-go/elevio"
)

/*
func (e *Elevator) dealWithNewReq (floor int) {
	for _, f := range e.Queue {
		if f == floor {
			return
		}
	}

	e.Queue = append(e.Queue, floor)

	current_dir := e.Direction

	if current_dir == elevio.MD_Up {
		for element := range e.Queue {
			if element > e.Floor_nr {
				e.Queue[]
			}
		}
	}


	fmt.Println("Added floor:", floor, "Updated Queue:", e.Queue)
}
*/

func (e *Elevator) dealWithNewReq(newFloor int, newButton elevio.ButtonType) {

	newOrder := Order{newFloor, newButton}

	for _, f := range e.Queue {
		if f.Button_direction == newButton && f.Floor == newFloor {
			return
		}
	}

	if len(e.Queue) == 0 {
		e.Queue = append(e.Queue, newOrder)
		return
	}


	// Determine best place to insert
	var newQueue []Order
	inserted := false

	switch e.Direction {
	case elevio.MD_Up:
		for _, order := range e.Queue {
			// Insert in ascending order while maintaining direction
			if newFloor < order.Floor && !inserted {
				newQueue = append(newQueue, newOrder)
				inserted = true
			}
			newQueue = append(newQueue, newOrder)
		}
	case elevio.MD_Down:
		for _, order := range e.Queue {
			// Insert in descending order while maintaining direction
			if newFloor > order.Floor && !inserted {
				newQueue = append(newQueue, newOrder)
				inserted = true
			}
			newQueue = append(newQueue, newOrder)
		}
	default:
		// Elevator is idle, just append in order
		newQueue = append(e.Queue, newOrder)
		inserted = true
	}

	// If not inserted, put it at the end
	if !inserted {
		newQueue = append(newQueue, newOrder)
	}

	e.Queue = newQueue
}
