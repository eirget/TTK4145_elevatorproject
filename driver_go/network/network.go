package network

import (
	"Driver_go/elevator"
	"Driver_go/network/bcast"
	"Driver_go/network/peers"
)

func NetworkInit(id string,
	peerUpdateCh chan peers.PeerUpdate,
	peerTxEnable chan bool,
	elevStateTx chan elevator.Elevator,
	elevStateRx chan elevator.Elevator,
	runHra chan struct{}, receiveRunHra chan struct{}) {

	// Start network-related goroutines
	go peers.Transmitter(15622, id, peerTxEnable)
	go peers.Receiver(15622, peerUpdateCh)

	go bcast.Transmitter(20456, elevStateTx)
	go bcast.Receiver(20456, elevStateRx)

	go bcast.Transmitter(20032, runHra)
	go bcast.Receiver(20032, receiveRunHra)

}
