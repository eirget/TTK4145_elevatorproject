package main


type Elevator struct {
	//mutex over states maybe to protect
	floor_nr int
	direction int
	on_floor bool
	door_open bool
}

//struct med "events" også? slik at

func elevatorInit() *Elevator{
	return &Elevator{
		floor_nr: 1,
		direction: 0,
		on_floor: false,
		door_open: false,
	}
}