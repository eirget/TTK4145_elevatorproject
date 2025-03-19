package main

import (
	"Driver_go/elevator"
	"Driver_go/elevio"
	"fmt"
	"time"
)

var new_order_flag bool

func fsm(elev *elevator.Elevator,
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
		case a := <-req_chan:
			fsmHandleRequestButtonPress(a, elev, elevStateTx, &new_order_flag)

		case <-time.After(100 * time.Millisecond):
			fsmHandleIdleState(elev, doorTimer)

		case a := <-new_floor_chan:
			fsmHandleNewFloor(a, elev, elevStateTx, doorTimer)

		case <-doorTimer.C:
			fsmHandleDoorTimeout(elev, doorTimer)

		case isObstructed := <-obstr_chan:
			fmt.Println("Obstruction happened")
			fsmHandleObstruction(isObstructed, elev)

		case a := <-stop_chan:
			fsmHandleEmergencyStop(a, elev, number_of_floors)

		case <-time.After(1 * time.Second):
			elev.SetLights()
		}
	}
}

func fsmHandleRequestButtonPress(a elevio.ButtonEvent, elev *elevator.Elevator, elevStateTx chan elevator.Elevator, new_order_flag *bool) {
	fmt.Printf("%+v\n", a)
	elev.Orders[a.Floor][a.Button].State = true
	elev.Orders[a.Floor][a.Button].Timestamp = time.Now()

	elevStateTx <- *elev

	fmt.Println("orders after button press: ", elev.Orders)

	if a.Button == elevio.BT_Cab {
		elevio.SetButtonLamp(a.Button, a.Floor, true)
	}
	*new_order_flag = true
	//elevStateTx <- *elev
}

func fsmHandleIdleState(elev *elevator.Elevator, doorTimer *time.Timer) {
	if elev.Behavior == elevator.EB_Idle {
		elev.HandleIdleState()
		if elev.Behavior == elevator.EB_DoorOpen {
			doorTimer.Reset(5 * time.Second)
		}
	}
}

func fsmHandleNewFloor(a int, elev *elevator.Elevator, elevStateTx chan elevator.Elevator, doorTimer *time.Timer) {
	elev.Floor_nr = a
	elevio.SetFloorIndicator(a)

	if elev.ShouldStop() {
		elev.StopAtFloor()
		doorTimer.Reset(3 * time.Second)
	}

	elevStateTx <- *elev
}

func fsmHandleDoorTimeout(elev *elevator.Elevator, doorTimer *time.Timer) {
	if elev.Obstruction {
		fmt.Println("Waiting for obstruction to clear...")
		doorTimer.Reset(500 * time.Millisecond)
	} else {
		elev.CloseDoorAndResume()
	}
}

func fsmHandleObstruction(isObstructed bool, elev *elevator.Elevator) {
	elev.Obstruction = isObstructed

	if isObstructed {
		fmt.Println("Obstruction detected! Waiting for it to clear...")
	}
}

func fsmHandleEmergencyStop(a bool, elev *elevator.Elevator, number_of_floors int) {
	fmt.Printf("%+v\n", a)
	elevio.SetStopLamp(true)
	elevio.SetMotorDirection(elevio.MD_Stop)
	elev.Behavior = elevator.EB_Idle
	elev.Direction = elevio.MD_Stop
	elev.Orders = [4][3]elevator.OrderType{}

	for f := 0; f < number_of_floors; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}

	time.Sleep(5 * time.Second)
	elevio.SetStopLamp(false)
}
