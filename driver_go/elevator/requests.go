package elevator

import (
	"Driver_go/config"
	"Driver_go/elevio"
	"fmt"
	"time"
)

func (e *Elevator) requestsAbove() bool {
	for floor := e.FloorNr + 1; floor < config.NumFloors; floor++ { //blir "+1" feil, er nok riktig
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Orders[floor][btn].State && e.Orders[floor][btn].ElevatorID == e.ID {
				return true
			}
		}
	}
	return false
}

// requestsBelow checks for requests below the current floor.
func (e *Elevator) requestsBelow() bool {
	for floor := 0; floor < e.FloorNr; floor++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Orders[floor][btn].State && e.Orders[floor][btn].ElevatorID == e.ID {
				return true
			}
		}
	}
	return false
}

// requestsHere checks for requests at the current floor.
func (e *Elevator) RequestsHere() bool {
	for btn := 0; btn < config.NumButtons; btn++ {
		if e.Orders[e.FloorNr][btn].State && e.Orders[e.FloorNr][btn].ElevatorID == e.ID {
			return true
		}
	}
	return false
}

// chooseDirection decides the next direction based on Orders
func (e *Elevator) ChooseDirection() (elevio.MotorDirection, ElevatorBehavior) {
	switch e.Direction {
	case elevio.MD_Up:
		if e.requestsAbove() {
			return elevio.MD_Up, EB_Moving
		}
		if e.RequestsHere() {
			return elevio.MD_Down, EB_DoorOpen
		}
		if e.requestsBelow() {
			return elevio.MD_Down, EB_Moving
		}
		return elevio.MD_Stop, EB_Idle
	case elevio.MD_Down:
		if e.requestsBelow() {
			return elevio.MD_Down, EB_Moving
		}
		if e.RequestsHere() {
			return elevio.MD_Up, EB_DoorOpen
		}
		if e.requestsAbove() {
			return elevio.MD_Up, EB_Moving
		}
		return elevio.MD_Stop, EB_Idle
	case elevio.MD_Stop:
		switch e.LastDirection {
		case elevio.MD_Up:
			if e.requestsAbove() {
				return elevio.MD_Up, EB_Moving
			}
			if e.requestsBelow() {
				return elevio.MD_Down, EB_Moving
			}
			if e.RequestsHere() {
				//if e.LastDirection
				return elevio.MD_Stop, EB_DoorOpen
			}
		case elevio.MD_Down:
			if e.requestsBelow() {
				return elevio.MD_Down, EB_Moving
			}
			if e.requestsAbove() {
				return elevio.MD_Up, EB_Moving
			}
			if e.RequestsHere() {
				//if e.LastDirection
				return elevio.MD_Stop, EB_DoorOpen
			}
		}

		return elevio.MD_Stop, EB_Idle
	default:
		return elevio.MD_Stop, EB_Idle
	}
}

// shouldStop checks if the elevator should stop at the current floor.
func (e *Elevator) ShouldStop() bool {
	switch e.Direction {
	case elevio.MD_Down:
		return (e.Orders[e.FloorNr][BT_HallDown].State && e.Orders[e.FloorNr][BT_HallDown].ElevatorID == e.ID) ||
			(e.Orders[e.FloorNr][BT_Cab].State && e.Orders[e.FloorNr][BT_Cab].ElevatorID == e.ID) ||
			!e.requestsBelow()
	case elevio.MD_Up:
		return (e.Orders[e.FloorNr][BT_HallUp].State && e.Orders[e.FloorNr][BT_HallUp].ElevatorID == e.ID) ||
			(e.Orders[e.FloorNr][BT_Cab].State && e.Orders[e.FloorNr][BT_Cab].ElevatorID == e.ID) ||
			!e.requestsAbove()
	case elevio.MD_Stop:
		return e.RequestsHere()
	default:
		return false
	}
}

func (e *Elevator) ClearAtCurrentFloor() {
	fmt.Printf("Clearing Orders at floor %v with LastDirection: %v\n", e.FloorNr, e.LastDirection)
	// Always clear cab call
	if e.Orders[e.FloorNr][BT_Cab].ElevatorID == e.ID {
		e.Orders[e.FloorNr][BT_Cab].State = false
		e.Orders[e.FloorNr][BT_Cab].Timestamp = time.Now()
	}
	switch e.Direction {
	case elevio.MD_Up:
		if e.Orders[e.FloorNr][BT_HallUp].ElevatorID == e.ID {
			e.clearHallCall(BT_HallUp)
		}
		if !e.requestsAbove() && e.Orders[e.FloorNr][BT_HallDown].ElevatorID == e.ID {
			e.clearHallCall(BT_HallDown)
		}
	case elevio.MD_Down:
		if e.Orders[e.FloorNr][BT_HallDown].ElevatorID == e.ID {
			e.clearHallCall(BT_HallDown)
		}
		if !e.requestsBelow() && e.Orders[e.FloorNr][BT_HallUp].ElevatorID == e.ID {
			e.clearHallCall(BT_HallUp)
		}
	case elevio.MD_Stop:
		// Clear hall call in the direction we just came from
		switch e.LastDirection {
		case elevio.MD_Up:
			if e.Orders[e.FloorNr][elevio.BT_HallUp].State && e.Orders[e.FloorNr][BT_HallUp].ElevatorID == e.ID && (e.FloorNr != config.NumFloors) {
				e.clearHallCall(BT_HallUp)
			} else if e.Orders[e.FloorNr][BT_HallDown].ElevatorID == e.ID {
				e.clearHallCall(BT_HallDown)
			}
		case elevio.MD_Down:
			if e.Orders[e.FloorNr][elevio.BT_HallDown].State && e.Orders[e.FloorNr][BT_HallDown].ElevatorID == e.ID && (e.FloorNr != 0) {
				e.clearHallCall(BT_HallDown)
			} else if e.Orders[e.FloorNr][BT_HallUp].ElevatorID == e.ID {
				e.clearHallCall(BT_HallUp)
			}
		}
	}
}

func AssignAllHallCallsToSelf(e *Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn <= 1; btn++ { // HallUp and HallDown only
			if e.Orders[floor][btn].State {
				e.Orders[floor][btn].ElevatorID = e.ID
				e.Orders[floor][btn].Timestamp = time.Now()
			}
		}
	}
}
