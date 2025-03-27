package elevator

import (
	"Driver_go/config"
	"Driver_go/elevio"
	"time"
)

func (e *Elevator) requestsAbove() bool {
	for floor := e.FloorNr + 1; floor < config.NumFloors; floor++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Orders[floor][btn].State && e.Orders[floor][btn].ElevatorID == e.ID {
				return true
			}
		}
	}
	return false
}

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
	case elevio.MDUp:
		if e.requestsAbove() {
			return elevio.MDUp, EBMoving
		}
		if e.RequestsHere() {
			return elevio.MDDown, EBDoorOpen
		}
		if e.requestsBelow() {
			return elevio.MDDown, EBMoving
		}
		return elevio.MDStop, EBIdle
	case elevio.MDDown:
		if e.requestsBelow() {
			return elevio.MDDown, EBMoving
		}
		if e.RequestsHere() {
			return elevio.MDUp, EBDoorOpen
		}
		if e.requestsAbove() {
			return elevio.MDUp, EBMoving
		}
		return elevio.MDStop, EBIdle
	case elevio.MDStop:
		switch e.LastDirection {
		case elevio.MDUp:
			if e.requestsAbove() {
				return elevio.MDUp, EBMoving
			}
			if e.requestsBelow() {
				return elevio.MDDown, EBMoving
			}
			if e.RequestsHere() {
				return elevio.MDStop, EBDoorOpen
			}
		case elevio.MDDown:
			if e.requestsBelow() {
				return elevio.MDDown, EBMoving
			}
			if e.requestsAbove() {
				return elevio.MDUp, EBMoving
			}
			if e.RequestsHere() {
				return elevio.MDStop, EBDoorOpen
			}
		}

		return elevio.MDStop, EBIdle
	default:
		return elevio.MDStop, EBIdle
	}
}

func (e *Elevator) ShouldStop() bool {
	switch e.Direction {
	case elevio.MDDown:
		return (e.Orders[e.FloorNr][BTHallDown].State && e.Orders[e.FloorNr][BTHallDown].ElevatorID == e.ID) ||
			(e.Orders[e.FloorNr][BTCab].State && e.Orders[e.FloorNr][BTCab].ElevatorID == e.ID) ||
			!e.requestsBelow()
	case elevio.MDUp:
		return (e.Orders[e.FloorNr][BTHallUp].State && e.Orders[e.FloorNr][BTHallUp].ElevatorID == e.ID) ||
			(e.Orders[e.FloorNr][BTCab].State && e.Orders[e.FloorNr][BTCab].ElevatorID == e.ID) ||
			!e.requestsAbove()
	case elevio.MDStop:
		return e.RequestsHere()
	default:
		return false
	}
}

func (e *Elevator) ClearAtCurrentFloor() {
	if e.Orders[e.FloorNr][BTCab].ElevatorID == e.ID {
		e.Orders[e.FloorNr][BTCab].State = false
		e.Orders[e.FloorNr][BTCab].Timestamp = time.Now()
	}
	switch e.Direction {
	case elevio.MDUp:
		if e.Orders[e.FloorNr][BTHallUp].ElevatorID == e.ID {
			e.clearHallCall(BTHallUp)
		}
		if !e.requestsAbove() && e.Orders[e.FloorNr][BTHallDown].ElevatorID == e.ID {
			e.clearHallCall(BTHallDown)
		}
	case elevio.MDDown:
		if e.Orders[e.FloorNr][BTHallDown].ElevatorID == e.ID {
			e.clearHallCall(BTHallDown)
		}
		if !e.requestsBelow() && e.Orders[e.FloorNr][BTHallUp].ElevatorID == e.ID {
			e.clearHallCall(BTHallUp)
		}
	case elevio.MDStop:
		// Clear hall call in the direction we came from
		switch e.LastDirection {
		case elevio.MDUp:
			if e.Orders[e.FloorNr][elevio.BTHallUp].State && e.Orders[e.FloorNr][BTHallUp].ElevatorID == e.ID && (e.FloorNr != config.NumFloors) {
				e.clearHallCall(BTHallUp)
			} else if e.Orders[e.FloorNr][BTHallDown].ElevatorID == e.ID {
				e.clearHallCall(BTHallDown)
			}
		case elevio.MDDown:
			if e.Orders[e.FloorNr][elevio.BTHallDown].State && e.Orders[e.FloorNr][BTHallDown].ElevatorID == e.ID && (e.FloorNr != 0) {
				e.clearHallCall(BTHallDown)
			} else if e.Orders[e.FloorNr][BTHallUp].ElevatorID == e.ID {
				e.clearHallCall(BTHallUp)
			}
		}
	}
}

func AssignAllHallCallsToSelf(e *Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn <= 1; btn++ {
			if e.Orders[floor][btn].State {
				e.Orders[floor][btn].ElevatorID = e.ID
				e.Orders[floor][btn].Timestamp = time.Now()
			}
		}
	}
}
