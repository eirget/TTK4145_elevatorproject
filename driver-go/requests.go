package main

import (
	"Driver-go/elevio"
	"fmt"
)

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

