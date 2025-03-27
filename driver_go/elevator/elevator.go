package elevator

import (
	"Driver_go/config"
	"Driver_go/elevio"
	"fmt"
	"time"
)

type OrderType struct {
	State      bool
	ElevatorID int
	Timestamp  time.Time
}

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_DoorOpen
	EB_Moving
)

const (
	BT_HallUp   = 0
	BT_HallDown = 1
	BT_Cab      = 2
)

type Elevator struct {
	ID            int
	FloorNr       int
	Direction     elevio.MotorDirection
	LastDirection elevio.MotorDirection
	OnFloor       bool
	DoorOpen      bool
	Obstruction   bool
	Orders        [4][3]OrderType
	Behavior      ElevatorBehavior
	LastActive    time.Time
}

var DirectionMap = map[elevio.MotorDirection]string{
	elevio.MD_Up:   "up",
	elevio.MD_Down: "down",
	elevio.MD_Stop: "stop",
}

var BehaviorMap = map[ElevatorBehavior]string{
	EB_Idle:     "idle",
	EB_DoorOpen: "doorOpen",
	EB_Moving:   "moving",
}

type Config struct {
	ClearRequestVariant ClearRequestVariant
}

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

func ElevatorInit(floorNr int, id int) *Elevator {
	return &Elevator{
		ID:            id,
		FloorNr:       floorNr,
		Direction:     elevio.MD_Stop,
		LastDirection: elevio.MD_Up,
		OnFloor:       true,
		DoorOpen:      false,
		Obstruction:   false,
		Orders: [4][3]OrderType{
			{{false, 100, time.Time{}}, {false, 555, time.Time{}}, {false, id, time.Time{}}},
			{{false, 100, time.Time{}}, {false, 100, time.Time{}}, {false, id, time.Time{}}},
			{{false, 100, time.Time{}}, {false, 100, time.Time{}}, {false, id, time.Time{}}},
			{{false, 555, time.Time{}}, {false, 100, time.Time{}}, {false, id, time.Time{}}},
		},
		Behavior:   EB_Idle,
		LastActive: time.Now(),
	}
}

func WaitForValidFloor(d elevio.MotorDirection, drv_floors chan int) int {
	floorChan := make(chan int)
	go func() {
		elevio.SetMotorDirection(d)
		for {
			select {
			case floorSensor := <-drv_floors:
				if floorSensor != -1 {
					fmt.Println("Started at floor: ", floorSensor)
					elevio.SetMotorDirection(elevio.MD_Stop)
					floorChan <- floorSensor
					return
				}
			case <-time.After(500 * time.Millisecond):
				fmt.Println("Waiting for valid floor signal...")
			}
		}
	}()

	return <-floorChan
}

func (e *Elevator) HandleIdleState() {
	e.LastActive = time.Now()
	e.Direction, e.Behavior = e.ChooseDirection()

	switch e.Behavior {
	case EB_Moving:
		e.StartMoving()
	case EB_DoorOpen:
		e.OpenDoor()
	}
}

func (e *Elevator) StartMoving() {
	e.LastActive = time.Now()
	e.Direction = elevio.MotorDirection(e.Direction) // Ensure direction is updated
	e.Behavior = EB_Moving                           // Update behavior
	elevio.SetMotorDirection(e.Direction)
	fmt.Println("Elevator started moving in direction:", e.Direction)
}

func (e *Elevator) OpenDoor() {
	e.LastActive = time.Now()
	//e.LastDirection = e.Direction
	e.Behavior = EB_DoorOpen // Set state to door open
	elevio.SetDoorOpenLamp(true)
	e.ClearAtCurrentFloor()
	fmt.Println("Door opened at floor", e.FloorNr)
}

func (e *Elevator) StopAtFloor() {
	e.LastActive = time.Now()
	e.LastDirection = e.Direction //NEW
	e.Direction = elevio.MD_Stop  // Stop the motor
	e.Behavior = EB_DoorOpen      // Set state to door open
	elevio.SetMotorDirection(e.Direction)
	elevio.SetDoorOpenLamp(true)
	e.ClearAtCurrentFloor()
}

func (e *Elevator) CloseDoorAndResume() {
	e.LastActive = time.Now()
	e.Behavior = EB_Idle
	elevio.SetDoorOpenLamp(false)
	//e.ClearAtCurrentFloor()   //this could be wrong
	e.Direction, e.Behavior = e.ChooseDirection()
	elevio.SetMotorDirection(e.Direction)
	if e.Behavior == EB_DoorOpen {
		e.OpenDoor()
	}
	fmt.Println("Resuming movement in direction:\n", e.Direction)
	fmt.Println("Resuming movement with behavior:\n", e.Behavior)

}

func (e *Elevator) SetLights() {
	for floor := 0; floor < config.NumFloors; floor++ {
		hallUp := e.Orders[floor][BT_HallUp].State
		hallDown := e.Orders[floor][BT_HallDown].State
		cab := e.Orders[floor][BT_Cab].State

		elevio.SetButtonLamp(BT_HallUp, floor, hallUp)
		elevio.SetButtonLamp(BT_HallDown, floor, hallDown)
		elevio.SetButtonLamp(BT_Cab, floor, cab)
	}
}

func (e *Elevator) HasPendingHallOrders() bool {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < config.NumButtons-1; btn++ {
			if e.Orders[floor][btn].State {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) ShouldReopenForOppositeHallCall() bool {
	switch e.LastDirection {
	case elevio.MD_Up:
		return e.Orders[e.FloorNr][BT_HallDown].State &&
			!e.requestsAbove()
		//e.requestsBelow()
	case elevio.MD_Down:
		return e.Orders[e.FloorNr][BT_HallUp].State &&
			!e.requestsBelow()
		//e.requestsAbove()
	default:
		return false
	}
}

func (e *Elevator) clearHallCall(btn int) {
	e.Orders[e.FloorNr][btn].State = false
	e.Orders[e.FloorNr][btn].Timestamp = time.Now()
	e.Orders[e.FloorNr][btn].ElevatorID = 100
}
