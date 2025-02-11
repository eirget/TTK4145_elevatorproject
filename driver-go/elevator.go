package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

type Elevator struct {
	//mutex over states maybe to protect
	Floor_nr  int
	Direction int
	On_floor  bool
	Door_open bool
	Queue     []int
}

const (
	Up   = 1
	Down = -1
	Stop = 0
)

func (e *Elevator) AddToQueue(floor int) {
	// Check if the floor is already in the queue to avoid duplicates
	for _, f := range e.Queue {
		if f == floor {
			return
		}
	}

	e.Queue = append(e.Queue, floor)
	fmt.Println("Added floor:", floor, "Updated Queue:", e.Queue)
}

//struct med "events" også? slik at

func elevatorInit() *Elevator {
	return &Elevator{
		Floor_nr:  1,
		Direction: 0,
		On_floor:  false,
		Door_open: false,
		Queue:     []int{},
	}
}

// Moves the elevator towards its destination
func (e *Elevator) processQueue() {
	for {
		if len(e.Queue) == 0 {
			time.Sleep(500 * time.Millisecond) // Wait and check again
			continue
		}

		targetFloor := e.Queue[0] // Get first floor in queue

		if e.Floor_nr < targetFloor {
			fmt.Println("Moving Up...")
			elevio.SetMotorDirection(elevio.MD_Up)
		} else if e.Floor_nr > targetFloor {
			fmt.Println("Moving Down...")
			elevio.SetMotorDirection(elevio.MD_Down)
		}

		// Wait for floor update in the main FSM
		time.Sleep(100 * time.Millisecond)
	}
}
