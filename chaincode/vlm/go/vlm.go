/*
 * This the smart contract for vehicle lifetime management
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// SmartContract structure
type SmartContract struct {
}

// CarStruct structure
type CarStruct struct {
	ChassisNo             string `json:"chassisNo"`
	Owner                 string `json:"owner"`
	RegistrationNo        string `json:"registrationNo"`
	RegistrationExpiryDae string `json:"registrationExpiryDae"`
	Status                string `json:"status"`
}

// Init SmartContract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke SmartContract Invoke
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "createCar" {
		return s.createCar(APIstub, args)
	} else if function == "transferCar" {
		return s.transferCar(APIstub, args)
	} else if function == "sellnRegisterCar" {
		return s.sellnRegisterCar(APIstub, args)
	} else if function == "scrapCar" {
		return s.scrapCar(APIstub, args)
	} else if function == "getCar" {
		return s.getCar(APIstub, args)
	} else if function == "getCarHistory" {
		return s.getCarHistory(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

// createCar - This is for manufacture to create cars.
func (s *SmartContract) createCar(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	chasisNo := args[0]

	Car := CarStruct{ChassisNo: chasisNo,
		Owner:                 "Maruti",
		RegistrationNo:        "",
		RegistrationExpiryDae: "",
		Status:                "New"}
	CarBytes, err := json.Marshal(Car)
	if err != nil {
		return shim.Error("JSON Marshal failed.")
	}

	APIstub.PutState(chasisNo, CarBytes)
	fmt.Println("Car Created -> ", Car)

	return shim.Success(nil)
}

// transferCar - This is for manufacture to transfer the cars dealer.
func (s *SmartContract) transferCar(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	chasisNo := args[0]
	owner := args[1]

	CarAsBytes, _ := APIstub.GetState(chasisNo)

	var car CarStruct

	err := json.Unmarshal(CarAsBytes, &car)
	if err != nil {
		return shim.Error("Issue with Car json unmarshaling")
	}

	if car.Status != "New" {
		return shim.Error("Only new car transfer is allowed")
	}

	Car := CarStruct{ChassisNo: car.ChassisNo,
		Owner:                 owner,
		RegistrationNo:        "",
		RegistrationExpiryDae: "",
		Status:                "Dealer"}

	CarBytes, err := json.Marshal(Car)
	if err != nil {
		return shim.Error("Issue with Car json marshaling")
	}

	APIstub.PutState(Car.ChassisNo, CarBytes)
	fmt.Println("Car trasnferred to dealer -> ", Car)

	return shim.Success(nil)
}

// sellnRegisterCar - This is for dealers to sell the cars to customer.
func (s *SmartContract) sellnRegisterCar(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	chasisNo := args[0]
	owner := args[1]
	registrationNo := args[2]
	registrationExpiryDae := args[3]

	CarAsBytes, _ := APIstub.GetState(chasisNo)
	var car CarStruct

	err := json.Unmarshal(CarAsBytes, &car)
	if err != nil {
		return shim.Error("Issue with Car json unmarshaling")
	}

	if car.Status == "Scrapped" {
		return shim.Error("Cannot sell Car that is already scrapped")
	}

	Car := CarStruct{ChassisNo: car.ChassisNo,
		Owner:                 owner,
		RegistrationNo:        registrationNo,
		RegistrationExpiryDae: registrationExpiryDae,
		Status:                "Customer"}

	CarBytes, err := json.Marshal(Car)
	if err != nil {
		return shim.Error("Issue with Car json marshaling")
	}

	APIstub.PutState(Car.ChassisNo, CarBytes)
	fmt.Println("Car sold to customer -> ", Car)

	return shim.Success(nil)
}

// scrapCar - This is for customer to scarp the cars.
func (s *SmartContract) scrapCar(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	chasisNo := args[0]

	CarAsBytes, _ := APIstub.GetState(chasisNo)
	var car CarStruct

	err := json.Unmarshal(CarAsBytes, &car)
	if err != nil {
		return shim.Error("Issue with Car json unmarshaling")
	}

	Car := CarStruct{ChassisNo: car.ChassisNo,
		Owner:                 car.Owner,
		RegistrationNo:        car.RegistrationNo,
		RegistrationExpiryDae: car.RegistrationExpiryDae,
		Status:                "Scrapped"}

	CarBytes, err := json.Marshal(Car)
	if err != nil {
		return shim.Error("Issue with Car json marshaling")
	}

	APIstub.PutState(Car.ChassisNo, CarBytes)
	fmt.Println("Car scrapped -> ", Car)

	return shim.Success(nil)
}

func (s *SmartContract) getCar(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	chasisNo := args[0]
	CarAsBytes, _ := APIstub.GetState(chasisNo)
	return shim.Success(CarAsBytes)
}

func (s *SmartContract) getCarHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	chasisNo := args[0]

	resultsIterator, err := APIstub.GetHistoryForKey(chasisNo)
	if err != nil {
		return shim.Error("Error retrieving Car history with GetHistoryForKey")
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the car
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Error retrieving next Car history.")
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getCarHistory returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// Main function is only relevant in unit test mode. Only included here for completeness.
func main() {
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
