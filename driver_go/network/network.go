package network

import (
	"Driver_go/elevator"
	"Driver_go/network/bcast"
	"Driver_go/network/peers"
)

func NetworkInit(id string) (
	chan peers.PeerUpdate,
	chan elevator.Elevator,
	chan elevator.Elevator,
	chan bool,
	chan bool,
) {
	// Channels for peer updates
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	// Channels for elevator state broadcasting
	elevStateTx := make(chan elevator.Elevator)
	elevStateRx := make(chan elevator.Elevator)

	// Channels for HRA (some custom communication)
	run_hra := make(chan bool, 10)
	receive_run_hra := make(chan bool, 10)

	// Start network-related goroutines
	go peers.Transmitter(15622, id, peerTxEnable)
	go peers.Receiver(15622, peerUpdateCh)

	go bcast.Transmitter(20456, elevStateTx)
	go bcast.Receiver(20456, elevStateRx)

	go bcast.Transmitter(20032, run_hra)
	go bcast.Receiver(20032, receive_run_hra)

	// Return all the created channels
	return peerUpdateCh, elevStateTx, elevStateRx, run_hra, receive_run_hra
}
