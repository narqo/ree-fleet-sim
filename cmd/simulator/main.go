package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/narqo/ree-fleet-sim/internal/vehicle"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	if err := run(ctx, os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("", flag.ExitOnError)

	var (
		fleetStateAddr     string
		vehiclesTotal      int
		tickInterval       time.Duration
		maxDistancePerTick float64
	)
	flags.StringVar(&fleetStateAddr, "fleetstate-server-addr", "http://127.0.0.1:10080", "address of fleetstate server")
	flags.IntVar(&vehiclesTotal, "vehicles-total", 20, "total number of vehicles to simulate")
	flags.DurationVar(&tickInterval, "vehicle-tick-interval", time.Second, "interval a vehicle sends an update to fleetstate server")
	flags.Float64Var(&maxDistancePerTick, "vehicle-max-distance-per-tick", 13, "max distance in meters a vehicle moves per pick")

	if err := flags.Parse(args); err != nil {
		return err
	}

	client := vehicle.NewFleetStateClient(fleetStateAddr)

	var vcs []*vehicle.Vehicle
	for n := vehiclesTotal; n > 0; n-- {
		vcs = append(vcs, vehicle.NewVehicle(client))
	}

	var wg sync.WaitGroup
	for _, vc := range vcs {
		log.Printf("starting simulation for %s", vc)

		wg.Add(1)
		go func(vc *vehicle.Vehicle) {
			defer wg.Done()

			vc.ReportPosition(ctx)

			ticker := time.NewTicker(tickInterval)
			for {
				select {
				case <-ticker.C:
					vc.MoveNearby(rand.Float64() * maxDistancePerTick)
					vc.ReportPosition(ctx)
				case <-ctx.Done():
					return
				}
			}
		}(vc)
	}

	wg.Wait()

	log.Println("exiting...")

	return nil
}
