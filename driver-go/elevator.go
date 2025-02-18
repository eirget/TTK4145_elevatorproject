package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

type Elevator struct {
	//mutex over states maybe to protect
	Floor_nr    int
	Direction   elevio.MotorDirection
	On_floor    bool
	Door_open   bool
	Obstruction bool
	Queue       []int
	Resumed     chan bool
}

const (
	Up   = 1
	Down = -1
	Stop = 0
)

//struct med "events" også? slik at

func ElevatorInit(floor_nr int) *Elevator {
	return &Elevator{
		Floor_nr:    floor_nr,
		Direction:   elevio.MD_Stop,
		On_floor:    true,
		Door_open:   false,
		Obstruction: false,
		Queue:       []int{},
		Resumed:     make(chan bool),
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
		fmt.Printf("TargetFloor: %+v\n", targetFloor)
		fmt.Printf("Floor_nr: %+v\n", e.Floor_nr)

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

/*
func obstruction_happened(e Elevator) {
	for e.Obstruction {
	}
	elevio.SetDoorOpenLamp(false)
	// Remove first floor from queue
	e.Queue = e.Queue[1:]

	// Turn off floor button light
	elevio.SetButtonLamp(elevio.BT_HallUp, e.Floor_nr, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, e.Floor_nr, false)
	elevio.SetButtonLamp(elevio.BT_Cab, e.Floor_nr, false)

	fmt.Println("Resuming movement...")
}
*/

func (e *Elevator) checkForObstruction() {
	for e.Obstruction {
		// Wait for obstruction to be cleared (this does not block other activities)
		time.Sleep(500 * time.Millisecond) // Just checking at intervals
	}
	// Once obstruction is cleared, we resume normal operation
	e.Resumed <- true
}
