# ree-fleet-sim

## Usage

### Run using docker-compose

*Expects Docker engine, with the support for multi-stage builds (v17.x+)*

```
$ docker-compose up --build

···
Starting ree-fleet-sim_simulator_1         ... done
Starting ree-fleet-sim_fleetstate-server_1 ... done
fleetstate-server_1  | 2020/10/06 06:32:07 listening: addr :10080
simulator_1          | 2020/10/06 06:32:07 starting simulation for Vehicle THE4432443819328064265VIN (-61.698146 -58.585985)
···
```

The command above builds docker image with the server and simulator and stars them as two docker services.

### Build and run yourself

*Expects Go 1.15+*

**Build and start server**

```
$ go build ./cmd/fleetstate-server/
$ ./fleetstate-server

2020/10/06 08:42:31 listening: addr 127.0.0.1:10080
```

See `./fleetstate-server --help` for available options.

**Build and start simulator**

While server is running, in a separate terminal window do:

```
$ go build ./cmd/simulator/
$ ./simulator

2020/10/06 08:44:20 starting simulation for Vehicle THE4898130556926864868VIN (xxx xxx)
···
```

See `./simulator --help` for available options.

### Consume the positions stream

`scripts/fleet-watch.sh` is a client script, that consumes the positions for a single vehicle.
With both server and simulator running, take of the vins reported by `simulator` and call:

```
./scripts/fleet-watch.sh THE4898130556926864868VIN
{"lat":22.626401108688476,"lon":59.35106377261182,"speed":2.2669004403351765}
{"lat":22.626460306194144,"lon":59.35100080870115,"speed":33.17355036917585}
···
```

`./scripts/fleet-watch.sh --help` shows available options.

### Run tests

```
$ go test ./...
```

## Project Structure and Overview

- `cmd/fleetstate-server` — Fleet State server's main package.
- `cmd/simulator` — Vehicles simulator's main package.
- `scripts/fleet-watch.sh` — A client script that watches a vehicle.

All supporting packages for the server and simulator are in `internal/`.

### fleetstate-server

`fleetstate-server` is an HTTP server that listens for incoming requests from either simulator or a client.

Server stores incoming positions in `Store`. For the demo solution, the only implementation of `Store` is an in-memory,
append-only storage, that keeps the incoming stream in the application's main memory.

Server does a high-level validation of the incoming request before storing the data. In case of an invalid request,
server returns HTTP 500, with the error description.

Server provides the following HTTP API:

**Update the lat-lon position for a vehicle `vin`**

```
POST /vehicle/<vin>
body lat=<lat>&lon<lon>

< 201 Created
```

**Stream the lat-lon position for a vehicle `vin`**

```
GET /vehicle/<vin>/stream
```

### simulator

`simulator` generates N vehicles, identified by a random VIN, in a random location.

Every second (`tick`), each vehicle "moves" in a random direction within a configurable radius (`max-distance-per-tick`),
and reports the new position to the server.
 
If the server can't be reached, the vehicle logs the error to stdout.

## Follow-up Questions

### 1\. How precise speed calculation, how to improve it

Server calculates the current speed by using the distance a vehicle moved during two subsequent updates.

The distance is calculated using a variant of [Haversine formula](https://en.wikipedia.org/wiki/Haversine_formula).
The formula sees the Earth as a perfect sphere with a radius R, and can reportedly, produce the results with up to 0.5% error.

The time delta between two subsequent updates is used to calculate the speed. Server records server's own time when stores the update. 
That ignores the latency between the vehicle and the server.

A vehicle could report a time-series `<ts> <lat>-<lon>`, rather than lat-lon pair only, to make the calculations more precise.

### 2\. How to make it reliable

Replace the in-memory storage with a persistent one. Because the data flow is heavily append-only,
[Redis Streams](https://redis.io/topics/streams-intro), [NATS](http://nats.io/), [Kafka](https://kafka.apache.org/), and similar,
will work fine.

The switch to a new storage is the matter of implementing `internal/fleetstate.Store` interface and injecting the new implementation
to server's request handler.

In addition to reporting the full time-series, a vehicle could keep a buffer of N previous data points. If didn't manage
to reach the server during several ticks, it could re-try sending the accumulated data from the buffer later. This will require
the support for back filling the old data on the server.

### 3\. How would you scale your solution on 200k vehicles, 20 cities 

From the server side:

Since every vehicle is identified by a single VIN, which is the part of the request's URL. And assuming,
a vehicle doesn't (usually) travel across regions, both server and the storage it connects to can be sharded (*geo-sharded if necessary*) by VIN.
An L7 load-balancer in front of the server instances will route the requests for a particular vin to a particular shard or group of shards.

From the simulator side:

Run multiple instances of simulator, storing the generated VINs in a shared data store, e.g. Redis's Hash Set,
to guaranty the unique VINs.
