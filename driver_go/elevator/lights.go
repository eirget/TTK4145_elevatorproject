package elevator

import "time"

func (e *Elevator) LightUpdater() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		e.SetLights()
	}
}
