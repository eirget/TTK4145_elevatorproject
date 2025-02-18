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

func (e *Elevator) dealWithNewReq(newFloor int) {
	if len(e.Queue) == 0 {
		e.Queue = append(e.Queue, newFloor)
		return
	}

	// Determine best place to insert
	var newQueue []int
	inserted := false

	switch e.Direction {
	case elevio.MD_Up:
		for _, floor := range e.Queue {
			// Insert in ascending order while maintaining direction
			if newFloor < floor && !inserted {
				newQueue = append(newQueue, newFloor)
				inserted = true
			}
			newQueue = append(newQueue, floor)
		}
	case elevio.MD_Down:
		for _, floor := range e.Queue {
			// Insert in descending order while maintaining direction
			if newFloor > floor && !inserted {
				newQueue = append(newQueue, newFloor)
				inserted = true
			}
			newQueue = append(newQueue, floor)
		}
	default:
		// Elevator is idle, just append in order
		newQueue = append(e.Queue, newFloor)
		inserted = true
	}

	// If not inserted, put it at the end
	if !inserted {
		newQueue = append(newQueue, newFloor)
	}

	e.Queue = newQueue
}
