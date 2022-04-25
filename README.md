## Prerequisites 

This requires a running OVN Northbound database (and future examples potentially a Southbound
database as well):
~~~
ovsdb-tool create /etc/ovn/ovnsb_db.db /usr/share/ovn/ovn-sb.ovsschema
ovsdb-tool create /etc/ovn/ovnnb_db.db /usr/share/ovn/ovn-nb.ovsschema
ovsdb-server -vconsole:info -vfile:off --log-file=/var/log/ovn/ovsdb-server-sb.log --remote=punix:/var/run/ovn/ovnsb_db.sock --pidfile=/var/run/ovn/ovnsb_db.pid --unixctl=/var/run/ovn/ovnsb_db.ctl --remote=ptcp:39642 /etc/ovn/ovns    b_db.db
ovsdb-server -vconsole:info -vfile:off --log-file=/var/log/ovn/ovsdb-server-nb.log --remote=punix:/var/run/ovn/ovnnb_db.sock --pidfile=/var/run/ovn/ovnnb_db.pid --unixctl=/var/run/ovn/ovnnb_db.ctl --remote=ptcp:39641 /etc/ovn/ovnn    b_db.db
~~~

Make sure to use the latest version of libovsdb. More recently, the library has gone through
quite a few changes. These examples are based on: v0.6.1-0.20220328142833-2cbe2d093e12
~~~
go get github.com/ovn-org/libovsdb@main
go install github.com/ovn-org/libovsdb/cmd/modelgen@main
~~~

## Open vSwitch Database Management Protocol - Specification

To better understand the Open vSwitch Database Management Protocol, read through the RFC:
https://datatracker.ietf.org/doc/html/rfc7047.txt

And also read through the most recent documentation which made some amendments to the RFC:
https://docs.openvswitch.org/en/latest/ref/ovsdb-server.7/

## main.go

main contains example code that executes the following steps:
  0) Load the Northbound Database's schema
  1) Connect to the Northbound Database
  2) Monitor all tables of this database, make ovsdbclient use its caching mechanism
  3) Send an echo request to the server to verify its liveness (for illustration purposes, not needed)
  4) Search all LogicalRouters with name 'myRouter'
  a) Generate the 'create' operations necessary to create a new LogicalRouter
  c) List all LogicalRouters again
  5) Create a new LogicalRouter with name lrName if no results were found
  5) a) Generate the 'create' operations necessary to create a new LogicalRouter
  5) b) Apply the 'create' operations with a 'transact' request
  5) c) List all LogicalRouters again
  6) Print the content of the LogicalRouter table.
  7) Update all LogicalRouters with name lrName and set the enabled field to true.
  7) a) Create the update operation.
  7) b) Transact applies the operations to the database.
  8) Mutate all LogicalRouters with name lrName and add an option.
  8) a) Create the mutation operation.
  8) b) Transact applies the operations to the database.
  9) Execute the ovn-nbctl shell command if it exists, for illustration purposes.
  10) Delete all entries in the table that match lrName
  11) Transact applies the operations to the database.

## Running the example code

Simply use `make`:
~~~
make
~~~
