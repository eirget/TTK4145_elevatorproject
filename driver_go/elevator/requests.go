package elevator

import (
	"Driver_go/config"
	"Driver_go/elevio"
	"time"
)

func (e *Elevator) requestsAbove() bool {
	for f := e.Floor_nr + 1; f < config.NumFloors; f++ { //blir "+1" feil, er nok riktig
		for btn := 0; btn < config.NumButtons; btn++ {
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
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Orders[f][btn].State && e.Orders[f][btn].ElevatorID == e.ID {
				return true
			}
		}
	}
	return false
}

// requestsHere checks for requests at the current floor.
func (e *Elevator) requestsHere() bool {
	for btn := 0; btn < config.NumButtons; btn++ {
		if e.Orders[e.Floor_nr][btn].State && e.Orders[e.Floor_nr][btn].ElevatorID == e.ID {
			return true
		}
	}
	return false
}

// chooseDirection decides the next direction based on the requests.
func (e *Elevator) ChooseDirection() (elevio.MotorDirection, ElevatorBehavior) {
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
		return true
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

			if !e.requestsAbove() {
				//fmt.Println("Clearing HallDown at Floor", e.Floor_nr)
				e.Orders[e.Floor_nr][BT_HallDown].State = false
				e.Orders[e.Floor_nr][BT_HallDown].Timestamp = time.Now()
			}
		case elevio.MD_Down:
			//fmt.Println("Clearing HallDown at Floor", e.Floor_nr)
			e.Orders[e.Floor_nr][BT_HallDown].State = false
			e.Orders[e.Floor_nr][BT_HallDown].Timestamp = time.Now()

			if !e.requestsBelow() {
				//fmt.Println("Clearing HallUp at Floor", e.Floor_nr)
				e.Orders[e.Floor_nr][BT_HallUp].State = false
				e.Orders[e.Floor_nr][BT_HallUp].Timestamp = time.Now()
			}
		case elevio.MD_Stop:
			//fmt.Println("Clearing HallUp and HallDown at Floor", e.Floor_nr)
			e.Orders[e.Floor_nr][BT_HallUp].State = false
			e.Orders[e.Floor_nr][BT_HallUp].Timestamp = time.Now()
			e.Orders[e.Floor_nr][BT_HallDown].State = false
			e.Orders[e.Floor_nr][BT_HallDown].Timestamp = time.Now()
		}
	}
	//fmt.Printf("After Clearing: Orders at Floor %d: %+v\n", e.Floor_nr, e.Orders[e.Floor_nr])
	e.SetLights()
}

/*
func ReassignOrders(elev Elevator) {
	for f := 0; f < config.NumFloors; f++ {
		for b := 0; b < config.NumButtons; b++ {
			if elev.Orders[f][b].State { //if theres an active order
				AssignToAnotherElevator(f, b)
				elev.Orders[f][b].State = false
			}
		}
	}
}

func AssignToAnotherElevator(floor int, button int) {
	for id, otherElev := range main.elevators { //must get the elevators map from main
		if !otherElev.Obstruction { //find an available elevator
			otherElev.Orders[floor][button].State = true
			fmt.Printf("Reassigned order on floor %d to elevator %s \n", floor, id)
			return
		}
	}
}
*/
