package main

import (
	"Driver_go/elevator"
	"Driver_go/elevio"
	"fmt"
	"time"
)

var new_order_flag bool

func fsm(e *elevator.Elevator,
	elevStateTx chan elevator.Elevator,
	req_chan chan elevio.ButtonEvent,
	new_floor_chan chan int,
	obstr_chan chan bool,
	stop_chan chan bool,
	number_of_floors int,
	run_hra chan bool) {

	doorTimer := time.NewTimer(0)
	<-doorTimer.C

	for {
		select {
		case newReq := <-req_chan:
			fsmHandleRequestButtonPress(newReq, e, elevStateTx, &new_order_flag)

		case <-time.After(100 * time.Millisecond):
			fsmHandleIdleState(e, doorTimer)

		case newFloor := <-new_floor_chan:
			fsmHandleNewFloor(newFloor, e, elevStateTx, doorTimer)

		case <-doorTimer.C:
			fsmHandleDoorTimeout(e, doorTimer)

		case isObstructed := <-obstr_chan:
			fmt.Println("Obstruction happened")
			fsmHandleObstruction(isObstructed, e)

		case isStopped := <-stop_chan:
			fsmHandleEmergencyStop(isStopped, e, number_of_floors)

		case <-time.After(1 * time.Second):
			e.SetLights()
		}
	}
}

func fsmHandleRequestButtonPress(a elevio.ButtonEvent, e *elevator.Elevator, elevStateTx chan elevator.Elevator, new_order_flag *bool) {
	fmt.Printf("%+v\n", a)
	e.Orders[a.Floor][a.Button].State = true
	e.Orders[a.Floor][a.Button].Timestamp = time.Now()

	elevStateTx <- *e

	fmt.Println("orders after button press: ", e.Orders)

	if a.Button == elevio.BT_Cab {
		elevio.SetButtonLamp(a.Button, a.Floor, true)
	}
	*new_order_flag = true
	//elevStateTx <- *elev
}

func fsmHandleIdleState(e *elevator.Elevator, doorTimer *time.Timer) {
	if e.Behavior == elevator.EB_Idle {
		e.HandleIdleState()
		if e.Behavior == elevator.EB_DoorOpen {
			doorTimer.Reset(5 * time.Second)
		}
	}
}

func fsmHandleNewFloor(a int, e *elevator.Elevator, elevStateTx chan elevator.Elevator, doorTimer *time.Timer) {
	e.Floor_nr = a
	elevio.SetFloorIndicator(a)

	if e.ShouldStop() {
		e.StopAtFloor()
		doorTimer.Reset(3 * time.Second)
	}

	elevStateTx <- *e
}

func fsmHandleDoorTimeout(e *elevator.Elevator, doorTimer *time.Timer) {
	if e.Obstruction {
		fmt.Println("Waiting for obstruction to clear...")
		doorTimer.Reset(500 * time.Millisecond)
	} else {
		e.CloseDoorAndResume()
	}
}

func fsmHandleObstruction(isObstructed bool, e *elevator.Elevator) {
	e.Obstruction = isObstructed

	if isObstructed {
		fmt.Println("Obstruction detected! Waiting for it to clear...")
	}
}

func fsmHandleEmergencyStop(a bool, e *elevator.Elevator, number_of_floors int) {
	fmt.Printf("%+v\n", a)
	elevio.SetStopLamp(true)
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.Behavior = elevator.EB_Idle
	e.Direction = elevio.MD_Stop
	e.Orders = [4][3]elevator.OrderType{}

	for floor := 0; floor < number_of_floors; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(btn, floor, false)
		}
	}

	time.Sleep(5 * time.Second)
	elevio.SetStopLamp(false)
}
