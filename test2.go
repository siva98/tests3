

/*
MIT License
Copyright (c) 2018 Gökhan Koçak www.gokhankocak.com
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

const (
	// Version information
	Version = "1.0"
)

// SharedCounters are used to keep statistics for operations on shared data
type SharedCounters struct {
	GetState   int64 `json:"GetState"`
	PutState   int64 `json:"Putstate"`
	DelState   int64 `json:"DelState"`
	GetHistory int64 `json:"GetHistory"`
	RichQuery  int64 `json:"RichQuery"`
	Success    int64 `json:"Success"`
	Errors     int64 `json:"Errors"`
}

// Statistics for Shared data areas
type Statistics struct {
	Shared SharedCounters `json:"Shared"`
}

// KeyModification holds the modification history entries
type KeyModification struct {
	TxId      string `json:"TxId"`
	Value     []byte `json:"Value"`
	Timestamp int64  `json:"TimeStamp"` // UNIX timestamp as nanoseconds
	IsDelete  bool   `json:"IsDelete"`
}

// QueryResults holds the result of rich query
type QueryResult struct {
	Namespace string `json:"Namespace"`
	Key       string `json:"Key"`
	Value     []byte `json:"Value"`
}

// AsenaSmartContract is the Smart Contract structure
type AsenaSmartContract struct {
	Logger *shim.ChaincodeLogger
	Stats  Statistics
	Config AsenaConfig
}

// AsenaConfig is used to configure the Asena Smart Contract on demand
type AsenaConfig struct {
	LogLevel  string `json:"LogLevel"`
	StatsdUrl string `json:"StatsdUrl"`
}

// StatsdReporter reports statistics to the given statsd server
func (asc *AsenaSmartContract) StatsdReporter(stub shim.ChaincodeStubInterface) peer.Response {

	var ErrorCount int

	for {
		time.Sleep(1 * time.Second)
		cli, err := statsd.NewClient(asc.Config.StatsdUrl, "AsenaSmartContract")
		if err != nil {
			ErrorCount++
			if ErrorCount > 1000 {
				break
			}
			time.Sleep(15 * time.Second)
		} else {
			cli.SetInt("Shared.GetState", asc.Stats.Shared.GetState, 1.0)
			cli.SetInt("Shared.PutState", asc.Stats.Shared.PutState, 1.0)
			cli.SetInt("Shared.DelState", asc.Stats.Shared.DelState, 1.0)
			cli.SetInt("Shared.GetHistory", asc.Stats.Shared.GetHistory, 1.0)
			cli.SetInt("Shared.RichQuery", asc.Stats.Shared.RichQuery, 1.0)
			cli.SetInt("Shared.Success", asc.Stats.Shared.Success, 1.0)
			cli.SetInt("Shared.Errors", asc.Stats.Shared.Errors, 1.0)
			cli.Close()
		}
	}

	return shim.Success(nil)
}

// Log is used to log messages
func (asc *AsenaSmartContract) Log(Level shim.LoggingLevel, format string, args ...interface{}) {

	if asc.Logger == nil {
		asc.Logger = shim.NewLogger("AsenaSmartContract")
	}

	if asc.Logger != nil {

		switch Level {
		case shim.LogCritical:
			asc.Logger.Criticalf(format, args)
		case shim.LogError:
			asc.Logger.Errorf(format, args)
		case shim.LogWarning:
			asc.Logger.Warningf(format, args)
		case shim.LogNotice:
			asc.Logger.Noticef(format, args)
		case shim.LogInfo:
			asc.Logger.Infof(format, args)
		case shim.LogDebug:
			asc.Logger.Debugf(format, args)
		default:
			asc.Logger.Infof(format, args)
		}
	}
}

// Init method is called when the Asena Smart Contract is instantiated by the blockchain network
// Best practice is to have any Ledger initialization in separate function -- see initLedger()
func (asc *AsenaSmartContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// InitLedger is called when instantiating the chaincode
func (asc *AsenaSmartContract) InitLedger(stub shim.ChaincodeStubInterface) peer.Response {

	type FirstData struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	}

	asc.Config = AsenaConfig{LogLevel: "INFO", StatsdUrl: "telegraf.org1.com:8125"}
	ConfigAsBytes, _ := json.Marshal(asc.Config)

	List := []FirstData{
		{Key: "AsenaSmartContract.Status", Value: "initialized"},
		{Key: "AsenaSmartContract.Version", Value: Version},
		{Key: "AsenaSmartContract.Config", Value: string(ConfigAsBytes)},
	}

	for _, d := range List {
		stub.PutState(d.Key, []byte(d.Value))
	}

	go asc.StatsdReporter(stub)

	return shim.Success([]byte("AsenaSmartContract.InitLedger(): returning success"))
}

// GetVersion returns the Asena Smart Contract version
func (asc *AsenaSmartContract) GetVersion(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte(Version))
}

// GetStats returns the statistics
func (asc *AsenaSmartContract) GetStats(stub shim.ChaincodeStubInterface) peer.Response {

	ResultAsBytes, err := json.Marshal(asc.Stats)
	if err != nil {
		return shim.Error("AsenaSmartContract.GetStats(): json.Marshal() failed: " + err.Error())
	}

	return shim.Success([]byte(ResultAsBytes))
}

// Invoke method is called as a result of an application request to run the Asena Smart Contract
// The calling application program has also specified the particular smart contract function to be called, with arguments
func (asc *AsenaSmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately

	asc.Log(shim.LogDebug, "Invoke(): called with function: %s and %d args", function, len(args))

	switch function {

	case "SetAsenaConfig":
		return asc.SetAsenaConfig(stub, args)
	case "GetAsenaConfig":
		return asc.GetAsenaConfig(stub, args)
	case "InitLedger":
		return asc.InitLedger(stub)
	case "GetVersion":
		return asc.GetVersion(stub)
	case "GetStats":
		return asc.GetStats(stub)
	case "GetState":
		return asc.GetState(stub, args)
	case "PutState":
		return asc.PutState(stub, args)
	case "DelState":
		return asc.DelState(stub, args)
	case "GetHistory":
		return asc.GetHistory(stub, args)
	case "GetQueryResult":
		return asc.GetQueryResult(stub, args)
	}

	return shim.Error("AsenaSmartContract.Invoke(): Invalid Smart Contract function name: " + function)
}

// main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Asena Smart Contract
	err := shim.Start(new(AsenaSmartContract))
	if err != nil {
		fmt.Println("Error creating new Asena Smart Contract:", err.Error())
	}
}

// SetAsenaConfig is used to configure Asena Smart Contract
func (asc *AsenaSmartContract) SetAsenaConfig(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("SetAsenaConfig(): expecting 1 argument")
	}

	err := json.Unmarshal([]byte(args[0]), &asc.Config)
	if err != nil {
		return shim.Error("SetAsenaConfig(): json.Unmarshal() failed: " + err.Error())
	}

	return shim.Success(nil)
}

// GetAsenaConfig is used to get the configuration of  Asena Smart Contract
func (asc *AsenaSmartContract) GetAsenaConfig(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	ValueAsBytes, err := json.Marshal(asc.Config)
	if err != nil {
		return shim.Error("SetAsenaConfig(): json.Marshal() failed: " + err.Error())
	}

	return shim.Success(ValueAsBytes)
}

// GetState returns the value of the specified `key` from the
// ledger. Note that GetState doesn't read data from the writeset, which
// has not been committed to the ledger. In other words, GetState doesn't
// consider data modified by PutState that has not been committed.
// If the key does not exist in the state database, (nil, nil) is returned.
// key := args[0]
func (asc *AsenaSmartContract) GetState(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	asc.Stats.Shared.GetState++
	if len(args) < 1 {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.GetState(): expecting at least 1 argument")
	}

	asc.Log(shim.LogDebug, "AsenaSmartContract.GetState(): called with argument: %s", args[0])

	ResultAsBytes, err := stub.GetState(args[0])
	if err != nil {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.GetState(): stub.GetState() failed: " + err.Error())
	}

	asc.Log(shim.LogDebug, "AsenaSmartContract.GetState(): returning success")

	asc.Stats.Shared.Success++
	return shim.Success(ResultAsBytes)
}

// PutState puts the specified `key` and `value` into the transaction's
// writeset as a data-write proposal. PutState doesn't effect the ledger
// until the transaction is validated and successfully committed.
// Simple keys must not be an empty string and must not start with null
// character (0x00), in order to avoid range query collisions with
// composite keys, which internally get prefixed with 0x00 as composite
// key namespace.
// key := args[0]
// value := args[1]
func (asc *AsenaSmartContract) PutState(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	asc.Stats.Shared.PutState++
	if len(args) != 2 {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.PutState(): expecting 2 arguments")
	}

	asc.Log(shim.LogDebug, "AsenaSmartContract.PutState(): called with arguments: %s %s %s", args[0], args[1])

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(args[1]), &m)
	if err != nil {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.PutState(): json.Unmarshal() failed: " + err.Error())
	}

	ValueAsBytes, err := json.Marshal(m)
	if err != nil {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.PutState(): json.Marshal() failed: " + err.Error())
	}

	err = stub.PutState(args[0], ValueAsBytes)
	if err != nil {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.PutState(): stub.PutState() failed: " + err.Error())
	}

	asc.Log(shim.LogDebug, "AsenaSmartContract.PutState(): returning success")

	asc.Stats.Shared.Success++
	return shim.Success([]byte(args[0])) // return the key
}

// DelState records the specified `key` to be deleted in the writeset of
// the transaction proposal. The `key` and its value will be deleted from
// the ledger when the transaction is validated and successfully committed.
// key := args[0]
func (asc *AsenaSmartContract) DelState(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	asc.Stats.Shared.DelState++
	if len(args) != 1 {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.DelState(): expecting 1 argument")
	}

	err := stub.DelState(args[0])
	if err != nil {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.DelState(): stub.DelState() failed: " + err.Error())
	}

	asc.Stats.Shared.Success++
	return shim.Success([]byte(args[0])) // return the key
}

// GetHistory returns a history of key values across time.
// For each historic key update, the historic value and associated
// transaction id and timestamp are returned. The timestamp is the
// timestamp provided by the client in the proposal header.
// GetHistoryForKey requires peer configuration
// core.ledger.history.enableHistoryDatabase to be true.
// The query is NOT re-executed during validation phase, phantom reads are
// not detected. That is, other committed transactions may have updated
// the key concurrently, impacting the result set, and this would not be
// detected at validation/commit time. Applications susceptible to this
// should therefore not use GetHistoryForKey as part of transactions that
// update ledger, and should limit use to read-only chaincode operations.
// key := args[0]
func (asc *AsenaSmartContract) GetHistory(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	asc.Stats.Shared.GetHistory++
	if len(args) != 1 {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.GetHistory(): expecting 1 argument")
	}

	HistoryIterator, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.GetHistory(): stub.GetHistoryForKey() failed: " + err.Error())
	}
	defer HistoryIterator.Close()

	var ModificationList []KeyModification

	for {
		if HistoryIterator.HasNext() {
			History, err := HistoryIterator.Next()
			if err != nil {
				asc.Stats.Shared.Errors++
				return shim.Error("AsenaSmartContract.GetHistory(): stub.HistoryIterator.Next() failed: " + err.Error())
			}
			Modification := new(KeyModification)
			Modification.TxId = History.GetTxId()
			Value := History.GetValue()
			Modification.Value = make([]byte, len(Value))
			copy(Modification.Value, Value)
			Modification.Timestamp = 1000000000*History.Timestamp.GetSeconds() + int64(History.Timestamp.GetNanos())
			Modification.IsDelete = History.GetIsDelete()
			ModificationList = append(ModificationList, *Modification)
		} else {
			break
		}
	}

	ResultAsBytes, err := json.Marshal(ModificationList)
	if err != nil {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.GetHistory(): json.Marshal() failed: " + err.Error())
	}

	asc.Stats.Shared.Success++
	return shim.Success(ResultAsBytes)
}

// GetQueryResult performs a "rich" query against a state database. It is
// only supported for state databases that support rich query,
// e.g.CouchDB. The query string is in the native syntax
// of the underlying state database. An iterator is returned
// which can be used to iterate over all keys in the query result set.
// However, if the number of keys in the query result set is greater than the
// totalQueryLimit (defined in core.yaml), this iterator cannot be used
// to fetch all keys in the query result set (results will be limited by
// the totalQueryLimit).
// The query is NOT re-executed during validation phase, phantom reads are
// not detected. That is, other committed transactions may have added,
// updated, or removed keys that impact the result set, and this would not
// be detected at validation/commit time.  Applications susceptible to this
// should therefore not use GetQueryResult as part of transactions that update
// ledger, and should limit use to read-only chaincode operations.
func (asc *AsenaSmartContract) GetQueryResult(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var ResultList []QueryResult

	asc.Stats.Shared.RichQuery++
	if len(args) != 1 {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.GetQueryResult(): expecting 1 argument")
	}

	QueryIterator, err := stub.GetQueryResult(args[0])
	if err != nil {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.GetQueryResult(): stub.GetQueryResult() failed: " + err.Error())
	}
	defer QueryIterator.Close()

	for {
		if QueryIterator.HasNext() {
			Query, err := QueryIterator.Next()
			if err != nil {
				asc.Stats.Shared.Errors++
				return shim.Error("AsenaSmartContract.GetQueryResult(): QueryIterator.Next() failed: " + err.Error())
			}
			Result := new(QueryResult)
			Result.Namespace = Query.GetNamespace()
			Result.Key = Query.GetKey()
			Value := Query.GetValue()
			Result.Value = make([]byte, len(Value))
			copy(Result.Value, Value)
			ResultList = append(ResultList, *Result)
		} else {
			break
		}
	}

	ResultAsBytes, err := json.Marshal(ResultList)
	if err != nil {
		asc.Stats.Shared.Errors++
		return shim.Error("AsenaSmartContract.GetQueryResult(): json.Marshal() failed: " + err.Error())
	}

	asc.Stats.Shared.Success++
	return shim.Success(ResultAsBytes)
}

