package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	// "reflect"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("ExchainChaincode")

var TICKETID = 0

// declare Lob_Name array to translate LoBID(int) into corresponding string name
const NumberOfLoBs = 8

var Lob_Name = [NumberOfLoBs]string{"MD_office", "HANA", "SMB", "IBS", "S4_HANA", "GS", "SF", "IoT"}

//Enum of LOBs
const (
	MD_office = iota
	HANA
	SMB
	IBS
	S4_HANA
	GS
	SF
	IoT
	//numberOfLoBs
)

const (
	P0 = iota
	P1
)

//TicketID status
//Created   ->  Applied  -> Ongoing  ->   Done
const (
	Created = iota
	Applied
	Ongoing
	Done
)

//Participant information
//   UserID:        iXXXXXX
//   UserName:      Bill Xu
//   Password:      *********
//   IsAdmin:       True or False
//   LoB:           0. MD_office  1. HANA  2. SMB...
type Participant struct {
	UserID   string `json:"Participant_UserID"`
	UserName string `json:"Participant_UserName"`
	Password string `json:"Participant_Password"`

	IsAdmin bool `json:"Participant_IsAdmin"`
	LoBID   int  `json:"Participant_LoBID"`
}

//Credit infomation
//UserID:     iXXXXXX
//Value:      123
//TicketIDs:  1. TicketNumber
//            2. Ticket array
type Credit struct {
	UserID string `json:"Credit_UserID"`
	Value  int    `json:"Credit_Value"`

	TicketIDs []string `json:"Credit_TicketIDs"`
}

// add methods to compare struct Credit based on Credit.Value
type Credits []Credit

func (s Credits) Len() int { // 重写 Len() 方法
	return len(s)
}
func (s Credits) Swap(i, j int) { // 重写 Swap() 方法
	s[i], s[j] = s[j], s[i]
}
func (s Credits) Less(i, j int) bool { // 重写 Less() 方法
	return s[i].Value > s[j].Value
}

type LoB struct {
	LoBID       int `json:"LoB_LoBID"`
	TotalCredit int `json:"LoB_TotalCredit"`

	UserIDs []string `json:"LoB_UserIDs"`
}

// Ticket information
//TicketID:
//Status:

//Title:
//Type:

//Value:
//UserID:

//DeadLine:
//Comment:
//Policy:

type Ticket struct {
	TicketID string `json:"Ticket_TicketID"`
	Status   int    `json:"Ticket_Status"`

	Title string `json:"Ticket_Title"`
	Type  int    `json:"Ticket_Type"`

	Value  int    `json:"Ticket_Value"`
	UserID string `json:"Ticket_UserID"`

	DeadLine time.Time `json:"Ticket_Deadline"`
	Comment  string    `json:"Ticket_Comment"`
	Policy   string    `json:"Ticket_Policy"`
}

// Order information
// TicketID:
// UserID:           iXXXXXX
// Status:           Created   ->  Applied  -> Ongoing  ->   Done
type Order struct {
	TicketID string `json:"TicketID"`
	UserID   string `json:"UserID"`
	Status   int    `json:"Status"`
}

//SmartContract - Chaincode for asset Reading
type SmartContract struct {
}

//ReadingIDIndex - Index on IDs for retrieval all Readings
type ReadingIDIndex struct {
	UserIDs []string `json:"UserIDs"`
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error starting Exchain chaincode function main(): %s", err)
	} else {
		fmt.Printf("Starting Exchain chaincode function main() executed successfully")
	}
}

//Init - The chaincode Init function: No arguments, only initializes a ID array as Index for retrieval of all Readings
func (rdg *SmartContract) Init(stub shim.ChaincodeStubInterface) peer.Response {

	//UseIDs, different LoB info and TICKETID are persistent

	var readingIDs ReadingIDIndex
	bytes, _ := stub.GetState("readingIDIndex")
	if len(bytes) == 0 {
		bytes, _ := json.Marshal(readingIDs)
		stub.PutState("readingIDIndex", bytes)
	} else {
		stub.PutState("readingIDIndex", bytes)
	}
	logger.Info("Func------Init----Get readingIDIndex" + string(bytes))

	var LobTemp LoB
	iter := 0
	for iter < NumberOfLoBs {
		bytes, _ := stub.GetState(Lob_Name[iter])
		if len(bytes) == 0 {
			LobTemp.LoBID = iter
			LobTemp.TotalCredit = 0
			bytes, _ := json.Marshal(LobTemp)
			stub.PutState(Lob_Name[iter], bytes)
		} else {
			stub.PutState(Lob_Name[iter], bytes)
		}
		logger.Info("Func------Init----Get LoB info" + string(bytes))
		iter = iter + 1
	}

	indexbytes, _ := stub.GetState("TICKETID")
	if len(indexbytes) == 0 {
		indexbytes, _ := json.Marshal(TICKETID)
		stub.PutState("TICKETID", indexbytes)
	} else {
		stub.PutState("TICKETID", indexbytes)
	}
	logger.Info("Func------Init----Get TICKETID" + string(indexbytes))

	return shim.Success(nil)
}

//Invoke - The chaincode Invoke function:
func (rdg *SmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Info(" ****** Invoke: function: ", function)

	switch function {
	//Participant Read Delete Update Add
	case "addParticipant":
		return rdg.addParticipant(stub, args)
	case "readParticipant":
		return rdg.readParticipant(stub, args[0])
	case "readAllParticipant":
		return rdg.readAllParticipant(stub)
	case "updateParticipant":
		return rdg.updateParticipant(stub, args)
	case "deleteParticipant":
		return rdg.deleteParticipant(stub, args[0])

	//Credit Read Delete Update Add
	case "CreditCreate":
		return rdg.CreditCreate(stub, args)
	case "CreditRead":
		return rdg.CreditRead(stub, args[0])
	case "CreditAdd":
		return rdg.CreditAdd(stub, args)
	case "CreditDelete":
		return rdg.CreditDelete(stub, args[0])
	case "TopTenCredit":
		return rdg.TopTenCredit(stub)

	// Lob Read
	case "LoBReadAll":
		return rdg.LoBReadAll(stub)
	case "LoBRead":
		return rdg.LoBRead(stub, args[0])

	//Ticket Read Delete Update Add
	case "TicketCreate":
		return rdg.TicketCreate(stub, args)
	case "TicketRead":
		return rdg.TicketRead(stub, args[0])
	case "TicketRead2":
		return rdg.TicketRead2(stub)
	case "TicketUpdate":
		return rdg.TicketUpdate(stub, args)
	case "AutoUpdateTicketStatus":
		return rdg.AutoUpdateTicketStatus(stub, args[0])
	case "TicketDelete":
		return rdg.TicketDelete(stub, args[0])

	//Order Read Delete Update Add
	case "OrderCreate":
		return rdg.OrderCreate(stub, args)
	//Read Single  It must be changed
	case "OrderRead":
		return rdg.OrderRead(stub, args)
	//Read All  It must be changed
	case "OrderRead2":
		return rdg.OrderRead2(stub, args)
	case "OrderUpdate":
		return rdg.OrderUpdate(stub, args)

		//Get HistoryTicket
	case "history":
		return rdg.TestGetHistoryTicket(stub, args)
	default:
		logger.Error("Received unknown function invocation: ", function)
	}
	return shim.Error("Received unknown function invocation")
}

//getReadingFromArgs - construct a reading structure from string array of arguments
func getParticipantFromArgs(args []string) (participant Participant, err error) {
	//check inputs!
	//  json:"Participant_UserID"
	//  json:"Participant_UserName"
	//  json:"Participant_Password"
	//  json:"Participant_IsAdmin"
	//  json:"Participant_LoB"
	if strings.Contains(args[0], "\"Participant_UserName\"") == false ||
		strings.Contains(args[0], "\"Participant_UserID\"") == false ||
		strings.Contains(args[0], "\"Participant_Password\"") == false ||
		strings.Contains(args[0], "\"Participant_IsAdmin\"") == false ||
		strings.Contains(args[0], "\"Participant_LoBID\"") == false {
		return participant, errors.New("Unknown field: Input JSON does not comply to schema")
	}

	err = json.Unmarshal([]byte(args[0]), &participant)
	if err != nil {
		return participant, err
	}
	return participant, nil
}

//Invoke Route: addNewReading
func (rdg *SmartContract) addParticipant(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//get Participant
	participant, err := getParticipantFromArgs(args)
	logger.Info("Func------addParticipant----Participant.LoBID" + string(participant.LoBID))

	if err != nil {
		return shim.Error("Reading participant is Corrupted")
	}
	//check Participant exists or not
	record, err := stub.GetState(participant.UserID)
	if record != nil {
		return shim.Error("This participant already exists: " + participant.UserID)
	}

	//if not exists, save
	participantAsBytes, err := rdg.saveParticipant(stub, participant)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = CreditInit(stub, participant.UserID, 0)
	if err != nil {
		return shim.Error(err.Error())
	}

	// updata LoB UserIDs array
	_, err = rdg.updateLoBUsers(stub, participant)
	if err != nil {
		return shim.Error(err.Error())
	}

	// update the ID index of
	_, err = rdg.updateReadingIDIndex(stub, participant)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(participantAsBytes)
}

//Helper: Save purchaser
func (rdg *SmartContract) saveParticipant(stub shim.ChaincodeStubInterface, participant Participant) ([]byte, error) {
	bytes, err := json.Marshal(participant)
	if err != nil {
		return bytes, errors.New("Error converting reading record JSON")
	}
	err = stub.PutState(participant.UserID, bytes)
	if err != nil {
		return bytes, errors.New("Error storing Reading record")
	}
	return bytes, nil
}

//Helper: add user to its LoB's UserIDs array
func (rdg *SmartContract) updateLoBUsers(stub shim.ChaincodeStubInterface, participant Participant) (bool, error) {

	var LoB_temp LoB

	LobName := Lob_Name[participant.LoBID]
	bytes, err := stub.GetState(LobName)
	if err != nil {
		return false, errors.New("updateLoBUsers: Error getting LoB info from state")
	}

	err = json.Unmarshal(bytes, &LoB_temp)
	if err != nil {
		return false, errors.New("updateLoBUsers: Error unmarshalling LoB JSON")
	}

	// To do: participant credit
	LoB_temp.LoBID = participant.LoBID
	LoB_temp.UserIDs = append(LoB_temp.UserIDs, participant.UserID)

	bytes, err = json.Marshal(LoB_temp)
	if err != nil {
		return false, errors.New("updateLoBUsers: Error marshalling new LoB info")
	}

	err = stub.PutState(LobName, bytes)
	if err != nil {
		return false, errors.New("updateLoBUsers: Error storing new LoB info")
	}
	return true, nil
}

//Helper: Update reading Holder - updates Index
func (rdg *SmartContract) updateReadingIDIndex(stub shim.ChaincodeStubInterface, participant Participant) (bool, error) {
	var participantIDs ReadingIDIndex
	bytes, err := stub.GetState("readingIDIndex")
	if err != nil {
		return false, errors.New("updateReadingIDIndex: Error getting readingIDIndex array Index from state")
	}
	logger.Info("Func------updateReadingIDIndex----Get readingIDIndex" + string(bytes))

	err = json.Unmarshal(bytes, &participantIDs)
	if err != nil {
		return false, errors.New("updateReadingIDIndex: Error unmarshalling readingIDIndex array JSON")
	}
	//Add participant.UserID
	logger.Info("Func------updateReadingIDIndex----Get participantID" + participant.UserID)
	participantIDs.UserIDs = append(participantIDs.UserIDs, participant.UserID)
	bytes, err = json.Marshal(participantIDs)
	logger.Info("Func------updateReadingIDIndex----json.Marshal(participantIDs)" + string(bytes))
	if err != nil {
		return false, errors.New("updateReadingIDIndex: Error marshalling new participant ID")
	}

	err = stub.PutState("readingIDIndex", bytes)
	if err != nil {
		return false, errors.New("updateReadingIDIndex: Error storing new participant ID in readingIDIndex (Index)")
	}
	return true, nil
}

//Query Route: readReading
func (rdg *SmartContract) readParticipant(stub shim.ChaincodeStubInterface, participantID string) peer.Response {
	participantAsByteArray, err := rdg.retrieveParticipant(stub, participantID)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(participantAsByteArray)
}

//Helper: Retrieve purchaser
//retrieve Participant
func (rdg *SmartContract) retrieveParticipant(stub shim.ChaincodeStubInterface, participantID string) ([]byte, error) {
	var participant Participant
	var participantAsByteArray []byte
	bytes, err := stub.GetState(participantID)

	if err != nil {
		return participantAsByteArray, errors.New("retrieveParticipant: Error retrieving participant with ID: " + participantID)
	}
	err = json.Unmarshal(bytes, &participant)
	if err != nil {
		return participantAsByteArray, errors.New("retrieveParticipant: Corrupt reading record " + string(bytes))
	}
	participantAsByteArray, err = json.Marshal(participant)
	if err != nil {
		return participantAsByteArray, errors.New("readParticipant: Invalid participant Object - Not a valid JSON")
	}
	return participantAsByteArray, nil
}

//Query Route: readAllReadings
func (rdg *SmartContract) readAllParticipant(stub shim.ChaincodeStubInterface) peer.Response {
	var readingIDs ReadingIDIndex
	bytes, err := stub.GetState("readingIDIndex")
	if err != nil {
		return shim.Error("readAllReadings: Error getting readingIDIndex array")
	}
	logger.Info("Func------readAllParticipant----Get readingIDIndex" + string(bytes))
	err = json.Unmarshal(bytes, &readingIDs)
	if err != nil {
		return shim.Error("readAllReadings: Error unmarshalling readingIDIndex array JSON")
	}
	result := "["

	var readingAsByteArray []byte

	for _, participantID := range readingIDs.UserIDs {
		readingAsByteArray, err = rdg.retrieveParticipant(stub, participantID)
		if err != nil {
			return shim.Error("Failed to retrieve participant with ID: " + participantID)
		}
		result += string(readingAsByteArray) + ","
	}
	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}
	return shim.Success([]byte(result))
}

//Helper: Reading readingStruct //change template
func (rdg *SmartContract) deleteParticipant(stub shim.ChaincodeStubInterface, participantID string) peer.Response {
	_, err := rdg.retrieveParticipant(stub, participantID)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.DelState(participantID)
	if err != nil {
		return shim.Error(err.Error())
	}
	_, err = rdg.deleteReadingIDIndex(stub, participantID)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//Helper: delete ID from readingStruct Holder
func (rdg *SmartContract) deleteReadingIDIndex(stub shim.ChaincodeStubInterface, participantID string) (bool, error) {
	var participantIDs ReadingIDIndex
	bytes, err := stub.GetState("readingIDIndex")
	if err != nil {
		return false, errors.New("deleteReadingIDIndex: Error getting readingIDIndex array Index from state")
	}
	err = json.Unmarshal(bytes, &participantIDs)
	if err != nil {
		return false, errors.New("deleteReadingIDIndex: Error unmarshalling readingIDIndex array JSON")
	}
	participantIDs.UserIDs, err = deleteKeyFromStringArray(participantIDs.UserIDs, participantID)
	if err != nil {
		return false, errors.New(err.Error())
	}
	bytes, err = json.Marshal(participantIDs)
	if err != nil {
		return false, errors.New("deleteReadingIDIndex: Error marshalling new readingStruct ID")
	}
	err = stub.PutState("readingIDIndex", bytes)
	if err != nil {
		return false, errors.New("deleteReadingIDIndex: Error storing new readingStruct ID in readingIDIndex (Index)")
	}
	return true, nil
}

//deleteKeyFromArray
func deleteKeyFromStringArray(array []string, key string) (newArray []string, err error) {
	for _, entry := range array {
		if entry != key {
			newArray = append(newArray, entry)
		}
	}
	if len(newArray) == len(array) {
		return newArray, errors.New("Specified Key: " + key + " not found in Array")
	}
	return newArray, nil
}

//Invoke Route: updateParticipant
func (rdg *SmartContract) updateParticipant(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var currParticipant Participant
	newParticipant, err := getParticipantFromArgs(args)
	participantAsByteArray, err := rdg.retrieveParticipant(stub, newParticipant.UserID)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = json.Unmarshal(participantAsByteArray, &currParticipant)
	if err != nil {
		return shim.Error("updateReading: Error unmarshalling readingStruct array JSON")
	}

	_, err = rdg.saveParticipant(stub, newParticipant)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (rdg *SmartContract) CreditCreate(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// ==== Check whether the participant already exsites. ====
	// todo
	// checke wether credit already exsites.
	record, err := stub.GetState("Credit_UerID_" + args[0])

	if record != nil {
		return shim.Error("This Credit" + args[0] + " credit has already existed.")
	}

	userID := args[0]
	value, err := strconv.Atoi(args[1])

	err = CreditInit(stub, userID, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func CreditInit(stub shim.ChaincodeStubInterface, userID string, value int) error {
	var credit Credit

	// ==== Create Credit object and Credit to JSON ====
	credit = Credit{UserID: userID, Value: value}

	creditAsByteArray, err := json.Marshal(credit)
	if err != nil {
		return errors.New(err.Error())
	}

	// ==== Save Credit to state ====
	err = stub.PutState("Credit_UerID_"+userID, creditAsByteArray)
	if err != nil {
		return errors.New(err.Error())
	}

	return nil
}

func (rdg *SmartContract) CreditRead(stub shim.ChaincodeStubInterface, UserID string) peer.Response {
	//to do
	creditAsByteArray, err := retrieveSingleCreditAsByteArray(stub, "Credit_UerID_"+UserID)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(creditAsByteArray)
}

func retrieveSingleCredit(stub shim.ChaincodeStubInterface, creditID string) (Credit, error) {
	var credit Credit
	var creditAsByteArray []byte
	var err error

	creditAsByteArray, err = stub.GetState(creditID)

	if err != nil {
		return credit, errors.New("CreditRead: Error credit read participant with ID: " + creditID)
	}
	// else if creditAsBytes == nil {
	//  logger.Error("CreditRead:  Corrupt reading record ", err.Error())
	//  return nil, errors.New("CreditRead: Credit does not exist " + creditID)
	// }

	// For log printing credit Information & check whether the credit does exist
	err = json.Unmarshal(creditAsByteArray, &credit)
	if err != nil {
		return credit, errors.New("CreditRead: Credit does not exist " + string(creditAsByteArray))
	}
	// For log printing credit Information

	return credit, nil
}

func retrieveSingleCreditAsByteArray(stub shim.ChaincodeStubInterface, creditID string) ([]byte, error) {
	var credit Credit
	var creditAsByteArray []byte
	var err error

	logger.Info("-----retrieveSingleCreditAsByteArray :creditID---------", creditID)
	creditAsByteArray, err = stub.GetState(creditID)

	if err != nil {
		return nil, errors.New("CreditRead: Error credit read participant with ID: " + creditID)
	}
	// else if creditAsBytes == nil {
	//  logger.Error("CreditRead:  Corrupt reading record ", err.Error())
	//  return nil, errors.New("CreditRead: Credit does not exist " + creditID)
	// }

	// For log printing credit Information & check whether the credit does exist
	err = json.Unmarshal(creditAsByteArray, &credit)
	if err != nil {
		return nil, errors.New("CreditRead: Credit does not exist " + string(creditAsByteArray))
	}
	// For log printing credit Information

	logger.Info("-----retrieveSingleCreditAsByteArray---------", credit)

	return creditAsByteArray, nil
}

func (rdg *SmartContract) CreditAdd(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var credit Credit
	var raw map[string]interface{}

	err := json.Unmarshal([]byte(args[0]), &raw)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Assign value to variable ====
	userID := raw["userID"].(string)
	logger.Info("*****CreditUpdate*******", userID)

	value := int(raw["value"].(float64))
	logger.Info("*****CreditUpdate*******", value)

	ticketID := raw["ticketID"].(string)
	logger.Info("*****CreditUpdate*******", ticketID)

	// === Check whether the credit already exist. ====
	creditAsByteArray, err := stub.GetState("Credit_UerID_" + userID)
	if err != nil {
		return shim.Error("CreditUpdate: Failed to get credit :" + err.Error())
	} else if creditAsByteArray == nil {
		errs := fmt.Sprintf("CreditUpdate: Credit_UerID_%s does not exist.", userID)
		logger.Info(" ****** " + errs)
		return shim.Error(errs)
	}

	err = json.Unmarshal(creditAsByteArray, &credit)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === if ticket is a constan string which only represent add constant credit ===
	if ticketID != "creditADD" {
		// === check whether the ticket has been add ===
		if ok := Is_Inarray(credit.TicketIDs, ticketID); ok {
			return shim.Error("CreditUpdate: This ticket has been existed.")
		}
	}

	credit.Value += value
	credit.TicketIDs = append(credit.TicketIDs, ticketID)

	creditAsByteArray, err = json.Marshal(credit)
	if err != nil {
		return shim.Error("CreditUpdate: " + err.Error())
	}

	err = stub.PutState("Credit_UerID_"+credit.UserID, creditAsByteArray)
	return shim.Success(creditAsByteArray)
}

func Is_Inarray(target []string, now string) bool {
	for _, entry := range target {
		if entry == now {
			return true
		}
	}
	return false
}

func (rdg *SmartContract) CreditDelete(stub shim.ChaincodeStubInterface, userID string) peer.Response {
	logger.Info(" ****** CreditDelete start ****** userID:" + userID)
	err := stub.DelState("Credit_UerID_" + userID)
	if err != nil {
		return shim.Error("CreditDelete: Failed to delete Credit state: " + err.Error())
	}

	//Log process for debug
	credit, err := stub.GetState("Credit_UerID_" + userID)
	logger.Info(" ****** CreditDelete ****** " + string(credit))

	return shim.Success(nil)
}

func (rdg *SmartContract) LoBReadAll(stub shim.ChaincodeStubInterface) peer.Response {
	var LoB_temp LoB
	var result string

	result += "["

	iter := 0
	for iter < NumberOfLoBs {
		bytes, err := stub.GetState(Lob_Name[iter])
		if err != nil {
			return shim.Error("LoBReadAll: Error getting LoB info from state")
		}

		err = json.Unmarshal(bytes, &LoB_temp)
		if err != nil {
			return shim.Error("LoBReadAll: Error unmarshalling LoB JSON")
		}

		LobAssest := LoB{LoBID: LoB_temp.LoBID, TotalCredit: LoB_temp.TotalCredit}
		LobAssestAsByteArray, err := json.Marshal(LobAssest)
		if err != nil {
			return shim.Error("LoBReadAll: Fail to Marshall LobAssest")
		}
		result += string(LobAssestAsByteArray) + ","
		iter = iter + 1
	}

	result = result[:len(result)-1] + "]"

	return shim.Success([]byte(result))
}

func (rdg *SmartContract) LoBRead(stub shim.ChaincodeStubInterface, LoBid string) peer.Response {
	var LoB_temp LoB
	var participant_temp Participant
	var participantAsByteArray []byte
	var credit_temp Credit
	var result string

	LoBID, _ := strconv.Atoi(LoBid)
	if LoBID < 0 || LoBID >= NumberOfLoBs {
		return shim.Error("Input LoBID is invalid, LoBID")
	}

	LobName := Lob_Name[LoBID]
	bytes, err := stub.GetState(LobName)
	if err != nil {
		return shim.Error("LoBRead: Error getting LoB info from state")
	}

	err = json.Unmarshal(bytes, &LoB_temp)
	if err != nil {
		return shim.Error("LoBRead: Error unmarshalling LoB JSON")
	}

	result += "["
	credit := strconv.Itoa(LoB_temp.TotalCredit)
	result += "TotalCredit: " + credit + ","

	for _, participantID := range LoB_temp.UserIDs {
		participantAsByteArray, err = rdg.retrieveParticipant(stub, participantID)
		if err != nil {
			return shim.Error("Failed to retrieve participant with ID: " + participantID)
		}

		err = json.Unmarshal(participantAsByteArray, &participant_temp)
		if err != nil {
			return shim.Error("LoBRead: Error unmarshalling Participant JSON")
		}

		credit_temp, _ = retrieveSingleCredit(stub, "Credit_UerID_"+participant_temp.UserID)

		Participant_UserID := "{\"participant_UserID\": " + participant_temp.UserID + ","
		Participant_UserName := "\"participant_UserName\": " + participant_temp.UserName + ","
		Participant_credit := "\"participant_credit\": " + strconv.Itoa(credit_temp.Value) + ","
		Participant_LoB := "\"participant_LoB\": " + strconv.Itoa(participant_temp.LoBID) + "},"
		result += Participant_UserID + Participant_UserName + Participant_credit + Participant_LoB
	}
	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "\n]"
	}
	return shim.Success([]byte(result))
}

func (rdg *SmartContract) TopTenCredit(stub shim.ChaincodeStubInterface) peer.Response {
	var readingIDs ReadingIDIndex
	var participant_temp Participant
	var credits Credits
	var credit_temp Credit
	var participantAsBytes []byte
	var err error
	var result string

	result += "["

	bytes, _ := stub.GetState("readingIDIndex")
	err = json.Unmarshal(bytes, &readingIDs)
	if err != nil {
		return shim.Error("TopTenCredit: Error unmarshalling readingIDIndex array JSON")
	}

	for _, participantID := range readingIDs.UserIDs {
		credit_temp, _ = retrieveSingleCredit(stub, "Credit_UerID_"+participantID)
		credits = append(credits, credit_temp)
	}

	sort.Sort(credits)

	for i := 0; i < Min(10, len(credits)); i++ {
		participantAsBytes, _ = rdg.retrieveParticipant(stub, credits[i].UserID)
		err = json.Unmarshal(participantAsBytes, &participant_temp)
		if err != nil {
			return shim.Error("TopTenParticipant: Error unmarshalling Participant JSON")
		}

		Participant_UserID := "{\"participant_UserID\": \"" + participant_temp.UserID + "\","
		Participant_UserName := "\"participant_UserName\": \"" + participant_temp.UserName + "\","
		Participant_credit := "\"participant_credit\": " + strconv.Itoa(credits[i].Value) + ","
		Participant_LoB := "\"participant_LoB\": " + strconv.Itoa(participant_temp.LoBID) + "},"
		result += Participant_UserID + Participant_UserName + Participant_credit + Participant_LoB
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "\n]"
	}
	return shim.Success([]byte(result))
}

// go lib doesn't have Min/Max(int, int) funct
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func getTicketFromArgs(args string) (ticket Ticket, err error) {
	if strings.Contains(args, "\"Ticket_Title\"") == false ||
		strings.Contains(args, "\"Ticket_Value\"") == false ||
		strings.Contains(args, "\"Ticket_UserID\"") == false ||
		strings.Contains(args, "\"Ticket_Type\"") == false {
		return ticket, errors.New("Unknown field: Input JSON does not comly to schema")
	}

	err = json.Unmarshal([]byte(args), &ticket)
	if err != nil {
		return ticket, err
	}

	return ticket, nil
}

func saveTicket(stub shim.ChaincodeStubInterface, ticket Ticket) ([]byte, error) {
	var ticketAsBytes []byte
	ticketAsBytes, err := json.Marshal(ticket)
	if err != nil {
		return ticketAsBytes, errors.New("saveTicket: " + err.Error())
	}
	err = stub.PutState(ticket.TicketID, ticketAsBytes)
	if err != nil {
		return ticketAsBytes, err
	}
	return ticketAsBytes, nil
}

func (sc *SmartContract) TicketCreate(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// ==== Get ticket from args ====
	// todo
	// ticket := Ticket{
	// 	TicketID: "ticket_1",
	// 	Status: 0,
	// 	Title: "The first Ticket",
	// 	Value: 30,
	// 	UserID: "1",
	// 	DeadLine: time.Now()}

	ticket, err := getTicketFromArgs(args[0])

	TICKETIDAsBytes, _ := stub.GetState("TICKETID")
	TICKETID, _ = strconv.Atoi(string(TICKETIDAsBytes))
	TICKETID++
	ticket.TicketID = strconv.Itoa(TICKETID)
	ticket.Status = 1
	if err != nil {
		return shim.Error("TicketCreate: " + err.Error())
	}

	// ==== Judge if the ticket already exists ====
	// ticketAsBytes, err := stub.GetState(ticket.TicketID)
	// if ticketAsBytes != nil {
	// 	return shim.Error("TicketCreate: The ticket already exists." + string(ticketAsBytes))
	// }
	// todo
	// check if userid is valid

	// ==== Put the ticket into ledger ====
	ticketAsBytes, err := saveTicket(stub, ticket)
	if err != nil {
		return shim.Error(err.Error())
	}

	TICKETIDAsBytes, _ = json.Marshal(TICKETID)
	stub.PutState("TICKETID", TICKETIDAsBytes)
	return shim.Success(ticketAsBytes)
}

func (sc *SmartContract) TicketDelete(stub shim.ChaincodeStubInterface, ticketID string) peer.Response {
	// ==== Judge if the ticket already exists ====
	var ticket Ticket
	ticketAsBytes, err := stub.GetState(ticketID)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = json.Unmarshal(ticketAsBytes, &ticket)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Info(" ****** TicketDelete:", ticket)

	err = stub.DelState(ticketID)
	return shim.Success(nil)
}

func (sc *SmartContract) TicketUpdate(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// ==== Get ticket from args ====
	// todo
	// ticket := Ticket{
	// 	TicketID: "ticket_1",
	// 	Status: 1,
	// 	Title: "The first Ticket 1111",
	// 	Value: 20,
	// 	UserID: "111",
	// 	DeadLine: time.Now()}

	// participantID := args[0]
	ticket, err := getTicketFromArgs(args[0])
	if err != nil {
		return shim.Error("TicketUpdate: " + err.Error())
	}
	// ==== Judge if the ticket already exists ====
	ticketAsBytes, err := stub.GetState(ticket.TicketID)
	if ticketAsBytes == nil {
		return shim.Error("TicketCreate: The ticket does not exist.")
	}

	// if participantID != ticket.UserID {
	// 	return shim.Error("TicketUpdate: You have no rights to update the ticket")
	// }

	// ==== Update the ledger ====
	ticketAsBytes, err = saveTicket(stub, ticket)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(ticketAsBytes)
}

func (sc *SmartContract) TicketRead(stub shim.ChaincodeStubInterface, args string) peer.Response {
	// ==== Read ticket from ledger ====
	var ticket Ticket
	ticketAsBytes, err := stub.GetState(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = json.Unmarshal(ticketAsBytes, &ticket)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Info(" ****** TicketRead:", ticket)
	return shim.Success(ticketAsBytes)
}

func (sc *SmartContract) TicketRead2(stub shim.ChaincodeStubInterface) peer.Response {
	TICKETIDAsBytes, _ := stub.GetState("TICKETID")
	TICKETID, _ = strconv.Atoi(string(TICKETIDAsBytes))

	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	i := 1
	//var ticket Ticket

	for i <= TICKETID {
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		//TicketArray = append(TicketArray, strconv.Itoa(i))
		//TicketIDs = strconv.Itoa(i)
		ticketAsBytes, _ := stub.GetState(strconv.Itoa(i))
		//item, _ := json.Marshal(sc.TicketRead(stub, strconv.Itoa(i)))
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(ticketAsBytes))
		bArrayMemberAlreadyWritten = true
		i = i + 1
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}

func (sc *SmartContract) TestGetHistoryTicket(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	ticketInterator, err := stub.GetHistoryForKey("xxx1")
	if err != nil {
		return shim.Error(err.Error())
	}

	// defer ticketInterator.Close()

	for ticketInterator.HasNext() {
		queryResponse, err := ticketInterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		logger.Info("------Test----" + queryResponse.String())
		item, _ := json.Marshal(queryResponse)
		logger.Info("------Test 1----" + string(item))

	}

	return shim.Success(nil)
}

func (sc *SmartContract) OrderCreate(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// json ticketID & userID
	//

	var order Order
	if strings.Contains(args[0], "\"TicketID\"") == false ||
		strings.Contains(args[0], "\"UserID\"") == false {
		return shim.Error("OrderCreate:Unknown field: Input JSON does not comly to schema")
	}

	err := json.Unmarshal([]byte(args[0]), &order)
	if err != nil {
		return shim.Error("OrderCreate:")
	}
	ticketID := order.TicketID
	userID := order.UserID

	key, _ := stub.CreateCompositeKey("Order", []string{ticketID, userID})
	logger.Info("------OrderCreate:" + key)

	// ==== check whether the order already exsit ====
	orderAsByte, _ := stub.GetState(key)
	if orderAsByte != nil {
		return shim.Error("OrderCreate: You have applied this Ticket")
	}

	order.Status = 1
	orderAsByte, _ = json.Marshal(order)
	stub.PutState(key, orderAsByte)

	return shim.Success(orderAsByte)
}

func (sc *SmartContract) OrderRead(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	ticketID := args[0]
	userID := args[1]

	logger.Info("OrderRead :", ticketID, userID)
	key, _ := stub.CreateCompositeKey("Order", []string{ticketID, userID})
	orderAsByte, _ := stub.GetState(key)

	logger.Info("OrderRead orderAsByte:", orderAsByte)
	var order Order
	_ = json.Unmarshal(orderAsByte, &order)

	logger.Info("OrderRead order:", order)

	return shim.Success(orderAsByte)
}

func (sc *SmartContract) OrderRead2(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	ticketID := args[0]

	orderInterator, _ :=
		stub.GetStateByPartialCompositeKey("Order", []string{ticketID})

	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	logger.Info("OrderRead2 Interator start:")
	for orderInterator.HasNext() {
		queryResponse, err := orderInterator.Next()
		logger.Info("OrderRead2 Interator :", queryResponse)
		if err != nil {
			return shim.Error(err.Error())
		}

		logger.Info("------Test----" + queryResponse.String())
		item, _ := json.Marshal(queryResponse)
		logger.Info("------Test x----" + string(item))

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true

	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

func OrderSaving(stub shim.ChaincodeStubInterface, order Order) ([]byte, error) {
	bytes, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	key, _ := stub.CreateCompositeKey(
		"Order",
		[]string{order.TicketID, order.UserID})

	err = stub.PutState(key, bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func OrderBlukUpdate(stub shim.ChaincodeStubInterface, ticketID interface{}, userID_array []interface{}, status int) (bool, error) {
	for _, userID := range userID_array {
		order := Order{
			TicketID: ticketID.(string),
			UserID:   userID.(string),
			Status:   status}
		if status != 0 {
			key, _ := stub.CreateCompositeKey("Order", []string{ticketID.(string), userID.(string)})
			orderAsByte, _ := stub.GetState(key)

			var order Order
			_ = json.Unmarshal(orderAsByte, &order)

			if order.Status == status-1 {
				order.Status = status
				OrderSaving(stub, order)
			}
		} else {
			order.Status = status
			OrderSaving(stub, order)
		}
	}
	return true, nil
}

func orderStatueEqual(stub shim.ChaincodeStubInterface, ticketID string, userID string, status int) bool {
	key, _ := stub.CreateCompositeKey("Order", []string{ticketID, userID})
	orderAsByte, _ := stub.GetState(key)

	var order Order
	_ = json.Unmarshal(orderAsByte, &order)

	if order.Status == status {
		return true
	} else {
		return false
	}
}

func award(stub shim.ChaincodeStubInterface, ticketID string, userID_array []interface{}, value int) (bool, error) {
	for _, userID := range userID_array {
		logger.Info("-----xxx---------", "Credit_UerID_"+userID.(string))
		credit, _ := retrieveSingleCredit(stub, "Credit_UerID_"+userID.(string))
		// if order is done and ticketID not in credit.TicketIDs
		if orderStatueEqual(stub, ticketID, userID.(string), 4) &&
			!Is_Inarray(credit.TicketIDs, ticketID) {
			credit.Value += value
			credit.TicketIDs = append(credit.TicketIDs, ticketID)
			logger.Info("-----xxx---------", credit)
			creditAsByteArray, _ := json.Marshal(credit)
			stub.PutState("Credit_UerID_"+credit.UserID, creditAsByteArray)

			// update user's LoB total credit
			_, err := updateLoBCredit(stub, userID.(string), value)
			if err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func updateLoBCredit(stub shim.ChaincodeStubInterface, userID string, value int) (bool, error) {
	var participant Participant
	var LoB_temp LoB

	bytes, err := stub.GetState(userID)
	if err != nil {
		return false, errors.New("updateLoBCredit: Error get participant with ID: " + userID)
	}
	err = json.Unmarshal(bytes, &participant)
	if err != nil {
		return false, errors.New("updateLoBCredit: Corrupt reading record " + string(bytes))
	}

	LobName := Lob_Name[participant.LoBID]
	bytes, err = stub.GetState(LobName)
	if err != nil {
		return false, errors.New("updateLoBCredit: Error getting LoB info from state")
	}

	err = json.Unmarshal(bytes, &LoB_temp)
	if err != nil {
		return false, errors.New("updateLoBCredit: Error unmarshalling LoB JSON")
	}
	LoB_temp.TotalCredit += value
	bytes, err = json.Marshal(LoB_temp)
	if err != nil {
		return false, errors.New("updateLoBCredit: Error marshalling new LoB info")
	}

	err = stub.PutState(LobName, bytes)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (sc *SmartContract) OrderUpdate(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var raw map[string]interface{}

	err := json.Unmarshal([]byte(args[0]), &raw)
	if err != nil {
		return shim.Error(err.Error())
	}

	var data []interface{}
	var status = 0
	ticketID := raw["TicketID"]
	confirm := raw["Confirm"]
	close := raw["Close"]
	done := raw["Done"]
	awarded := raw["Award"]

	if ticketID == nil {
		return shim.Error("OrderUpdate: TicketID is needed")
	}
	if close != nil {
		_, err = OrderBlukUpdate(stub, ticketID, close.([]interface{}), status)
		if err != nil {
			return shim.Error("OrderUpdate:" + err.Error())
		}
	}

	logger.Info("[OrderUpdate]--------------", raw)
	logger.Info("[OrderUpdate]-----if start---------")
	if confirm != nil {
		data = confirm.([]interface{})
		status = 2
		_, err = OrderBlukUpdate(stub, ticketID, data, status)
	} else if done != nil {
		data = done.([]interface{})
		status = 3
		_, err = OrderBlukUpdate(stub, ticketID, data, status)
	} else if awarded != nil {
		data = awarded.([]interface{})
		status = 4
		_, err = OrderBlukUpdate(stub, ticketID, data, status)

		logger.Info("-----1---------")
		// Award user
		var ticket Ticket
		ticketAsBytes, err := stub.GetState(ticketID.(string))
		if err != nil {
			return shim.Error(err.Error())
		}

		err = json.Unmarshal(ticketAsBytes, &ticket)
		if err != nil {
			return shim.Error(err.Error())
		}
		award(stub, ticket.TicketID, data, ticket.Value)
	}
	logger.Info("[OrderUpdate]-----if end---------")

	if err != nil {
		return shim.Error("OrderUpdate:" + err.Error())
	}

	logger.Info("[OrderUpdate]-----order read2 start---------")
	sc.OrderRead2(stub, []string{ticketID.(string)})
	logger.Info("[OrderUpdate]-----order read2 start---------")

	logger.Info("[OrderUpdate]-----time sleep start---------")
	//time.Sleep(time.Second*30)
	logger.Info("[OrderUpdate]-----time sleep end---------")
	// update ticket status

	logger.Info("[OrderUpdate]-----AutoUpdateTicketStatus start---------")
	sc.AutoUpdateTicketStatus(stub, ticketID.(string))
	logger.Info("[OrderUpdate]-----AutoUpdateTicketStatus end---------")
	// ticket id & ticket object list

	logger.Info("[OrderUpdate]-----order read2 2 start---------")
	sc.OrderRead2(stub, []string{ticketID.(string)})
	logger.Info("[OrderUpdate]-----order read2 2 start---------")

	// to do
	return shim.Success(nil)
}

func (sc *SmartContract) AutoUpdateTicketStatus(stub shim.ChaincodeStubInterface, args string) peer.Response {
	var maxStatus = 0
	var ticketID = args
	var order Order
	var ticket Ticket
	logger.Info("AutoUpdateTicketStatus ticketID:", ticketID)
	// Get all order to get max status
	orderInterator, err :=
		stub.GetStateByPartialCompositeKey("Order", []string{ticketID})

	if err != nil {
		logger.Info("AutoUpdateTicketStatus Error:", err)
	}

	var lognum = 1
	logger.Info("AutoUpdateTicketStatus orderInterator 1 :", orderInterator)

	queryResponse, _ := orderInterator.Next()
	json.Unmarshal(queryResponse.Value, &order)

	if maxStatus < order.Status {
		maxStatus = order.Status
	}
	logger.Info("AutoUpdateTicketStatus order  : ", lognum, order)

	logger.Info("AutoUpdateTicketStatus Interator for start...")
	for orderInterator.HasNext() {
		queryResponse, _ := orderInterator.Next()

		json.Unmarshal(queryResponse.Value, &order)

		lognum = lognum + 1
		logger.Info("AutoUpdateTicketStatus order :", lognum, order)
		if maxStatus < order.Status {
			maxStatus = order.Status
		}
	}

	ticketAsBytes, _ := stub.GetState(ticketID)
	json.Unmarshal(ticketAsBytes, &ticket)

	ticket.Status = maxStatus
	logger.Info("AutoUpdateTicketStatus:", maxStatus)

	ticketAsBytes, _ = json.Marshal(ticket)
	stub.PutState(ticket.TicketID, ticketAsBytes)
	return shim.Success(ticketAsBytes)
}

func string2time(st string) (theTime time.Time, err error) {
	timeFormated := "2018-11-26 18:05:00"
	loc, _ := time.LoadLocation("Local")
	theTime, err = time.ParseInLocation(st, timeFormated, loc)
	if err != nil {
		return theTime, err
	}
	return theTime, nil
}
