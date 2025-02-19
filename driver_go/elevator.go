package main

import (
	"Driver_go/elevio"
)

type OrderType struct {
	State bool
	ElevatorID int
}

type Elevator struct {
	//mutex over states maybe to protect
	Floor_nr    int
	Direction   elevio.MotorDirection
	On_floor    bool
	Door_open   bool
	Obstruction bool
	Orders [3][4]OrderType
}

func ElevatorInit(floor_nr int) *Elevator {
	return &Elevator{
		Floor_nr:    floor_nr,
		Direction:   elevio.MD_Stop,
		On_floor:    true,
		Door_open:   false,
		Obstruction: false,
		Orders: [3][4]OrderType{
				{{false, 0}, {false,0}, {false, 0}, {false, 0}},
				{{false, 0}, {false,0}, {false, 0}, {false, 0}},
				{{false, 0}, {false,0}, {false, 0}, {false, 0}},
		},
	}
}