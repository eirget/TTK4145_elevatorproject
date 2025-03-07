package main

import (
	"Driver_go/elevio"
)

type OrderType struct {
	State      bool
	ElevatorID int
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
	Config      Config
}

var directionMap = map[elevio.MotorDirection]string{
	elevio.MD_Up:   "up",
	elevio.MD_Down: "down",
	elevio.MD_Stop: "stop",
}

var behaviorMap = map[ElevatorBehavior]string{
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
			{{false, 1000}, {false, 5}, {false, id}},
			{{false, 1000}, {false, 1000}, {false, id}},
			{{false, 1000}, {false, 1000}, {false, id}},
			{{false, 5}, {false, 1000}, {false, id}},
		},
		Behavior: EB_Idle,
		Config:   Config{ClearRequestVariant: CV_InDirn},
	}
}
