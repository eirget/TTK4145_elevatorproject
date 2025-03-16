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
	//mutex over states maybe to protect
	ID          int
	Floor_nr    int
	Direction   elevio.MotorDirection
	On_floor    bool
	Door_open   bool
	Obstruction bool
	Orders      [4][3]OrderType
	Behavior    ElevatorBehavior
	LastActive  time.Time
	Config      Config
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

func ElevatorInit(floor_nr int, id int) *Elevator {
	return &Elevator{
		ID:          id,
		Floor_nr:    floor_nr,
		Direction:   elevio.MD_Stop,
		On_floor:    true,
		Door_open:   false,
		Obstruction: false,
		Orders: [4][3]OrderType{
			{{false, 100, time.Time{}}, {false, 555, time.Time{}}, {false, id, time.Time{}}},
			{{false, 100, time.Time{}}, {false, 100, time.Time{}}, {false, id, time.Time{}}},
			{{false, 100, time.Time{}}, {false, 100, time.Time{}}, {false, id, time.Time{}}},
			{{false, 555, time.Time{}}, {false, 100, time.Time{}}, {false, id, time.Time{}}},
		},
		Behavior:   EB_Idle,
		LastActive: time.Now(),
		Config:     Config{ClearRequestVariant: CV_InDirn},
	}
}

func (e *Elevator) HandleIdleState() {
	e.LastActive = time.Now()
	dirn, newBehavior := e.ChooseDirection()
	e.Behavior = newBehavior
	e.Direction = dirn

	switch newBehavior {
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
	e.Behavior = EB_DoorOpen // Set state to door open
	elevio.SetDoorOpenLamp(true)
	e.ClearAtCurrentFloor()
	fmt.Println("Door opened at floor", e.Floor_nr)
}

func (e *Elevator) StopAtFloor() {
	e.LastActive = time.Now()
	e.Direction = elevio.MD_Stop // Stop the motor
	e.Behavior = EB_DoorOpen     // Set state to door open
	elevio.SetMotorDirection(e.Direction)
	elevio.SetDoorOpenLamp(true)
	e.ClearAtCurrentFloor()
}

func (e *Elevator) CloseDoorAndResume() {
	e.LastActive = time.Now()
	e.Behavior = EB_Idle
	elevio.SetDoorOpenLamp(false)
	e.Direction, e.Behavior = e.ChooseDirection()
	elevio.SetMotorDirection(e.Direction)
	fmt.Println("Resuming movement in direction:", e.Direction)
}

func (e *Elevator) SetLights() {
	for f := 0; f < config.NumFloors; f++ {
		for b := elevio.ButtonType(0); b < config.NumButtons; b++ {

			// Hall lights should be identical across all elevators
			if b == BT_HallUp || b == BT_HallDown {
				// Set the hall light ON/OFF based on the shared Orders array
				elevio.SetButtonLamp(b, f, e.Orders[f][b].State)

				// Cab lights should be set based on each individual elevator's Orders
			} else if b == BT_Cab {
				elevio.SetButtonLamp(BT_Cab, f, e.Orders[f][BT_Cab].State)
			}
		}
	}
}
