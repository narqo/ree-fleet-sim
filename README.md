# ree-fleet-sim

## Usage

### Run using docker-compose

```
$ docker-compose up --build

···
Starting ree-fleet-sim_simulator_1         ... done
Starting ree-fleet-sim_fleetstate-server_1 ... done
fleetstate-server_1  | 2020/10/06 06:32:07 listening: addr :10080
simulator_1          | 2020/10/06 06:32:07 starting simulation for Vehicle THE4432443819328064265VIN (-61.698146 -58.585985)
simulator_1          | 2020/10/06 06:32:07 starting simulation for Vehicle THE8559006319754845886VIN (-87.363958 -117.395366)
```

The command above builds docker image with with the server and simulator and stars two docker-compose
services — `fleetstate-server` and `simulator`.

### Build and run localy

*Expects Go 1.15*

**Build and run server**

```
$ go build ./cmd/fleetstate-server/
$ ./fleetstate-server

2020/10/06 08:42:31 listening: addr 127.0.0.1:10080
```

See `./fleetstate-server --help` for available options.

**Build and run simulator**

While server is running, in a separate terminal window, do the following:

```
$ go build ./cmd/simulator/
$ ./simulator

2020/10/06 08:44:20 starting simulation for Vehicle THE4898130556926864868VIN (47.085652 -177.346404)
···
```

See `./simulator --help` for available options.

### Consume the positions stream

`scripts/fleet-watch.sh` provides a client script to consume the positions for a single vehicle.
With both server and simulator running, do:

```
./scripts/fleet-watch.sh THE4898130556926864868VIN
{"lat":22.626401108688476,"lon":59.35106377261182,"speed":2.2669004403351765}
{"lat":22.626460306194144,"lon":59.35100080870115,"speed":33.17355036917585}
···
```

`./scripts/fleet-watch.sh --help` shows available options and an example.

## Project Structure and Overview

- `cmd/fleetstate-server` — Fleet State server's main package.
- `cmd/simulator` — Vehicles simulator's main package.
- `scripts/fleet-watch.sh` — A client script that watches a vehicle.

All supporting packakge for the server and simulator are in `internal/`.

### fleetstate-server

An HTTP server that listens for requests from either simulator or fleet-watcher.

Server provides the following HTTP API

**Update the lat-lon position for a vehicle `vin`**

```
POST /vehicle/<vin>
body lat=<lat>&lon<lon>
```

Server stores incoming positions in `Store`...

**Stream the lat-lon position for a vehicle `vin`**

```
GET /vehicle/<vin>/stream
```

### simulator

Generates