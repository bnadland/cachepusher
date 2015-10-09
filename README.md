# cachepusher

1. Split app into backend / frontend
2. Use cachepusher
3. ???
4. Profit

## Setup

### Without local go installation

* `git clone https://github.com/bnadland/cachepusher`
* `cd cachepusher`
* `vagrant up`
* `vagrant ssh -c "tail -f /var/log/cachepusher/*"`

### With local go installation

* `go get github.com/bnadland/cachepusher/...`
* `cd $GOPATH/src/github.com/bnadland/cachepusher`
* `vagrant up`
* `vagrant ssh -c "sudo supervisorctl stop cp"`
* `go run syncer/main.go`

## How it works

There is a customer and a address table in postgresql. There is a python script
fakedata.py that simulates a backend and inserts some customers and addresses
into postgresql. There is database trigger on both tables, so whenever one is
updated a event on the `customer_updated` channel is fired with the id_customer
as payload. If a customer row gets deleted the trigger fires an event on
`customer_deleted` also with the id_customer as a payload.

The cachepusher program (code in ./syncer/main.go) opens a connection to
postgresql and listens on those two channels.

The denormalized data in redis uses the id_customer as a cachekey (i.e.
`customer:1` for id_customer=1) and contains a json datastructure with the
customer data and embedded into that the corresponding addresses.

Whenever a `customer_deleted` event is fired, cachepusher deletes the key from
redis.

When a `customer_updated` event is fired, cachepusher calls a stored procedure
with the id_customer from the payload and stores the returned json directly
into redis.


There is also a cache warmup functionality that gets called when the
cachepusher program starts up: It starts with deleting all keys under it's
control (i.e. `customer:*`), then sets up the listener to get updates and the
calls a stored procedure that 'touches' all customer records by inserting a new
timestamp into the touched_at column which then gets propagated to the redis
instance via the normal listener functionality described above.

## License

MIT License
