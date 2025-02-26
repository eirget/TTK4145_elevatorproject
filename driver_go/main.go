package main

import (
	"Driver_go/elevio"
	"Driver_go/network/bcast"
	"Driver_go/network/peers"
	"flag"
	"fmt"
	"strconv"
)

type HelloMsg struct {
	Message string
	Iter 	int
}


var NumFloors = 4
var NumButtons = 3

func main() {
	
	//NETWORK

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()


	//make a channel for receiving updates on the id's of the peers that are alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	//we can enable/disable the transmitter after it has been started, coulb be used to signal that we are unavailable
	peerTxEnable := make(chan bool)
	//we make channels for sending and receiving our custom data types
	ordersTx := make(chan [4][3]OrderType)
	ordersRx := make(chan [4][3]OrderType)

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	go bcast.Transmitter(20022, ordersTx)
	go bcast.Receiver(20022, ordersRx)

	elevio.Init("localhost:15657", NumFloors) //gjør til et flag

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

	for f := 0; f < NumFloors; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)

	a := <-drv_floors
	for a == -1 {
		elevio.SetMotorDirection(d)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)

	id_num, _ := strconv.Atoi(id)
	elevator := ElevatorInit(a, id_num) //kanskje det blir penere å bare bruke string

	elevio.SetFloorIndicator(elevator.Floor_nr)

	go fsm(elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, NumFloors)
	

	//the example message, we just send one of these every second
	/*
	go func() {
		for {
			ordersTx <- elevator.Orders
			time.Sleep(1 * time.Second)
		}
	}()
	*/

	fmt.Printf("Started elevator system")

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-ordersRx:  //kam hende denne vil miste orders om 
			fmt.Printf("Recieved: \n")
			fmt.Printf("%+v\n", a[0])
			fmt.Printf("%+v\n", a[1])
			fmt.Printf("%+v\n", a[2])
			fmt.Printf("%+v\n", a[3])
		}
	}


}
