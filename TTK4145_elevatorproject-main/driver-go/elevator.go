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
	// dette vil ikke fungere, fordi hvis noen skal fra 2 til 3, mens noen skal fra 1 til 3. Vil ikke den siste ble med
	// kanskje heller sjekke for duplikater hos elementene som er rett ved siden av det elementet som legges til i køen

	e.Queue = append(e.Queue, floor)
	fmt.Println("Added floor:", floor, "Updated Queue:", e.Queue)
}

//struct med "events" også? slik at

func ElevatorInit() *Elevator {
	return &Elevator{
		Floor_nr:  elevio.GetFloor(),
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
		//fmt.Println("You are in the queue")

		// Problems: når heis skal ned, går den først en etasje opp før den går ned til rett etasje
		// Når vi er i øverste etasje prøver den på det samme, og låser seg 
		// Per nå klikker den etter at dør har åpnet og lukket seg igjen etter første request (kø eller dørlys som er problem?)
		if e.Floor_nr < targetFloor {
			fmt.Println("Moving Up...")
			elevio.SetMotorDirection(elevio.MD_Up) // burde disse to linjene bytte rekkefølge? blir det dumt å gi en garanti før det skjer?
		} else if e.Floor_nr > targetFloor {
			fmt.Println("Moving Down...")
			elevio.SetMotorDirection(elevio.MD_Down)
		}

		// Wait for floor update in the main FSM
		time.Sleep(100 * time.Millisecond)
	}
}
