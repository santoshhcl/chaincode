package main


import (
	"encoding/json"
	"fmt"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type InboxService struct {

}

type InboxRequest struct {
	CallingEntityName string `json:"callingEntityName"`
	InboxName         string `json:"inboxName"`
}

type InboxResponse struct {
	ShipmentWayBills []ShipmentWayBill `json:"shipmentWayBill"`
	EWWayBills []EWWayBill `json:"ewWaybill"`
}



func (t *InboxService) Inbox(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Inbox " + args[0])
	var pageLoadClass ShipmentPageLoadService
	var resp InboxResponse
	var err error
	var tmpEntity Entity
	
	request := parseInboxRequest(args[0])
	
	tmpEntity, err = pageLoadClass.fetchEntities(stub, request.CallingEntityName)
	if err != nil {
		resp.ShipmentWayBills = t.createShipmentArray(stub, tmpEntity, request.InboxName)
		resp.EWWayBills = t.createEWWayBillArray(stub, tmpEntity, request.InboxName)
	}
	return json.Marshal(resp)
}

func (t *InboxService) createShipmentArray(stub shim.ChaincodeStubInterface, tmpEntity Entity, inboxName string) ([]ShipmentWayBill) {
	var shipmentWayBillIndex ShipmentWayBillIndex
	var err error
	var shipmentWayBillArray []ShipmentWayBill

	shipmentWayBillIndex, err = fetchShipmentWayBillIndex(stub)

	lenOfArray := len(shipmentWayBillIndex.ShipmentNumber)

	for i := 0; i < lenOfArray; i++ {
		var tmpShipmentWayBill ShipmentWayBill
		tmpShipmentWayBill,err = fetchShipmentWayBillData(stub, shipmentWayBillIndex.ShipmentNumber[i])

		if(err != nil &&  t.checkInboxCondition(tmpEntity.EntityId, tmpEntity.EntityType, inboxName, tmpShipmentWayBill.Status, tmpShipmentWayBill.Consigner, tmpShipmentWayBill.Consignee, tmpShipmentWayBill.Carrier, tmpShipmentWayBill.CustodianHistory, tmpShipmentWayBill.Custodian) == "true") {
			shipmentWayBillArray = append(shipmentWayBillArray, tmpShipmentWayBill)
		}

	}

	return shipmentWayBillArray
}


func (t *InboxService) createEWWayBillArray(stub shim.ChaincodeStubInterface, tmpEntity Entity, inboxName string) ([]EWWayBill) {
	var allEWWayBillIndex AllEWWayBill
	var err error
	var shipmentWayBillArray []EWWayBill

	allEWWayBillIndex, err = fetchEWWayBillIndex(stub)

	lenOfArray := len(allEWWayBillIndex.AllWayBillNumber)

	for i := 0; i < lenOfArray; i++ {
		var tmpShipmentWayBill EWWayBill
		tmpShipmentWayBill,err = fetchEWWayBillData(stub, allEWWayBillIndex.AllWayBillNumber[i])

		if(err != nil &&  t.checkInboxCondition(tmpEntity.EntityId, tmpEntity.EntityType, inboxName, tmpShipmentWayBill.Status, tmpShipmentWayBill.Consigner, tmpShipmentWayBill.Consignee, "", tmpShipmentWayBill.CustodianHistory, tmpShipmentWayBill.Custodian) == "true") {
			shipmentWayBillArray = append(shipmentWayBillArray, tmpShipmentWayBill)
		}

	}

	return shipmentWayBillArray
}


func (t *InboxService) checkInboxCondition(entityId string, entityType string, inboxName string, status string, consignerName string,consigneeName string, carrier string, custodianHistory  []string,custodian string) (string) {
	var util Utility
	if entityType == "Manufacturer" {
		if (inboxName == "Created" && status == "ShipmentCreated" && consignerName == entityId){
			return "true"
		}
		if (inboxName == "InTransit" && status == "WaybillCreated" && consignerName == entityId){
			return "true"
		}
		if (inboxName == "Delivered" && status == "WaybillDelivered" && consignerName == entityId){
			return "true"
		}
		if (inboxName == "Cancelled" && status == "ShipmentCancelled" && consignerName == entityId){
			return "true"
		}
	}

	if entityType == "3PL" {
		if (inboxName == "Scheduled" && (status == "ShipmentCreated" || status == "DCShipmentCreated") && carrier == entityId){
			return "true"
		}
		if (inboxName == "InTransit" && (status == "WaybillCreated" || status == "DCWaybillCreated") && carrier == entityId){
			return "true"
		}
		if (inboxName == "Scheduled" && (status == "WaybillDelivered" || status == "DCWaybillDelivered") && carrier == entityId){
			return "true"
		}
	}

	if entityType == "DC" {
		if (inboxName == "Scheduled" && status == "WaybillCreated" && consigneeName == entityId){
			return "true"
		}
		if (inboxName == "Created" && status == "DCShipmentCreated" && consignerName == entityId){
			return "true"
		}
		if (inboxName == "InTransit" && status == "DCWaybillCreated" && (consignerName == entityId || consigneeName == entityId)){
			return "true"
		}
		if (inboxName == "Delivered" && ((status == "DCWaybillDelivered" && consignerName == entityId) || (status == "WaybillDelivered" && consigneeName == entityId))){
			return "true"
		}
		if (inboxName == "Cancelled" && status == "DCShipmentCancelled" && consignerName == entityId){
			return "true"
		}
	}

	if entityType == "warehouse" {	
		if (inboxName == "Scheduled" && ((status == "DCWaybillCreated" && entityId == custodian) || (status == "EWWaybillAtOCCargo" || consigneeName == entityId))){
			return "true"
		}
		if (inboxName == "Created" && status == "EWWaybillCreated" && consignerName == entityId){
			return "true"
		}

		if (inboxName == "InTransit" && (status == "EWWaybillAtCargo" || status == "EWWaybillAtVessel" || status == "EWWaybillAtOCCargo") && (consignerName == entityId || consigneeName == entityId)){
			return "true"
		}
		if (inboxName == "Delivered" && ((status == "EWWaybillDelivered" && consignerName == entityId) || (status == "EWWaybillDelivered" && consigneeName == entityId) || (status == "DCWaybillDelivered" && entityId == custodian))){
			return "true"
		}
		if (inboxName == "Cancelled" && status == "EWWaybillCancelled" && consignerName == entityId){
			return "true"
		}
	}	

	if entityType == "Cargo" {
		if (inboxName == "Arrived" && (status == "EWWaybillAtCargo" || status == "EWWaybillAtOCCargo") && entityId == custodian){
			return "true"
		}

		if (inboxName == "Delivered" && (status == "EWWaybillAtVessel" || status == "EWWaybillAtOCCargo" || status == "EWWaybillDelivered") && util.hasString(custodianHistory, entityId)){
			return "true"
		}
	}

	if entityType == "Vessel" {
		if (inboxName == "Arrived" && status == "EWWaybillAtVessel" && custodian == entityId){
			return "true"
		}

		if (inboxName == "Delivered" && (status == "EWWaybillAtOCCargo" || status == "EWWaybillDelivered") && util.hasString(custodianHistory, entityId)){
			return "true"
		}
	}

	return "false"
}


func parseInboxRequest(jsondata string) (InboxRequest) {
	fmt.Println("Entering parseInboxRequest " + jsondata)
	res := InboxRequest{}
	json.Unmarshal([]byte(jsondata), &res)
	
	fmt.Println("======================")
	fmt.Println(res)
	fmt.Println("======================")

	return res
}


func fetchShipmentWayBillIndex(stub shim.ChaincodeStubInterface) (ShipmentWayBillIndex, error) {
	var shipmentWayBill ShipmentWayBillIndex

	indexByte, err := stub.GetState("ShipmentWayBillIndex")
	if err != nil {
		fmt.Println("Could not retrive  Shipment WayBill ", err)
		return shipmentWayBill, err
	}

	if marshErr := json.Unmarshal(indexByte, &shipmentWayBill); marshErr != nil {
		fmt.Println("Could not retrieve Shipment WayBill from ledger", marshErr)
		return shipmentWayBill, marshErr
	}

	return shipmentWayBill, nil

}


func fetchEWWayBillIndex(stub shim.ChaincodeStubInterface) (AllEWWayBill, error) {
	var shipmentWayBill AllEWWayBill

	indexByte, err := stub.GetState("AllEWWayBill")
	if err != nil {
		fmt.Println("Could not retrive  Shipment WayBill ", err)
		return shipmentWayBill, err
	}

	if marshErr := json.Unmarshal(indexByte, &shipmentWayBill); marshErr != nil {
		fmt.Println("Could not retrieve Shipment WayBill from ledger", marshErr)
		return shipmentWayBill, marshErr
	}

	return shipmentWayBill, nil

}

