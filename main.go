package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	mspprotos "github.com/hyperledger/fabric/protos/msp"
	sc "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("example_cc0")

type SmartContract struct {
}

type workspace struct {
	Wtype            string           `json:"WorksapceType"`
	PhysicalLocation physicalLocation `json:"physicalLocation"`
	Furnitures       []furniture      `json:"furniture"`
	EAsset           electricalAsset  `json:electricalAsset`
	NetAsset         networkAsset     `json:"networkAsset"`
	Requests         []WSpaceSchedule `json:"requests"`
	Schedule         []WSpaceSchedule `json:"schedule"`
}
type WSpaceSchedule struct {
	ScheduleId string `json:"schedleId"`
	UserId     string `json:"UserID"`
}
type physicalLocation struct {
	Country       string `json:"country"`
	City          string `json:"city"`
	BuildingName  string `json:buildingName`
	Floor         string `json:"floor"`
	Wing          string `json:"wing"`
	WorkSpaceName string `json:"wspaceName"`
}

type furniture struct {
	Fname    string `json:"fname"`
	Quantity string `json:"quantity"`
}

type electricalAsset struct {
	Switches []switchs
}

type switchs struct {
	SwitchID         string `json:"switchID"`
	SwitchAssignedTo string `json:"switchAssignedTo"`
	Status           string `json:"Status"`
}

type networkAsset struct {
	IpPorts   []ipPortsConfig `json:"ipports"`
	Telephone telePortsConfig `json:"telePorts"`
}

type ipPortsConfig struct {
	PortNo          string `json:"ipportNo"`
	ConfigurationID string `json:"configID"`
}
type telePortsConfig struct {
	Extension_Number string `json:"telePhoneNo"`
	Organization     string `json:"org"`
}

type config struct {
	Org           string   `json:"org"`
	RestrictedIP  []string `json:"restrictedIP"`
	Gateway       string   `json:"gateway"`
	Netmask       string   `json:"netMask"`
	WhiteList     []string `json:"whiteList"`
	BlackList     []string `json:"blackList"`
	DNS_Primary   string   `json:"DNS_Primary"`
	DNS_Secondary string   `json:"DNS_Secondary"`
}

type user struct {
	FirstName string         `json:"fname"`
	LastName  string         `json:"lastName"`
	Age       string         `json:"age"`
	EmailID   string         `json:"emailID"`
	Org       string         `json:"org"`
	Calendar  []UserSchedule `json:"calander"`
	Policy    []policy       `json:"policy"`
}

type UserSchedule struct {
	ScheduleId  string `json:"scheduleID"`
	WorkSpaceId string `json:"wrkSpaceID"`
}

type schedule struct {
	StartTime     int    `json:"startTime"`
	EndTime       int    `json:"endTime"`
	BookingTime   int    `json:"bookingTime"`
	BookingStatus string `json:"bookStatus"`
	OccupiedTxID  string `json:"occupiedTxID"`
	BookingTxID   string `json:BookingTxID`
}

type policy struct {
	WhiteList   []string `json:"userwhitelist"`
	BlackList   []string `json:"blackList"`
	PermittedIP []string `json:"permittedIP"`
}

type IDs struct {
	CubicleID  string `json:"cubID"`
	RoomID     string `json:"roomID"`
	ConfRoomID string `json:"confRoomID"`
	BookingId  string `json:"BookingId"`
}

type Response struct {
	RestrictedSites  []string `json:"restrictedSites"`
	RestrictedIP     []string `json:"restrictedIP"`
	Extension_Number string   `json:"extensionNumber"`
	SwitchID         string   `json:"switchID"`
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	logger.Info("-------------------Invoke----------")
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	logger.Info("fucntion")
	logger.Info(function)
	logger.Info("args")
	logger.Info(args)
	if function == "newuser" {
		return s.KYCRegistration(APIstub, args)
	} else if function == "newWorkSpace" {
		return s.CreateWorkspace(APIstub)
	} else if function == "queryWorkspace" {
		return s.QueryWorkspace(APIstub, args)
	} else if function == "createConfig" {
		return s.CreateConfig(APIstub)
	} else if function == "bookWorkSpace" {
		return s.BookWorkspace(APIstub)
	} else if function == "ApproveOrDeny" {
		return s.ApproveOrDeny(APIstub, args)
	} else if function == "occupyWorkSpace" {
		return s.OccupyWorkSpace(APIstub, args)
	} else if function == "query" {
		return s.QueryAllschedules(APIstub, args)
	} else if function == "switching" {
		return s.Switching(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("-----------Instantiating---------------------------")
	id := []IDs{
		IDs{CubicleID: "Cubicle101", RoomID: "Room101", ConfRoomID: "ConfRoom101", BookingId: "1001"},
	}

	idsAsBytes, _ := json.Marshal(id[0])
	APIstub.PutState("ids", idsAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) KYCRegistration(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	logger.Info("########### KYCRegistration ###########")

	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}

	id := &mspprotos.SerializedIdentity{}
	err = proto.Unmarshal(creator, id)
	block, _ := pem.Decode(id.GetIdBytes())
	cert, err := x509.ParseCertificate(block.Bytes)
	enrollID := cert.Subject.CommonName

	userAsbytes, _ := stub.GetState(enrollID)

	if userAsbytes != nil {
		return shim.Error("User already exists")
	}
	fmt.Printf("enrollID: %s", enrollID)
	mspID := id.GetMspid()

	newuser := user{FirstName: args[0], LastName: args[1], Age: args[2], EmailID: args[3]}
	newuser.Org = mspID
	newuserAsBytes, _ := json.Marshal(newuser)
	stub.PutState(enrollID, newuserAsBytes)
	// fmt.Println("enrollID", enrollID)
	fmt.Println("user object ", newuser)
	return shim.Success(nil)
}

func (s *SmartContract) CreateWorkspace(stub shim.ChaincodeStubInterface) sc.Response {

	logger.Info("##################CreateWorkspace##################")

	atrrValue, _, _ := cid.GetAttributeValue(stub, "Role")
	logger.Info("Role ", atrrValue)

	if strings.Compare(strings.ToLower(string(atrrValue)), "admin") != 0 {
		//		logger.Info("acess denied")
		return shim.Error("Access denied")
	}
	args := stub.GetArgs()
	newWrkSpace := workspace{}
	idsAsBytes, _ := stub.GetState("ids")
	id := IDs{}
	json.Unmarshal(idsAsBytes, &id)
	key := ""

	//------------------------------get IDS--------------------------------
	if strings.Compare(strings.ToLower(string(args[1])), "cubicle") == 0 {

		key = id.CubicleID
		j := string([]rune(id.CubicleID)[7])
		cubId_no, _ := strconv.Atoi(j)
		cubId_no = cubId_no + 1
		cubId := "Cubicle" + strconv.Itoa(cubId_no)
		id.CubicleID = cubId
		idsAsBytes, _ = json.Marshal(id)
		stub.PutState("ids", idsAsBytes)

	} else if strings.Compare(strings.ToLower(string(args[1])), "room") == 0 {
		key = id.RoomID
		j := string([]rune(id.RoomID)[4])
		roomId_no, _ := strconv.Atoi(j)
		roomId_no = roomId_no + 1
		roomId := "Room" + strconv.Itoa(roomId_no)
		id.RoomID = roomId
		idsAsBytes, _ = json.Marshal(id)
		stub.PutState("ids", idsAsBytes)
	} else if strings.Compare(strings.ToLower(string(args[1])), "confroom") == 0 {
		key = id.ConfRoomID
		j := string([]rune(id.ConfRoomID)[8])
		confID_no, _ := strconv.Atoi(j)
		confID_no = confID_no + 1
		confID := "ConfRoom" + strconv.Itoa(confID_no)
		id.ConfRoomID = confID
		idsAsBytes, _ = json.Marshal(id)
		stub.PutState("ids", idsAsBytes)
	} else {
		shim.Error("Invlaid workspace type")
	}
	// ------------------------Wtype---------------
	newWrkSpace.Wtype = string(args[1])
	fmt.Println(string(args[1]))

	// ---------------------physical location-----------------
	physlocation := physicalLocation{}
	pyslocAsbytes := args[2]

	if err := json.Unmarshal(pyslocAsbytes, &physlocation); err != nil {
		log.Fatal(err)
	}
	//	------------------------validate wether it already exists------------------
	indexName := "location~wspID"
	it, _ := stub.GetStateByPartialCompositeKey(indexName, []string{physlocation.Country, physlocation.City, physlocation.BuildingName, physlocation.Floor, physlocation.Wing, physlocation.WorkSpaceName})
	defer it.Close()

	_, err := it.Next()
	if err == nil {

		return shim.Error("Worksapce Already Exists")
	}
	newWrkSpace.PhysicalLocation = physlocation

	//	----------------------------Adding composite key--------------------
	location, _ := stub.CreateCompositeKey(indexName, []string{physlocation.Country, physlocation.City, physlocation.BuildingName, physlocation.Floor, physlocation.Wing, physlocation.WorkSpaceName, key})
	logger.Info("composite key", location)
	logger.Info("array :", physlocation.Country, physlocation.City, physlocation.BuildingName, physlocation.Floor, physlocation.Wing, physlocation.WorkSpaceName)
	value := []byte{0x00}
	stub.PutState(location, value)

	// --------------------------------furnitures------------------
	furn := []furniture{}
	furnAsbytes := args[3]
	if err := json.Unmarshal(furnAsbytes, &furn); err != nil {
		log.Fatal(err)
	}
	newWrkSpace.Furnitures = furn
	// -----------------------electricalAsset-------------------
	switchArrAsBytes := args[4]
	fmt.Println("electrical Asset ", string(args[4]))
	var switchArrData [][]string
	json.Unmarshal(switchArrAsBytes, &switchArrData)
	sw := []switchs{}
	for i := 0; i < len(switchArrData); i++ {

		s := switchs{}
		s.SwitchID = switchArrData[i][0]
		s.SwitchAssignedTo = switchArrData[i][1]
		s.Status = switchArrData[i][2]
		sw = append(sw, s)
	}

	newWrkSpace.EAsset.Switches = sw
	// --------------------------NetworkAsset-------------------
	fmt.Println("Nasset ", string(args[5]))
	portsAsBytes := args[5]
	nA := networkAsset{}
	json.Unmarshal(portsAsBytes, &nA)
	newWrkSpace.NetAsset = nA
	fmt.Println("Workspace; ", newWrkSpace, "key: ", key)
	workspaceAsbytes, _ := json.Marshal(newWrkSpace)
	stub.PutState(key, workspaceAsbytes)

	return shim.Success(nil)

}

func (s *SmartContract) QueryWorkspace(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}
	logger.Info("strings", args)
	indexName := "location~wspID"
	it, _ := APIstub.GetStateByPartialCompositeKey(indexName, args)
	defer it.Close()
	logger.Info("it :", it)

	locationRange, err := it.Next()
	if err != nil {

		return shim.Error("Worksapce not found")
	}

	_, compositeKeyParts, err := APIstub.SplitCompositeKey(locationRange.Key)
	if err != nil {
		return shim.Error(err.Error())
	}

	keyAsString := compositeKeyParts[6]
	logger.Info("key: ", keyAsString)
	wrkSpaceAsBytes, _ := APIstub.GetState("Cubicle101")
	wrkspace := workspace{}
	json.Unmarshal(wrkSpaceAsBytes, &wrkspace)
	logger.Info("workspace ;", wrkspace)
	return shim.Success(wrkSpaceAsBytes)

}

func (s *SmartContract) CreateConfig(stub shim.ChaincodeStubInterface) sc.Response {
	logger.Info("-------------------create config----------")
	org, _ := cid.GetMSPID(stub)
	orgAsbytes, _ := stub.GetState(org)

	if orgAsbytes != nil {
		return shim.Error("Configuration already exists")
	}
	atrrValue, _, _ := cid.GetAttributeValue(stub, "Role")
	logger.Info("er ", atrrValue)

	if strings.Compare(strings.ToLower(string(atrrValue)), "network admin") != 0 {
		logger.Info("attr value: ", strings.Compare(strings.ToLower(string(atrrValue)), "network admin"))
		return shim.Error("Access denied")
	}

	Config := config{}
	Config.Org = org
	args := stub.GetArgs()
	configAsBytes := args[1]
	json.Unmarshal(configAsBytes, &Config)
	fmt.Println("config")
	fmt.Println(Config)
	configAsBytes, _ = json.Marshal(Config)
	logger.Info("config ", Config)
	stub.PutState(org, configAsBytes)
	return shim.Success(nil)
}
func (s *SmartContract) BookWorkspace(stub shim.ChaincodeStubInterface) sc.Response {

	logger.Info("=============BookworkSpace---------------")
	args := stub.GetArgs()
	indexName := "location~wspID"
	str := []string{}
	json.Unmarshal(args[1], &str)
	fmt.Println("string array of cubicle details:  ", str)
	it, _ := stub.GetStateByPartialCompositeKey(indexName, str)
	defer it.Close()
	logger.Info("it :", it)
	// ------------------------------------------
	logger.Info("arg2", args[2])
	// -----------------------------------------------
	locationRange, err := it.Next()
	if err != nil {

		return shim.Error("Worksapce not found")
	}

	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}

	id := &mspprotos.SerializedIdentity{}
	err = proto.Unmarshal(creator, id)
	block, _ := pem.Decode(id.GetIdBytes())
	cert, err := x509.ParseCertificate(block.Bytes)
	enrollID := cert.Subject.CommonName
	//---------------------------------------------
	fmt.Println("enrollID", enrollID)
	//----------------------------------------------------
	_, compositeKeyParts, err := stub.SplitCompositeKey(locationRange.Key)
	if err != nil {
		return shim.Error(err.Error())
	}
	keyAsString := compositeKeyParts[6]

	//---------------------------------------------------------
	logger.Info("key: ", keyAsString)
	//-------------------------------------------------------------
	idsAsBytes, _ := stub.GetState("ids")
	ids := IDs{}
	json.Unmarshal(idsAsBytes, &ids)
	BookingId, _ := strconv.Atoi(ids.BookingId)
	ids.BookingId = strconv.Itoa(BookingId + 1)
	//--------------------------------------------------------------
	fmt.Println("IDs after incerment of booking ID", ids)
	//--------------------------------------------------------------------
	idsAsBytes, _ = json.Marshal(ids)
	stub.PutState("ids", idsAsBytes)

	wrkSpaceAsBytes, _ := stub.GetState(keyAsString)
	wrkSpace := workspace{}
	json.Unmarshal(wrkSpaceAsBytes, &wrkSpace)
	//-------------------------------------------------------------
	logger.Info("workspace :", wrkSpace)
	//-----------------------------------------------------------------
	sch := schedule{}
	json.Unmarshal(args[2], &sch)
	sch.BookingStatus = "pending"
	BookingIdAsstring := strconv.Itoa(BookingId)
	sch.BookingTxID = stub.GetTxID()
	schAsbytes, _ := json.Marshal(sch)
	stub.PutState(BookingIdAsstring, schAsbytes)
	ush := UserSchedule{}
	ush.ScheduleId = BookingIdAsstring
	//---------------------------------------------------------
	fmt.Println("user schedule", ush)
	//----------------------------------------------------------
	ush.WorkSpaceId = keyAsString
	userAsbytes, _ := stub.GetState(enrollID)

	u := user{}
	json.Unmarshal(userAsbytes, &u)
	//--------------------------------------------------------------
	fmt.Println("user got from enroll ID at line 440", u)
	//---------------------------------------------------------------
	u.Calendar = append(u.Calendar, ush)
	userAsbytes, _ = json.Marshal(u)
	stub.PutState(enrollID, userAsbytes)
	//----------------------------------------------------------------
	fmt.Println("user after assigning schedule", u)
	//-----------------------------------------------------------------
	wsch := WSpaceSchedule{}
	wsch.UserId = enrollID
	wsch.ScheduleId = BookingIdAsstring
	wrkSpace.Requests = append(wrkSpace.Requests, wsch)

	//-------------------------------------------------------------------
	logger.Info("workspace after appending", wrkSpace)
	//------------------------------------------------------------

	wrkbytes, _ := json.Marshal(wrkSpace)
	stub.PutState(keyAsString, wrkbytes)
	wb, _ := stub.GetState(keyAsString)
	//-------------------------------------------------------------
	logger.Info("keyafter mofication of schedule,", keyAsString)
	//-----------------------------------------------------------------
	wrk := workspace{}
	json.Unmarshal(wb, &wrk)
	//----------------------------------------------
	logger.Info("wrksapce", wrk)
	//------------------------------------------------
	return shim.Success(nil)
}

func (s *SmartContract) QueryAllschedules(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	logger.Info("-------------------query All schedules ----------")

	atrrValue, _, _ := cid.GetAttributeValue(stub, "Role")
	logger.Info("Role ", atrrValue)

	if strings.Compare(strings.ToLower(string(atrrValue)), "manager") != 0 {
		//		logger.Info("acess denied")
		return shim.Error("Access denied")
	}
	startKey := "1001"
	endKey := "9999"
	var txid []string

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		sch := schedule{}
		json.Unmarshal(queryResponse.Value, &sch)
		if strings.Compare(sch.BookingStatus, "pending") == 0 {
			txid = append(txid, sch.BookingTxID)
		}
	}

	fmt.Printf("sch obj:\n%s\n", txid)
	txidAsBytes, _ := json.Marshal(txid)
	return shim.Success(txidAsBytes)
}

func (s *SmartContract) ApproveOrDeny(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	logger.Info("------------------- ApproveOrDeny----------")
	//args[0]=cubNo. arg[1]=schId arg[2]=uid arg[3]=yes/no
	var buffer bytes.Buffer
	//##################################################
	logger.Info("args received", args)
	atrrValue, _, _ := cid.GetAttributeValue(stub, "Role")
	logger.Info("er ", atrrValue)

	if strings.Compare(strings.ToLower(string(atrrValue)), "manager") != 0 {

		return shim.Error("Access denied")
	}

	wrkSpaceAsBytes, _ := stub.GetState(args[0])
	wrkSpace := workspace{}
	json.Unmarshal(wrkSpaceAsBytes, &wrkSpace)
	sch := schedule{}
	schAsBytes, _ := stub.GetState(args[1])
	json.Unmarshal(schAsBytes, &sch)
	//	-------------------------------------------------
	fmt.Println(sch.BookingStatus)
	//	------------------------------------------
	fmt.Println(strings.Compare(strings.ToLower(sch.BookingStatus), "pending"))
	//	================================================================================

	if strings.Compare(strings.ToLower(sch.BookingStatus), "pending") == 0 {
		sch.BookingStatus = args[3]
	} else {
		return shim.Error("Already " + sch.BookingStatus)
	}
	wrksch := WSpaceSchedule{}
	if strings.Compare(strings.ToLower(args[3]), "yes") == 0 {
		wrksch.ScheduleId = args[1]
		wrksch.UserId = args[2]
		buffer.WriteString("Approved succefully")
	}
	if strings.Compare(strings.ToLower(args[3]), "no") == 0 {
		buffer.WriteString("Denied succefully")
	}
	reqs := wrkSpace.Requests
	var index int
	var element WSpaceSchedule
	for index, element = range reqs {
		if strings.Compare(strings.ToLower(element.ScheduleId), args[1]) == 0 {
			reqs = append(reqs[:index], reqs[index+1:]...)
		}
	}

	wrkSpace.Requests = reqs
	wrkSpace.Schedule = append(wrkSpace.Schedule, element)
	wrkSpaceAsBytes, _ = json.Marshal(wrkSpace)
	stub.PutState(args[0], wrkSpaceAsBytes)

	schAsBytes, _ = json.Marshal(sch)
	stub.PutState(args[1], schAsBytes)
	//	=============================================
	fmt.Println("workspace", wrkSpace)
	//	================================================
	fmt.Println("sch", sch)
	//================================================
	return shim.Success(buffer.Bytes())

}

func (s *SmartContract) OccupyWorkSpace(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	logger.Info("-------------------OccupyWorkspace----------")

	//	schId=arg[0] userId=arg[1] curTime=arg[2] cubId=arg[3]
	sch := schedule{}
	logger.Info("args", args)
	schAsBytes, _ := stub.GetState(args[0])
	json.Unmarshal(schAsBytes, &sch)
	logger.Info("sch", sch)
	curTime, _ := strconv.Atoi(args[2])

	if sch.OccupiedTxID != "" {
		return shim.Error("Already Occupied with TxID" + sch.OccupiedTxID)
	}
	logger.Info("permission comparision")
	logger.Info(strings.Compare(strings.ToLower(sch.BookingStatus), "yes"))

	if strings.Compare(strings.ToLower(sch.BookingStatus), "yes") != 0 {
		return shim.Error("Operatio failed\nRequest Status: " + sch.BookingStatus)
	}
	// if sch.BookingTime-curTime > 600 {
	// 	return shim.Error("Try Before 10 minutes of schedule")
	// }
	if sch.EndTime <= curTime {
		return shim.Error("operation falied\n Booking schedule is Ended")
	}
	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}
	id := &mspprotos.SerializedIdentity{}
	err = proto.Unmarshal(creator, id)
	block, _ := pem.Decode(id.GetIdBytes())
	cert, err := x509.ParseCertificate(block.Bytes)
	enrollID := cert.Subject.CommonName

	if strings.Compare(args[1], enrollID) != 0 {
		shim.Error("Unknown user")
	}
	sch.OccupiedTxID = stub.GetTxID()
	schAsBytes, _ = json.Marshal(sch)
	stub.PutState(args[0], schAsBytes)
	usr := user{}
	fmt.Println("uid", args[1])
	usrAsBytes, _ := stub.GetState(args[1])
	json.Unmarshal(usrAsBytes, &usr)
	fmt.Println("user")
	fmt.Println(usr)
	org, _ := cid.GetMSPID(stub)
	conf := config{}

	confAsBytes, _ := stub.GetState(org)
	json.Unmarshal(confAsBytes, &conf)
	fmt.Println("conf")
	fmt.Println(conf)
	restrictedSites := []string{}
	restrictedSites = conf.BlackList
	RestrictedIP := []string{}
	RestrictedIP = conf.RestrictedIP

	WSAsBytes, _ := stub.GetState(args[3])
	WS := workspace{}
	json.Unmarshal(WSAsBytes, &WS)
	Ext_number := WS.NetAsset.Telephone.Extension_Number
	SwitchID := WS.EAsset.Switches[0].SwitchID
	resp := Response{}
	resp.RestrictedIP = RestrictedIP
	resp.RestrictedSites = restrictedSites
	resp.Extension_Number = Ext_number
	resp.SwitchID = SwitchID
	respAsbytes, _ := json.Marshal(resp)
	fmt.Println("resp", resp)
	return shim.Success(respAsbytes)
}
func (s *SmartContract) Switching(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	logger.Info("arguments received ", args)
	// switchID:=args[1]
	WSAsBytes, _ := stub.GetState(args[2])
	WS := workspace{}
	json.Unmarshal(WSAsBytes, &WS)
	// WS.EAsset.Switches[0].Status=args[0]
	for index, switches := range WS.EAsset.Switches {
		if strings.Compare(args[1], switches.SwitchID) == 0 {
			WS.EAsset.Switches[index].Status = args[0]
		}
	}

	WSAsBytes, _ = json.Marshal(WS)
	logger.Info("WS after updating switch State", WS)
	stub.PutState(args[2], WSAsBytes)
	return shim.Success(nil)

}
func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		logger.Error("Error starting Simple chaincode: %s", err)
	}
}
