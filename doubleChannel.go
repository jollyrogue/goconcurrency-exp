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
	// Channel to for the output function.
	output := make(chan string, 16)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	printInterval := 100 * time.Millisecond
	hostInterval := 15 * time.Second
	serviceInterval := 5 * time.Second

	// Starting the schedule for the output function.
	out := outSchedule(printOutput, printInterval, work, output)
	// Running each function manually since we want the checks to fire off without waiting the first time. After the first time, the checks will fire off on schedule.
	log.Println("------ First Checks Start ------")
	go hostAlive(output)
	go serviceStatus(output)

	// Starting to fire off the checks at scheduled intervals.
	log.Println("------ Scheduled Checks Start ------")
	host := checkSchedule(hostAlive, hostInterval, work, output)
	service := checkSchedule(serviceStatus, serviceInterval, work, output)

	// Waiting for signal to finish executing the rest of the code.
	flag := <-sig
	log.Printf("Caught %s signal. Stopping loop...", flag)
	close(work)
	log.Println("------ Scheduled Checks End ------")
	host.Stop()
	service.Stop()
	out.Stop()
}

func outSchedule(
	sendOut func(<-chan string),
	interval time.Duration,
	work <-chan bool,
	output <-chan string,
) (clock *time.Ticker) {
	clock = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-clock.C:
				sendOut(output)
			case <-work:
				return
			}
		}
	}()
	return clock
}

func checkSchedule(
	check func(chan<- string),
	interval time.Duration,
	work <-chan bool,
	output chan<- string,
) *time.Ticker {
	clock := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-clock.C:
				check(output)
			case <-work:
				return
			}
		}
	}()
	return clock
}

func hostAlive(output chan<- string) {
	output <- "I'm alive!"
}

func serviceStatus(output chan<- string) {
	output <- "Service is OK."
}

func printOutput(output <-chan string) {
	select {
	case result := <-output:
		log.Printf(result)
	default:
		return
	}
}
