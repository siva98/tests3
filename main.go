//=================================================
//Account.go  包括创建学生 部门账户，更新密码，删除账号，登录的功能。
//
//=================================================

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("Account")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//====================================================================================
//jieyaojilu
//====================================================================================

type MoveInf struct {
	Admin2   string `json:"Admin2"`   // 管理员ID
	Student2 string `json:"Student2"` //学生ID
	Point    string `json:"Point"`    // 点数
	Password string `json:"Password"` // Password
	Message  string `json:"Message"`  // Message
}

type Transaction struct {
	Transaction2  string `json:"Transaction2"`  // Transaction Number
	AdminID       string `json:"AdminID"`       // 数字资产ID
	StudentID     string `json:"StudentID"`     // 管理员ID
	AdminPassword string `json:"AdminPassword"` //学生ID
	Money         string `json:"Money"`         // 点数
	Time          string `json:"Time"`          //交易时间
	message       string `json:"message"`       //Message
}

//====================================================================================
//学生角色
//====================================================================================
type Student struct {
	Xuehao   string `json:"xuehao"`
	Name     string `json:"name"`
	School   string `json:"school"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Money    string `json:"money"`
}

//====================================================================================
//部门角色
//====================================================================================
type Admin struct {
	Gonghao  string `json:"gonghao"`
	Name     string `json:"name"`
	School   string `json:"school"`
	Password string `json:"password"`
	Partment string `json:"partment"`
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### Account Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "CreateStudent" {
		return t.CreateStudent(stub, args)
	}
	if function == "CreateAdmin" {
		return t.CreateAdmin(stub, args)
	}
	if function == "StudentUpdatePassword" {
		return t.StudentUpdatePassword(stub, args)
	}
	if function == "AdminUpdatePassword" {
		return t.AdminUpdatePassword(stub, args)
	}

	if function == "QueryAccount" {
		return t.QueryAccount(stub, args)
	}
	if function == "DeleteStudent" {
		return t.DeleteStudent(stub, args)
	}
	if function == "DeleteAdmin" {
		return t.DeleteAdmin(stub, args)
	}
	if function == "loginAdmin" {
		return t.loginAdmin(stub, args)
	}
	if function == "loginStudent" {
		return t.loginStudent(stub, args)
	}
	if function == "movePoint" {
		return t.movePoint(stub, args)
	}
	if function == "getHistoryForKey" {
		return t.getHistoryForKey(stub, args)
	}
	if function == "CreatCredit" {
		return t.CreatCredit(stub, args)
	}
	logger.Errorf("Unknown action, check the first argument, must be one of 'CreateAccount', 'ChangePassword', 'Verify', 'DeleteAccount', or 'QueryAccount'. But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'CreateAccount', 'ChangePassword', 'Verify', 'DeleteAccount', or 'QueryAccount'. But got: %v", args[0]))
}

//============================================================================================================
//创建学生账号
//============================================================================================================
func (t *SimpleChaincode) CreateStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("CreateStudent")

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	var student Student
	student.Xuehao = args[0]
	student.Name = args[1]
	student.Password = args[2]
	student.School = args[3]
	student.Email = args[4]
	student.Money = args[5]

	Bytes, _ := json.Marshal(student)

	// ==== Check if account already exists ====
	bytes, err := stub.GetState(student.Xuehao)
	if err != nil {
		return shim.Error("Failed to get this student: " + err.Error())
	}
	if bytes != nil {
		return shim.Error("This student already exists: " + student.Xuehao)
	}

	//create
	err = stub.PutState(student.Xuehao, Bytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(Bytes))

}

//============================================================================================================
//创建部门账号
//============================================================================================================
func (t *SimpleChaincode) CreateAdmin(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("CreateAdmin")

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	var admin Admin
	admin.Gonghao = args[0]
	admin.Name = args[1]
	admin.Password = args[2]
	admin.School = args[3]
	admin.Partment = args[4]

	Bytes, _ := json.Marshal(admin)

	// ==== Check if account already exists ====
	bytes, err := stub.GetState(admin.Gonghao)
	if err != nil {
		return shim.Error("Failed to get this admin: " + err.Error())
	}
	if bytes != nil {
		return shim.Error("This admin already exists: " + admin.Gonghao)
	}

	//create
	err = stub.PutState(admin.Gonghao, Bytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(Bytes))

}

//======================================================================================================
// 更改学生密码
// args: 学号|原密码|新密码
//======================================================================================================
func (t *SimpleChaincode) StudentUpdatePassword(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("UpdatePassword")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	//var account Account
	accountID := args[0]       //学号
	accountPassword := args[1] //旧密码
	newPassword := args[2]     //新密码
	var err error

	Bytes, _ := stub.GetState(accountID)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	if Bytes == nil {
		return shim.Error("This account does not exists: " + accountID)
	}
	var student Student

	err = json.Unmarshal(Bytes, &student)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	if student.Password == accountPassword {
		student.Password = newPassword
	} else {
		return shim.Error("wrong password")
	}
	bytes, _ := json.Marshal(student)
	err = stub.PutState(accountID, bytes) //这个地方会报错？？是否需要重新putstate?
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("{\"updatepassword\":\"sucessful\"}"))

}

//======================================================================================================
// 更改部门密码
// args: 工号|原密码|新密码
//======================================================================================================
func (t *SimpleChaincode) AdminUpdatePassword(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("UpdatePassword")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	//var account Account
	accountID := args[0]       //工号
	accountPassword := args[1] //旧密码
	newPassword := args[2]     //新密码
	var err error

	Bytes, _ := stub.GetState(accountID)
	if err != nil {
		return shim.Error("{\"Failed to get account\": \"" + err.Error() + "\"}")
	}
	if Bytes == nil {
		return shim.Error("This accountt does not exists: " + accountID)
	}
	var admin Admin

	err = json.Unmarshal(Bytes, &admin)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	if admin.Password == accountPassword {
		admin.Password = newPassword
	} else {
		return shim.Error("{\"result\":\"wrong password\"}")
	}

	bytes, _ := json.Marshal(admin)
	err = stub.PutState(accountID, bytes) //这个地方会报错？？是否需要重新putstate?
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("{\"updatepassword\":\"sucessful\"}"))

}

//==============================================================================
// 查询账户是否存在
// args: ID
//==============================================================================
func (t *SimpleChaincode) QueryAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	accountID := args[0]

	// Get the state from the ledger
	bytes, err := stub.GetState(accountID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + accountID + "\"}"
		return shim.Error(jsonResp)
	}
	if bytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + accountID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(bytes)
}

//=============================================================================================
//zhuanyizixhan simple
//args:
//=============================================================================================
func (t *SimpleChaincode) movePoint(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var transaction MoveInf
	var err error
	transaction.Admin2 = args[0]
	transaction.Student2 = args[1]
	transaction.Password = args[3]
	transaction.Message = args[4]
	accountPassword := args[3]
	//transaction.Point, err = strconv.Atoi(args[2])

	// ==== Check if Seller exists ====
	bytesAdmin, err := stub.GetState(transaction.Admin2)
	if err != nil {
		return shim.Error("Failed to get Seller: " + err.Error())
	}
	if bytesAdmin == nil {
		return shim.Error("This Admin not exists: ")
	}
	var admin Admin

	err = json.Unmarshal(bytesAdmin, &admin)
	if err != nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Fail to get Admin Account \"}")
	}
	if admin.Password == accountPassword {
		// ==== Check if Student exists ====
		bytesStudent, err := stub.GetState(transaction.Student2)
		if err != nil {
			return shim.Error("Failed to get Student: " + err.Error())
		}
		if bytesStudent == nil {
			return shim.Error("This Student not exists: ")
		}
		var digitalStudent Student
		err = json.Unmarshal(bytesStudent, &digitalStudent)
		if err != nil {
			return shim.Error("Failed to get Student: " + err.Error())
		}
		// ==== Check if Point is a integer ====

		// ==== Move Action ====

		//digitalStudent.Point = digitalStudent.Point + transaction.Point

		var s int
		s1, err := strconv.Atoi(digitalStudent.Money)
		s2, err := strconv.Atoi(args[2])
		s = s1 + s2
		digitalStudent.Money = strconv.Itoa(s) // must change into int????

		DigitalStudentBytes, _ := json.Marshal(digitalStudent)
		err = stub.PutState(transaction.Student2, []byte(DigitalStudentBytes))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte("{\"Result\":\"MovePointSuccess\",\"message\":{" + args[4] + "}}"))
	}
	return shim.Error("\"Result\":\"fail\",\"Message\":\"Incorrect password\"")
}

func (t *SimpleChaincode) CreatCredit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 6 {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Incorrect number of arguments. Expecting 7 \"}")
	}

	var transaction Transaction
	var err error
	transaction.Transaction2 = args[0]
	transaction.AdminID = args[1]
	transaction.StudentID = args[2]
	transaction.Money = args[3]
	transaction.Time = args[4]
	transaction.message = args[5]

	// ==== Check if admin exists ====
	bytesAdmin, err := stub.GetState(transaction.AdminID)
	if err != nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Failed to get Admin \"}")
	}
	if bytesAdmin == nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"This Admin not exists \"}")
	}
	var admin Admin

	err = json.Unmarshal(bytesAdmin, &admin)
	if err != nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Fail to get Admin Account \"}")
	}

	// ==== Check if Student exists ====
	bytesStudent, err := stub.GetState(transaction.StudentID)
	if err != nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Fail to get Student\"}")
	}
	if bytesStudent == nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Fail to get Student Account\"}")
	}
	var student Student

	err = json.Unmarshal(bytesStudent, &student)
	if err != nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Fail to get Account \"" + err.Error() + "\"}")
	}
	// ==== Check if Point is a integer ====

	// ==== Move Action ====
	/*var s int
	s1, err := strconv.Atoi(student.Money)
	if err != nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Not number\"}")
	}
	s2, err1 := strconv.Atoi(args[4])
	if err1 != nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Not number\"}")
	}
	s = s1 + s2
	student.Money = strconv.Itoa(s) // must change into int????
	DigitalStudentBytes, _ := json.Marshal(student)
	err = stub.PutState(transaction.StudentID, []byte(DigitalStudentBytes))
	if err != nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Fail to get Student ID\"" + err.Error() + "\"}")
	}
	//return shim.Success([]byte("Success move point"))
	*/
	// ======Creat train Message=========
	DigitaltransactionBytes, _ := json.Marshal(transaction)
	bytes, err := stub.GetState(transaction.Transaction2)
	if err != nil {
		return shim.Error("Failed to get transaction: " + err.Error())
	}
	if bytes == nil {

	}

	err = stub.PutState(transaction.Transaction2, []byte(DigitaltransactionBytes))
	if err != nil {
		return shim.Error("{\"Result\":\"fail\",\"Message\":\"Fail to creat Transaction Message\"}")
	}
	return shim.Success([]byte("{\"Result\":\"CreatCrditsuccess\",\"Message\":\"Success to creat Transaction Message\"}"))

	return shim.Error("{\"Result\":\"fail\",\"Message\":\"Wrong Password\"}")

}

//=============================================================================================
//删除学生
//args: ID
//=============================================================================================
func (t *SimpleChaincode) DeleteStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("DeleteAccount")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	accountID := args[0]
	bytes, err := stub.GetState(accountID)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	if bytes == nil {
		return shim.Error("this account is not found")
	}

	var student Student
	err = json.Unmarshal(bytes, &student)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}

	student = Student{} //delete the struct
	bytes, err = json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(accountID, bytes)

	err = stub.DelState(accountID)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	bytes, err = stub.GetState(accountID)
	if err != nil {
		return shim.Success([]byte("{\"delete \":\"delete sucessful\"}"))
	}

	return shim.Success([]byte("{\"delete \":\"delete sucessful\"}")) //
}

//=============================================================================================
//删除学生
//args: ID
//=============================================================================================
func (t *SimpleChaincode) DeleteAdmin(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("DeleteAccount")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	accountID := args[0]
	bytes, err := stub.GetState(accountID)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	if bytes == nil {
		return shim.Error("this account is not found")
	}

	var admin Admin
	err = json.Unmarshal(bytes, &admin)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}

	admin = Admin{} //delete the struct
	bytes, err = json.Marshal(admin)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(accountID, bytes)

	err = stub.DelState(accountID)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	bytes, err = stub.GetState(accountID)
	if err != nil {
		return shim.Success([]byte("{\"delete \":\"delete sucessful\"}"))
	}

	return shim.Success([]byte("{\"delete \":\"delete sucessful\"}"))
}

//================
//验证Student账号密码是否匹配 args:ID|Password
//================
func (t *SimpleChaincode) loginStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	accountID := args[0]
	password := args[1]

	//query the ledger
	bytes, err := stub.GetState(accountID)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	if bytes == nil {
		return shim.Error("This account does not exists: " + accountID)
	}
	var student Student
	err = json.Unmarshal(bytes, &student)
	if err != nil {
		return shim.Error("Failed to get Student account: " + err.Error())
	}
	if student.Password == password {
		return shim.Success([]byte("{\"login\":\"loginSuccess\"}"))
	} else {
		return shim.Error("{\"login\":\"wrong password\"}")
	}
}

//================
//验证Admin账号密码是否匹配 args:ID|Password
//================
func (t *SimpleChaincode) loginAdmin(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	accountID := args[0]
	password := args[1]

	//query the ledger
	bytes, err := stub.GetState(accountID)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	}
	if bytes == nil {
		return shim.Error("This account does not exists: " + accountID)
	}
	var admin Admin
	err = json.Unmarshal(bytes, &admin)
	if err != nil {
		return shim.Error("Failed to get admin account: " + err.Error())
	}
	if admin.Password == password {
		return shim.Success([]byte("{\"login\":\"loginSuccess\"}"))
	} else {
		return shim.Error("{\"login\":\"wrong password\"}")
	}
}

func (t *SimpleChaincode) getHistoryForKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2,function followed by 1 accountID and 1 value")
	}

	var accountID string //Entities
	var err error
	accountID = args[0]
	//Get the state from the ledger
	//TODD:will be nice to have a GetAllState call to ledger
	HisInterface, err := stub.GetHistoryForKey(accountID)
	fmt.Println(HisInterface)
	Avalbytes, err := getHistoryListResult(HisInterface)
	if err != nil {
		return shim.Error("Failed to get history")
	}
	return shim.Success([]byte(Avalbytes))
}

func getHistoryListResult(resultsIterator shim.HistoryQueryIteratorInterface) ([]byte, error) {

	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		item, _ := json.Marshal(queryResponse)
		buffer.Write(item)
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("queryResult:\n%s\n", buffer.String())
	return buffer.Bytes(), nil
}

//=======================================================================================
//main function
//=================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
