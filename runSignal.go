package main

import (
	"log"
	"time"
	"os/signal"
	"syscall"
	"os"
)

func main() {
	// Creating a channel to watch for signals.
	sig := make(chan os.Signal)
	// Channel to signal the goroutines to stop.
	work := make(chan bool)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	hostInterval := 15 * time.Second
	serviceInterval := 5 * time.Second


	// Running each function manually since we want the checks to fire off without waiting the first time. After the first time, the checks will fire off on schedule.
	log.Println("------ First Checks Start ------")
	go hostAlive()
	go serviceStatus()

	// Starting to fire off the checks at scheduled intervals.
	log.Println("------ Scheduled Checks Start ------")
	host := schedule(hostAlive, hostInterval, work)
	service := schedule(serviceStatus, serviceInterval, work)

	// Waiting for signal to finish executing the rest of the code.
	flag := <-sig
	log.Printf("Caught %s signal. Stopping loop...", flag)
	close(work)
	log.Println("------ Scheduled Checks End ------")
	host.Stop()
	service.Stop()
}

func schedule(check func(), interval time.Duration, work <-chan bool) *time.Ticker {
	clock := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-clock.C:
				check()
			case <-work:
				return
			}
		}
	}()
	return clock
}

func hostAlive() {
	log.Printf("I'm alive!\n")
}

func serviceStatus() {
	log.Printf("Service Check\n")
}
