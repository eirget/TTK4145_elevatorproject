package elevator

import (
	"fmt"
	"time"
)

func MonitorActivity(e *Elevator, runHraCh chan<- struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if time.Since(e.LastActive) > 7*time.Second {
			fmt.Println("I have been inactive")
			if e.HasPendingHallOrders() {
				fmt.Println("And I have pending orders, calling hall request assigner")
				select {
				case runHraCh <- struct{}{}: // Trigger hall request reassignment
				default:

				}
			}
		}
	}
}
