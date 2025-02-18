package main

import (
	"Driver-go/elevio"
	"fmt"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	elevator := ElevatorInit()

	var d elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	a := <-drv_floors
	for a < 0 {
		elevio.SetMotorDirection(d)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)

	elevio.SetFloorIndicator(elevator.Floor_nr) //Kobler floor sensor med floor indicator light

	for f := 0; f < numFloors; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}

	fmt.Println("At first destination")
	fmt.Print(elevator.Queue)



	go SimpleFsm(elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, numFloors) //try to make this work instead
	go elevator.processQueue()

	select {}

	/*
	   for {
	       select {
	       case a := <- drv_buttons:
	           fmt.Printf("%+v\n", a)
	           elevio.SetButtonLamp(a.Button, a.Floor, true)

	       case a := <- drv_floors:
	           fmt.Printf("%+v\n", a)
	           if a == numFloors-1 {
	               d = elevio.MD_Down
	           } else if a == 0 {
	               d = elevio.MD_Up
	           }
	           //elevio.SetMotorDirection(d)


	       case a := <- drv_obstr:
	           fmt.Printf("%+v\n", a)
	           if a {
	               //elevio.SetMotorDirection(elevio.MD_Stop)
	           } else {
	               //elevio.SetMotorDirection(d)
	           }

	       case a := <- drv_stop:
	           fmt.Printf("%+v\n", a)
	           for f := 0; f < numFloors; f++ {
	               for b := elevio.ButtonType(0); b < 3; b++ {
	                   elevio.SetButtonLamp(b, f, false)
	               }
	           }
	       }
	   } */
}
