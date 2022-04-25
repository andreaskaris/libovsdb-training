package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os/exec"
	"strings"

	"example.com/libovs-training/nbdb"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
)

const (
	nbdbSocket = "127.0.0.1:39641"
)

// Make sure to use the latest version of libovsdb. More recently, the library has gone through
// quite a few changes. These examples are based on: v0.6.1-0.20220328142833-2cbe2d093e12
// go get github.com/ovn-org/libovsdb@main
// go install github.com/ovn-org/libovsdb/cmd/modelgen@main

// To better understand the Open vSwitch Database Management Protocol, read through the RFC:
// https://datatracker.ietf.org/doc/html/rfc7047.txt
// And also read through the most recent documentation which made some amendments to the RFC:
// https://docs.openvswitch.org/en/latest/ref/ovsdb-server.7/

// main contains example code that executes the following steps:
//   0) Load the Northbound Database's schema
//   1) Connect to the Northbound Database
//   2) Monitor all tables of this database, make ovsdbclient use its caching mechanism
//   3) Send an echo request to the server to verify its liveness (for illustration purposes, not needed)
//   4) Search all LogicalRouters with name 'myRouter'
//   a) Generate the 'create' operations necessary to create a new LogicalRouter
//   c) List all LogicalRouters again
//   5) Create a new LogicalRouter with name lrName if no results were found
//   5) a) Generate the 'create' operations necessary to create a new LogicalRouter
//   5) b) Apply the 'create' operations with a 'transact' request
//   5) c) List all LogicalRouters again
//   6) Print the content of the LogicalRouter table.
//   7) Update all LogicalRouters with name lrName and set the enabled field to true.
//   7) a) Create the update operation.
//   7) b) Transact applies the operations to the database.
//   8) Mutate all LogicalRouters with name lrName and add an option.
//   8) a) Create the mutation operation.
//   8) b) Transact applies the operations to the database.
//   9) Execute the ovn-nbctl shell command if it exists, for illustration purposes.
//   10) Delete all entries in the table that match lrName
//   11) Transact applies the operations to the database.
func main() {
	// 0)
	log.Println("")
	log.Println("0) Load the Northbound Database's schema")
	log.Println("===================")
	log.Println("")
	// Load the full database model.
	// The database schemas for fedora systems are part of the ovn-central RPM and
	// can be found in /usr/share/ovn/ovn-nb.ovsschema and /usr/share/ovn/ovn-sb.ovsschema
	// Models were then generated with:
	// modelgen -p nbdb -o nbdb /usr/share/ovn/ovn-nb.ovsschema
	// modelgen -p sbdb -o sbdb /usr/share/ovn/ovn-sb.ovsschema
	dbModelReq, err := nbdb.FullDatabaseModel()
	if err != nil {
		log.Fatal(err)
	}

	// 1)
	log.Println("")
	log.Println("1) Connect to the Northbound Database")
	log.Println("===================")
	log.Println("")
	// Create a new ovsdbclient object with the NBDB schema and connect to the
	// NorthboundDB. The NorthboundDB in this case listens on localhost:39641 and the connection
	// is not encrypted.
	ovsdbclient, err := client.NewOVSDBClient(dbModelReq, client.WithEndpoint("tcp:"+nbdbSocket))
	if err != nil {
		log.Fatal(err)
	}
	err = ovsdbclient.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 2)
	log.Println("")
	log.Println("2) Monitor all tables of this database, make ovsdbclient use its caching mechanism")
	log.Println("===================")
	log.Println("")
	// Monitor all tables of this database according to
	// https://datatracker.ietf.org/doc/html/rfc7047.txt#section-4.1.5.
	// Only needed if you want to use the built-in cache
	ovsdbclient.MonitorAll(context.TODO())

	// 3)
	log.Println("")
	log.Println("3) Send an echo request to the server to verify its liveness (for illustration purposes, not needed)")
	log.Println("===================")
	log.Println("")
	// Verify liveness of database connection.
	// See: https://datatracker.ietf.org/doc/html/rfc7047.txt#section-4.1.11
	err = ovsdbclient.Echo(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// 4)
	log.Println("")
	log.Println("4) Search all LogicalRouters with name 'myRouter'")
	log.Println("===================")
	log.Println("")
	// Search for all LogicalRouters with lr.Name == lrName
	lrName := "myRouter"
	logicalRouterResults := []nbdb.LogicalRouter{}
	err = ovsdbclient.WhereCache(func(lr *nbdb.LogicalRouter) bool {
		return lr.Name == lrName
	}).List(context.TODO(), &logicalRouterResults)
	if err != nil {
		log.Fatal(err)
	}

	// 5)
	log.Println("")
	log.Println("5) Create a new LogicalRouter with name lrName if no results were found")
	log.Println("===================")
	log.Println("")
	// Create a new LogicalRouter with name lrName if no results were found.
	if len(logicalRouterResults) == 0 {
		log.Printf("No routers found, creating one.")
		// 5) a)
		log.Println("")
		log.Println("5) a) Generate the 'create' operations necessary to create a new LogicalRouter")
		log.Println("===================")
		log.Println("")
		// Create a *LogicalRouter, as a pointer to a Model is required by the API
		lr := &nbdb.LogicalRouter{
			Name: lrName,
		}
		// Create will create the necessary operations to create
		// a new LR. In this case:
		// [{"op":"insert","table":"Logical_Router","row":{"name":"myRouter"}}]
		// See: https://datatracker.ietf.org/doc/html/rfc7047.txt#section-5
		ops, err := ovsdbclient.Create(lr)
		if err != nil {
			log.Fatal(err)
		}
		jsonOps, err := json.Marshal(ops)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Operations: %v", string(jsonOps))

		// 5) b)
		log.Println("")
		log.Println("5) b) Apply the 'create' operations with a 'transact' request")
		log.Println("===================")
		log.Println("")
		// Transact applies the operations to the database.
		// See: https://datatracker.ietf.org/doc/html/rfc7047.txt#section-4.1.3
		// It returns an array of operation results, in this case something like:
		// [{"uuid":["uuid","4e342c6e-3a01-40b4-9c7b-918c0aebc293"]}]
		res, err := ovsdbclient.Transact(context.TODO(), ops...)
		if err != nil {
			log.Fatal(err)
		}
		jsonOperationResults, err := json.Marshal(res)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Operation results: %v", string(jsonOperationResults))

		// 5) c)
		log.Println("")
		log.Println("5) c) List all LogicalRouters again")
		log.Println("===================")
		log.Println("")
		// List the LogicalRouter table contents again.
		err = ovsdbclient.WhereCache(func(lr *nbdb.LogicalRouter) bool {
			return lr.Name == "myRouter"
		}).List(context.TODO(), &logicalRouterResults)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 6)
	log.Println("")
	log.Println("6) Print the content of the LogicalRouter table.")
	log.Println("===================")
	log.Println("")
	// Print the content of the LogicalRouter table.
	jsonLogicalRouterResults, err := json.Marshal(logicalRouterResults)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Result: %v", string(jsonLogicalRouterResults))

	// 7)
	log.Println("")
	log.Println("7) Update all LogicalRouters with name lrName and set the enabled field to true.")
	log.Println("===================")
	log.Println("")
	// Update all LogicalRouters with name lrName and set the enabled field to true.
	// See: https://datatracker.ietf.org/doc/html/rfc7047.txt#section-5.2.3
	for _, lr := range logicalRouterResults {
		lr := lr
		log.Printf("Updating LogicalRouter: %v", lr)
		// See: https://www.ovn.org/support/dist-docs/ovn-nb.5.html
		enabled := false
		lr.Enabled = &enabled

		// 7) a)
		log.Println("")
		log.Println("7) a) Create the update operation.")
		log.Println("===================")
		log.Println("")
		// Create the update operation.
		// This will yield something like:
		// [{"op":"update","table":"Logical_Router","row":{"enabled":false},"where":[["_uuid","==",["uuid","c3c92eb7-0f61-44f1-a01c-740987a4de40"]]]}]
		ops, err := ovsdbclient.Where(&lr).Update(&lr, &lr.Enabled)
		if err != nil {
			log.Fatal(err)
		}
		jsonOps, err := json.Marshal(ops)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Operations: %v", string(jsonOps))

		// 7) b)
		log.Println("")
		log.Println("7) b) Transact applies the operations to the database.")
		log.Println("===================")
		log.Println("")
		// Transact applies the operations to the database.
		// See: https://datatracker.ietf.org/doc/html/rfc7047.txt#section-4.1.3
		// It returns an array of operation results, in this case something like:
		// [{"count":1,"uuid":["named-uuid",""]}
		res, err := ovsdbclient.Transact(context.TODO(), ops...)
		if err != nil {
			log.Fatal(err)
		}
		jsonOperationResults, err := json.Marshal(res)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Operation results: %v", string(jsonOperationResults))
	}

	// 8)
	log.Println("")
	log.Println("8) Mutate all LogicalRouters with name lrName and add an option.")
	log.Println("===================")
	log.Println("")
	// Mutate all LogicalRouters with name lrName and add an option.
	// See: https://datatracker.ietf.org/doc/html/rfc7047.txt#section-5.2.4
	for _, lr := range logicalRouterResults {
		log.Printf("Mutating LogicalRouter: %v", lr)
		// 8) a)
		log.Println("")
		log.Println("8) a) Create the mutation operation.")
		log.Println("===================")
		log.Println("")
		// Create the mutation operation.
		// This will yield something like:
		// [{"op":"mutate",
		//   "table":"Logical_Router",
		//   "mutations":[
		//     ["options","insert",["map",[["mcast_relay","true"]]]]],
		//         "where":[["_uuid","==",["uuid","481500fd-9622-421e-bec7-5ebe22c256ec"]]]
		// }]
		// For more info about the fields that can be mutated,
		// see: https://www.ovn.org/support/dist-docs/ovn-nb.5.html
		ops, err := ovsdbclient.Where(&lr).Mutate(&lr, model.Mutation{
			Field:   &lr.Options,
			Mutator: ovsdb.MutateOperationInsert,
			Value:   map[string]string{"mcast_relay": "true"},
		})
		if err != nil {
			log.Fatal(err)
		}
		jsonOps, err := json.Marshal(ops)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Operations: %v", string(jsonOps))

		// 8) b)
		log.Println("")
		log.Println("8) b) Transact applies the operations to the database.")
		log.Println("===================")
		log.Println("")
		// Transact applies the operations to the database.
		// See: https://datatracker.ietf.org/doc/html/rfc7047.txt#section-4.1.3
		// It returns an array of operation results, in this case something like:
		// [{"count":1,"uuid":["named-uuid",""]}]
		res, err := ovsdbclient.Transact(context.TODO(), ops...)
		if err != nil {
			log.Fatal(err)
		}
		jsonOperationResults, err := json.Marshal(res)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Operation results: %v", string(jsonOperationResults))
	}

	// 9)
	log.Println("")
	log.Println("9) Execute the ovn-nbctl shell command if it exists, for illustration purposes.")
	log.Println("===================")
	log.Println("")
	// Execute the ovn-nbctl shell command if it exists, for illustration purposes.
	path, err := exec.LookPath("ovn-nbctl")
	if err == nil {
		log.Printf("Running: ovn-nbctl list Logical_Router")
		cmd := exec.Command(path, "--db=tcp:"+nbdbSocket, "list", "Logical_Router")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(strings.NewReader(out.String()))
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}

	// 10)
	log.Println("")
	log.Println("10) Delete all entries in the table that match lrName")
	log.Println("===================")
	log.Println("")
	// Delete all entries in the table that match lrName
	// See: https://datatracker.ietf.org/doc/html/rfc7047.txt#section-5.2.5
	ops, err := ovsdbclient.WhereCache(func(lr *nbdb.LogicalRouter) bool {
		return lr.Name == "myRouter"
	}).Delete()
	if err != nil {
		log.Fatal(err)
	}
	jsonOps, err := json.Marshal(ops)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Operations: %v", string(jsonOps))

	// 11)
	log.Println("")
	log.Println("11) Transact applies the operations to the database.")
	log.Println("===================")
	log.Println("")
	// Transact applies the operations to the database.
	// See: https://datatracker.ietf.org/doc/html/rfc7047.txt#section-4.1.3
	// It returns an array of operation results, in this case something like:
	// [{"count":1,"uuid":["named-uuid",""]}]
	res, err := ovsdbclient.Transact(context.TODO(), ops...)
	if err != nil {
		log.Fatal(err)
	}
	jsonOperationResults, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Operation results: %v", string(jsonOperationResults))
}
