package elevator

import (
	"Driver_go/config"
	"Driver_go/elevio"
	"time"
)

func (e *Elevator) requestsAbove() bool {
	for floor := e.Floor_nr + 1; floor < config.NumFloors; floor++ { //blir "+1" feil, er nok riktig
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
	for floor := 0; floor < e.Floor_nr; floor++ {
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
		if e.Orders[e.Floor_nr][btn].State && e.Orders[e.Floor_nr][btn].ElevatorID == e.ID {
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
		if e.RequestsHere() {
			return elevio.MD_Stop, EB_Idle //did say DoorOpen
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
func (e *Elevator) ShouldStop() bool {
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
		return e.RequestsHere()
	default:
		return false
	}
}

// clearAtCurrentFloor clears requests at the current floor.
func (e *Elevator) ClearAtCurrentFloor() {
	//fmt.Printf("Before Clearing: Orders at Floor %d: %+v\n", e.Floor_nr, e.Orders[e.Floor_nr])

	switch e.Config.ClearRequestVariant {
	case CV_All:
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Orders[e.Floor_nr][btn].ElevatorID == e.ID {
				e.Orders[e.Floor_nr][btn].State = false
			}
		}
	case CV_InDirn:
		//fmt.Println("Clearing CAB order at Floor", e.Floor_nr)
		e.Orders[e.Floor_nr][BT_Cab].State = false
		e.Orders[e.Floor_nr][BT_Cab].Timestamp = time.Now()

		switch e.Direction {
		case elevio.MD_Up:
			//fmt.Println("Clearing HallUp at Floor", e.Floor_nr)
			e.Orders[e.Floor_nr][BT_HallUp].State = false
			e.Orders[e.Floor_nr][BT_HallUp].Timestamp = time.Now()
			e.Orders[e.Floor_nr][BT_HallUp].ElevatorID = 100

			if !e.requestsAbove() {
				//fmt.Println("Clearing HallDown at Floor", e.Floor_nr)
				e.Orders[e.Floor_nr][BT_HallDown].State = false
				e.Orders[e.Floor_nr][BT_HallDown].Timestamp = time.Now()
				e.Orders[e.Floor_nr][BT_HallDown].ElevatorID = 100
			}
		case elevio.MD_Down:
			//fmt.Println("Clearing HallDown at Floor", e.Floor_nr)
			e.Orders[e.Floor_nr][BT_HallDown].State = false
			e.Orders[e.Floor_nr][BT_HallDown].Timestamp = time.Now()
			e.Orders[e.Floor_nr][BT_HallDown].ElevatorID = 100

			if !e.requestsBelow() {
				//fmt.Println("Clearing HallUp at Floor", e.Floor_nr)
				e.Orders[e.Floor_nr][BT_HallUp].State = false
				e.Orders[e.Floor_nr][BT_HallUp].Timestamp = time.Now()
				e.Orders[e.Floor_nr][BT_HallUp].ElevatorID = 100
			}
		case elevio.MD_Stop:
			// Serve only one hall call at a time
			if e.Orders[e.Floor_nr][BT_HallUp].State && e.Orders[e.Floor_nr][BT_HallDown].State {
				if e.Orders[e.Floor_nr][BT_HallDown].Timestamp.After(e.Orders[e.Floor_nr][BT_HallUp].Timestamp) {
					// Serve UP first
					e.Orders[e.Floor_nr][BT_HallUp].State = false
					e.Orders[e.Floor_nr][BT_HallUp].Timestamp = time.Now()
					e.Orders[e.Floor_nr][BT_HallUp].ElevatorID = 100

					e.PendingSecondCall = true
				} else {
					// Serve DOWN first
					e.Orders[e.Floor_nr][BT_HallDown].State = false
					e.Orders[e.Floor_nr][BT_HallDown].Timestamp = time.Now()
					e.Orders[e.Floor_nr][BT_HallDown].ElevatorID = 100

					e.PendingSecondCall = true
				}
				return
			}

			// If only one is active
			if e.Orders[e.Floor_nr][BT_HallUp].State {
				e.Orders[e.Floor_nr][BT_HallUp].State = false
				e.Orders[e.Floor_nr][BT_HallUp].Timestamp = time.Now()
				e.Orders[e.Floor_nr][BT_HallUp].ElevatorID = 100
				return
			}

			if e.Orders[e.Floor_nr][BT_HallDown].State {
				e.Orders[e.Floor_nr][BT_HallDown].State = false
				e.Orders[e.Floor_nr][BT_HallDown].Timestamp = time.Now()
				e.Orders[e.Floor_nr][BT_HallDown].ElevatorID = 100
				return
			}
		}
	}
	//fmt.Printf("After Clearing: Orders at Floor %d: %+v\n", e.Floor_nr, e.Orders[e.Floor_nr])
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
