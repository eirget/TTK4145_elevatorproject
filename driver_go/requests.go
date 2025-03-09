package main

import "Driver_go/elevio"

func (e *Elevator) requestsAbove() bool {
	for f := e.Floor_nr + 1; f < NumFloors; f++ { //blir "+1" feil, er nok riktig
		for btn := 0; btn < NumButtons; btn++ { 
			if e.Orders[f][btn].State && e.Orders[f][btn].ElevatorID == e.ID {
				return true
			}
		}
	}
	return false
}

// requestsBelow checks for requests below the current floor.
func (e *Elevator) requestsBelow() bool {
	for f := 0; f < e.Floor_nr; f++ {
		for btn := 0; btn < NumButtons; btn++ { 
			if e.Orders[f][btn].State && e.Orders[f][btn].ElevatorID == e.ID {
				return true
			}
		}
	}
	return false
}

// requestsHere checks for requests at the current floor.
func (e *Elevator) requestsHere() bool {
	for btn := 0; btn < NumButtons; btn++ {
		if e.Orders[e.Floor_nr][btn].State && e.Orders[e.Floor_nr][btn].ElevatorID == e.ID {
			return true
		}
	}
	return false
}

// chooseDirection decides the next direction based on the requests.
func (e *Elevator) chooseDirection() (elevio.MotorDirection, ElevatorBehavior) {
	switch e.Direction {
	case elevio.MD_Up:
		if e.requestsAbove() {
			return elevio.MD_Up, EB_Moving
		}
		if e.requestsHere() {
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
		if e.requestsHere() {
			return elevio.MD_Up, EB_DoorOpen
		}
		if e.requestsAbove() {
			return elevio.MD_Up, EB_Moving
		}
		return elevio.MD_Stop, EB_Idle
	case elevio.MD_Stop:
		if e.requestsHere() {
			return elevio.MD_Stop, EB_DoorOpen
		}
		if e.requestsAbove() {
			return elevio.MD_Up, EB_Moving
		}
		if e.requestsBelow() {
			return elevio.MD_Down, EB_Moving
		}
		return elevio.MD_Stop, EB_Idle
	default:
		return elevio.MD_Stop, EB_Idle
	}
}

// shouldStop checks if the elevator should stop at the current floor.
func (e *Elevator) shouldStop() bool {
	switch e.Direction {
	case elevio.MD_Down:
		return (e.Orders[e.Floor_nr][BT_HallDown].State && e.Orders[e.Floor_nr][BT_HallDown].ElevatorID == e.ID) ||
			(e.Orders[e.Floor_nr][BT_Cab].State && e.Orders[e.Floor_nr][BT_Cab].ElevatorID == e.ID) ||
			!e.requestsBelow()
	case elevio.MD_Up:
		return (e.Orders[e.Floor_nr][BT_HallUp].State && e.Orders[e.Floor_nr][BT_HallUp].ElevatorID == e.ID) ||
			(e.Orders[e.Floor_nr][BT_Cab].State && e.Orders[e.Floor_nr][BT_Cab].ElevatorID == e.ID) ||
			!e.requestsAbove()
	case elevio.MD_Stop:
		return true
	default:
		return false
	}
}

// clearAtCurrentFloor clears requests at the current floor.
func (e *Elevator) clearAtCurrentFloor() {
	switch e.Config.ClearRequestVariant {
	case CV_All:
		for btn := 0; btn < NumButtons; btn++ {
			if e.Orders[e.Floor_nr][btn].ElevatorID == e.ID {
				e.Orders[e.Floor_nr][btn].State = false
			}
		}
	case CV_InDirn:
		if e.Orders[e.Floor_nr][BT_Cab].ElevatorID == e.ID {
			e.Orders[e.Floor_nr][BT_Cab].State = false
			elevio.SetButtonLamp(BT_Cab, e.Floor_nr, false)
		}
		switch e.Direction {
		case elevio.MD_Up:
			if !e.requestsAbove() && e.Orders[e.Floor_nr][BT_HallUp].ElevatorID == e.ID {
				e.Orders[e.Floor_nr][BT_HallDown].State = false
				elevio.SetButtonLamp(BT_HallDown, e.Floor_nr, false)
			}
			if e.Orders[e.Floor_nr][BT_HallUp].ElevatorID == e.ID {
				e.Orders[e.Floor_nr][BT_HallUp].State = false
				elevio.SetButtonLamp(BT_HallUp, e.Floor_nr, false)
			}
		case elevio.MD_Down:
			if !e.requestsBelow() && e.Orders[e.Floor_nr][BT_HallDown].ElevatorID == e.ID {
				e.Orders[e.Floor_nr][BT_HallUp].State = false
				elevio.SetButtonLamp(BT_HallUp, e.Floor_nr, false)
			}
			if e.Orders[e.Floor_nr][BT_HallDown].ElevatorID == e.ID {
				e.Orders[e.Floor_nr][BT_HallDown].State = false
				elevio.SetButtonLamp(BT_HallDown, e.Floor_nr, false)
			}
		case elevio.MD_Stop:
			if e.Orders[e.Floor_nr][BT_HallUp].ElevatorID == e.ID {
				e.Orders[e.Floor_nr][BT_HallUp].State = false
				elevio.SetButtonLamp(BT_HallUp, e.Floor_nr, false)
			}
			if e.Orders[e.Floor_nr][BT_HallDown].ElevatorID == e.ID {
				e.Orders[e.Floor_nr][BT_HallDown].State = false
				elevio.SetButtonLamp(BT_HallDown, e.Floor_nr, false)
			}
		}
	}
}
