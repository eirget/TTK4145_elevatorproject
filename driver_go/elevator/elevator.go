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
	EBIdle ElevatorBehavior = iota
	EBDoorOpen
	EBMoving
)

const (
	BTHallUp   = 0
	BTHallDown = 1
	BTCab      = 2
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
	elevio.MDUp:   "up",
	elevio.MDDown: "down",
	elevio.MDStop: "stop",
}

var BehaviorMap = map[ElevatorBehavior]string{
	EBIdle:     "idle",
	EBDoorOpen: "doorOpen",
	EBMoving:   "moving",
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
		Direction:     elevio.MDStop,
		LastDirection: elevio.MDUp,
		OnFloor:       true,
		DoorOpen:      false,
		Obstruction:   false,
		Orders: [4][3]OrderType{
			{{false, 100, time.Time{}}, {false, 555, time.Time{}}, {false, id, time.Time{}}},
			{{false, 100, time.Time{}}, {false, 100, time.Time{}}, {false, id, time.Time{}}},
			{{false, 100, time.Time{}}, {false, 100, time.Time{}}, {false, id, time.Time{}}},
			{{false, 555, time.Time{}}, {false, 100, time.Time{}}, {false, id, time.Time{}}},
		},
		Behavior:   EBIdle,
		LastActive: time.Now(),
	}
}

func WaitForValidFloor(d elevio.MotorDirection, drvFloors chan int) int {
	floorCh := make(chan int)
	go FloorInit(d, drvFloors, floorCh) // HAR NÅ DEFINERT DENNE UNDER. SJEKK OM DET FUNKER
	return <-floorCh
}

func FloorInit(d elevio.MotorDirection, drvFloors chan int, floorCh chan int) {
	elevio.SetMotorDirection(d)
	for {
		select {
		case floorSensor := <-drvFloors: // BURDE DISSE TO CASENE BYTTE PLASS?
			if floorSensor != -1 {
				fmt.Println("Started at floor: ", floorSensor)
				elevio.SetMotorDirection(elevio.MDStop)
				floorCh <- floorSensor
				return
			}
		case <-time.After(500 * time.Millisecond):
			fmt.Println("Waiting for valid floor signal...")
		}
	}
}

func (e *Elevator) HandleIdleState(doorTimer *time.Timer) {
	e.LastActive = time.Now()
	e.Direction, e.Behavior = e.ChooseDirection()

	switch e.Behavior {
	case EBMoving:
		e.StartMoving()
	case EBDoorOpen:
		e.OpenDoor(doorTimer)
	}
}

func (e *Elevator) StartMoving() {
	e.LastActive = time.Now()
	e.Direction = elevio.MotorDirection(e.Direction)
	e.Behavior = EBMoving
	elevio.SetMotorDirection(e.Direction)
}

func (e *Elevator) OpenDoor(doorTimer *time.Timer) {
	e.LastActive = time.Now()
	e.Behavior = EBDoorOpen
	elevio.SetDoorOpenLamp(true)
	e.ClearAtCurrentFloor()

	if doorTimer != nil {
		doorTimer.Reset(3 * time.Second)
	}
}

func (e *Elevator) StopAtFloor() {
	e.LastActive = time.Now()
	e.LastDirection = e.Direction
	e.Direction = elevio.MDStop
	e.Behavior = EBDoorOpen
	elevio.SetMotorDirection(e.Direction)
	elevio.SetDoorOpenLamp(true)
	e.ClearAtCurrentFloor()
}

func (e *Elevator) CloseDoorAndResume(doorTimer *time.Timer) {
	e.LastActive = time.Now()
	e.Behavior = EBIdle
	elevio.SetDoorOpenLamp(false)
	e.Direction, e.Behavior = e.ChooseDirection()
	elevio.SetMotorDirection(e.Direction)
	if e.Behavior == EBDoorOpen {
		e.OpenDoor(doorTimer)
	}
}

func (e *Elevator) SetLights() {
	for floor := 0; floor < config.NumFloors; floor++ {
		hallUp := e.Orders[floor][BTHallUp].State
		hallDown := e.Orders[floor][BTHallDown].State
		cab := e.Orders[floor][BTCab].State

		elevio.SetButtonLamp(BTHallUp, floor, hallUp)
		elevio.SetButtonLamp(BTHallDown, floor, hallDown)
		elevio.SetButtonLamp(BTCab, floor, cab)
	}
}

func (e *Elevator) HasPendingHallOrders() bool {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < config.NumHallButtons; btn++ {
			if e.Orders[floor][btn].State {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) ShouldReopenForOppositeHallCall() bool {
	switch e.LastDirection {
	case elevio.MDUp:
		return e.Orders[e.FloorNr][BTHallDown].State &&
			!e.requestsAbove()
	case elevio.MDDown:
		return e.Orders[e.FloorNr][BTHallUp].State &&
			!e.requestsBelow()
	default:
		return false
	}
}

func (e *Elevator) ClearHallCall(btn int) {
	e.Orders[e.FloorNr][btn].State = false
	e.Orders[e.FloorNr][btn].Timestamp = time.Now()
	e.Orders[e.FloorNr][btn].ElevatorID = 100
}
