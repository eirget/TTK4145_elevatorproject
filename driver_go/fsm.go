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
			handleRequestButtonPress(a, elev, elevStateTx)

		case <-time.After(100 * time.Millisecond):
			handleIdleState(elev, doorTimer)

		case a := <-new_floor_chan:
			handleNewFloor(a, elev, elevStateTx, doorTimer)

		case <-doorTimer.C:
			handleDoorTimeout(elev, doorTimer)

		case isObstructed := <-obstr_chan:
			handleObstruction(isObstructed, elev, run_hra)

		case a := <-stop_chan:
			handleEmergencyStop(a, elev, number_of_floors)

		case <-time.After(1 * time.Second):
			elev.SetLights()
		}
	}
}

func handleRequestButtonPress(a elevio.ButtonEvent, elev *elevator.Elevator, elevStateTx chan elevator.Elevator) {
	fmt.Printf("%+v\n", a)
	elev.Orders[a.Floor][a.Button].State = true
	elev.Orders[a.Floor][a.Button].Timestamp = time.Now()

	fmt.Println("orders after button press: ", elev.Orders)

	if a.Button == elevio.BT_Cab {
		elevio.SetButtonLamp(a.Button, a.Floor, true)
	}
	new_order_flag = true
	elevStateTx <- *elev
}

func handleIdleState(elev *elevator.Elevator, doorTimer *time.Timer) {
	if elev.Behavior == elevator.EB_Idle {
		elev.HandleIdleState()
		if elev.Behavior == elevator.EB_DoorOpen {
			doorTimer.Reset(5 * time.Second)
		}
	}
}

func handleNewFloor(a int, elev *elevator.Elevator, elevStateTx chan elevator.Elevator, doorTimer *time.Timer) {
	elev.Floor_nr = a
	elevio.SetFloorIndicator(a)

	if elev.ShouldStop() {
		elev.StopAtFloor()
		doorTimer.Reset(3 * time.Second)
	}

	elevStateTx <- *elev
}

func handleDoorTimeout(elev *elevator.Elevator, doorTimer *time.Timer) {
	if elev.Obstruction {
		fmt.Println("Waiting for obstruction to clear...")
		doorTimer.Reset(500 * time.Millisecond)
	} else {
		elev.CloseDoorAndResume()
	}
}

func handleObstruction(isObstructed bool, elev *elevator.Elevator, runHra chan bool) {
	elev.Obstruction = isObstructed

	if isObstructed {
		fmt.Println("Obstruction detected! Waiting for it to clear...")

		// Start a timer in a separate goroutine
		go func() {
			time.Sleep(2 * time.Second) // If obstruction lasts more than 2 sec, reassign orders
			if elev.Obstruction {
				fmt.Println("Obstruction still active! Running hallRequestAssigner to redistribute hall orders...")
				runHra <- true // ✅ Instead of ReassignOrders(), trigger hallRequestAssigner()
			} else {
				fmt.Println("Obstruction cleared!")
			}
		}()
	}
}

func handleEmergencyStop(a bool, elev *elevator.Elevator, number_of_floors int) {
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
