package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const NODATA_ERROR_CODE string = "400"

const NODATA_ERROR_MSG string = "No data found"

const INVALID_INPUT_ERROR_CODE string = "401"
const INVALID_INPUT_ERROR_MSG string = "Invalid Input"

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

const (
	SHIPMENT   = "SHIPMENT"
	WAYBILL    = "WAYBILL"
	DCSHIPMENT = "DCSHIPMENT"
	DCWAYBILL  = "DCWAYBILL"
	EWWAYBILL  = "EWWAYBILL"
)

type B4SCChaincode struct {
}

//////////////////////////@@@@@@@@@@@@@@@@@  santosh compliance document   @@@@@@@@@@@@@@@\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\
//storing compliance document mdetadata and hash
type ComplianceDocument struct {
	compliance_id      string
	manufacturer       string
	regulator          string
	documentTitle      string
	document_mime_type string
	documentHash       string
	documentType       string
	createdBy          string
	createdDate        string
}

//mapping for entity and corresponding document
type EntityComplianceDocMapping struct {
	complianceIds []string
}

//collection of all the compliance document ids
type ComplianceIds struct {
	complianceIds []string
}

//list of compliance document
type ComplianceDocumentList struct {
	complianceDocumentList []ComplianceDocument
}

//method for storing complaince document metadata and hash
func uploadComplianceDocument(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	resp := BlockchainResponse{}
	fmt.Println("uploading compliance document", args[0])
	compDoc, _ := parseComplianceDocument(args[0])
	fmt.Println("uploading compliance document", compDoc)
	complianceId := compDoc.compliance_id
	saveErr := saveComplianceDocument(stub, complianceId, compDoc)
	if saveErr != nil {
		resp.Err = "000"
		resp.ErrMsg = complianceId
		resp.Message = "Document Not saved"
		respString, _ := json.Marshal(resp)
		return []byte(respString), saveErr
	}
	entityCompMapRequest := EntityComplianceDocMapping{}
	entityCompMap, err := fetchEntityComplianceDocumentMapping(stub, compDoc.manufacturer)
	if err != nil {
		entityCompMapRequest.complianceIds = append(entityCompMapRequest.complianceIds, complianceId)
		saveEntityComplianceDocumentMapping(stub, entityCompMapRequest, compDoc.manufacturer)
	} else {
		entityCompMapRequest.complianceIds = append(entityCompMap.complianceIds, complianceId)
		fmt.Println("Updated entity compliance document mapping", entityCompMapRequest)
		saveEntityComplianceDocumentMapping(stub, entityCompMapRequest, compDoc.manufacturer)
	}
	complianceidsRequest := ComplianceIds{}
	complianceids, err := fetchComplianceDocumentIds(stub, "CompDocIDs")
	if err != nil {
		complianceidsRequest.complianceIds = append(complianceidsRequest.complianceIds, complianceId)
		saveComplianceDocumentIds(stub, complianceidsRequest)
	} else {
		complianceidsRequest.complianceIds = append(complianceids.complianceIds, complianceId)
		fmt.Println("Updated entity compliance document mapping", entityCompMapRequest)
		saveComplianceDocumentIds(stub, complianceidsRequest)
	}
	if err != nil {
		fmt.Println("Could not uploaded compliance document", err)
		return nil, err
	}
	resp.Err = "200"
	resp.ErrMsg = "Data Saved"
	resp.Message = "Successfully uploaded compliance document to ledger"
	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully uploaded compliance document to ledger")
	return []byte(respString), nil
}

//save entity compliance document mapping in blockchain
func saveEntityComplianceDocumentMapping(stub shim.ChaincodeStubInterface, entityCompMapRequest EntityComplianceDocMapping, entityname string) ([]byte, error) {
	dataToStore, _ := json.Marshal(entityCompMapRequest)
	entitykey := entityname + "ComDoc"
	err := stub.PutState(entitykey, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Entity compliance Mapping to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = entityname

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Entity WayBill Mapping")
	return []byte(respString), nil

}

//save compliance document ids in blockchain
func saveComplianceDocumentIds(stub shim.ChaincodeStubInterface, comids ComplianceIds) ([]byte, error) {
	dataToStore, _ := json.Marshal(comids)
	err := stub.PutState("CompDocIDs", []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save complianceIds to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = "CompDocIDs"
	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved compliance IDs")
	return []byte(respString), nil

}

//get entity name from compliance Document json
func parseComplianceDocument(jsonComDoc string) (ComplianceDocument, error) {

	complianceDoc := ComplianceDocument{}

	if marshErr := json.Unmarshal([]byte(jsonComDoc), &complianceDoc); marshErr != nil {
		fmt.Println("Could not Unmarshal compliance Document", marshErr)
		return complianceDoc, marshErr
	}
	fmt.Println("Unmarshal compliance Document", complianceDoc)
	return complianceDoc, nil
}

//save compliance document to blockchain
func saveComplianceDocument(stub shim.ChaincodeStubInterface, complianceId string, compDoc ComplianceDocument) error {
	dataToStore, _ := json.Marshal(compDoc)
	fmt.Println("compliance id ....", complianceId)
	err := stub.PutState(complianceId, []byte(dataToStore))
	if err != nil {
		fmt.Println("compliance document not uploaded to ledger", err)
		return err
	}
	return err
}

//fetch entity compliance document mapping
func fetchEntityComplianceDocumentMapping(stub shim.ChaincodeStubInterface, entityname string) (EntityComplianceDocMapping, error) {
	entityComplianceDocMapping := EntityComplianceDocMapping{}
	entitykey := entityname + "ComDoc"
	indexByte, err := stub.GetState(entitykey)
	if err != nil {
		fmt.Println("Could not retrive entity compliance mapping ", err)
		return entityComplianceDocMapping, err
	}

	if marshErr := json.Unmarshal(indexByte, &entityComplianceDocMapping); marshErr != nil {
		fmt.Println("Could not retrive entity compliance mapping from ledger", marshErr)
		return entityComplianceDocMapping, marshErr
	}

	return entityComplianceDocMapping, nil

}

//fetch compliance ids collection
func fetchComplianceDocumentIds(stub shim.ChaincodeStubInterface, compkey string) (ComplianceIds, error) {
	complianceids := ComplianceIds{}
	indexByte, err := stub.GetState(compkey)
	if err != nil {
		fmt.Println("Could not retrive complianceids", err)
		return complianceids, err
	}

	if marshErr := json.Unmarshal(indexByte, &complianceids); marshErr != nil {
		fmt.Println("Could not retrive complianceids from ledger", marshErr)
		return complianceids, marshErr
	}

	return complianceids, nil

}

func getComplianceDocumentByEntityName(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	complianceDocumentList := ComplianceDocumentList{}
	entityComplianceMapping, err := fetchEntityComplianceDocumentMapping(stub, args[0])
	if err != nil {
		return nil, nil
	} else {
		iterator := len(entityComplianceMapping.complianceIds)
		for i := 0; i < iterator; i++ {
			complianceDocuments, _ := fetchComplianceDocumentByComplianceId(stub, entityComplianceMapping.complianceIds[i])
			complianceDocumentList.complianceDocumentList = append(complianceDocumentList.complianceDocumentList, complianceDocuments)
		}
		dataToReturn, _ := json.Marshal(complianceDocumentList)
		return []byte(dataToReturn), nil
	}
	return nil, nil
}
func fetchComplianceDocumentByComplianceId(stub shim.ChaincodeStubInterface, complianceid string) (ComplianceDocument, error) {
	complianceDocument := ComplianceDocument{}
	indexByte, err := stub.GetState(complianceid)
	if err != nil {
		fmt.Println("Could not retrive compliance document", err)
		return complianceDocument, err
	}

	if marshErr := json.Unmarshal(indexByte, &complianceDocument); marshErr != nil {
		fmt.Println("Could not retrive complianceids from ledger", marshErr)
		return complianceDocument, marshErr
	}

	return complianceDocument, nil

}

func getAllComplianceDocument(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	complianceDocumentList := ComplianceDocumentList{}
	complianceIds, err := fetchComplianceDocumentIds(stub, args[0])
	if err != nil {
		return nil, nil
	} else {
		iterator := len(complianceIds.complianceIds)
		for i := 0; i < iterator; i++ {
			complianceDocuments, _ := fetchComplianceDocumentByComplianceId(stub, complianceIds.complianceIds[i])
			complianceDocumentList.complianceDocumentList = append(complianceDocumentList.complianceDocumentList, complianceDocuments)
		}
		dataToReturn, _ := json.Marshal(complianceDocumentList)
		return []byte(dataToReturn), nil
	}
	return nil, nil
}

///////////////////////////////////////////////////////end compliance docuent \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

//custom data models

type Pallet struct {
	PalletId    string
	Modeltype   string
	CartonId    []string
	ShipmentIds []string
}

type Carton struct {
	CartonId    string
	PalletId    string
	AssetId     []string
	ShipmentIds []string
}

type Asset struct {
	AssetId     string
	Modeltype   string
	Color       string
	CartonId    string
	PalletId    string
	ShipmentIds []string
}

type WayBillHistory struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
	Notes     string  `json:"notes"`
	Lat       float64 `json:"lat"`
	Log       float64 `json:"log"`
}

type Shipment struct {
	ShipmentNumber        string           `json:"shipmentNumber"`
	WayBillNo             string           `json:"wayBillNo"`
	WayBillType           string           `json:"wayBillType"`
	PersonConsigningGoods string           `json:"personConsigningGoods"`
	Consigner             string           `json:"consigner"`
	ConsignerAddress      string           `json:"consignerAddress"`
	Consignee             string           `json:"consignee"`
	ConsigneeAddress      string           `json:"consigneeAddress"`
	ConsigneeRegNo        string           `json:"consigneeRegNo"`
	Quantity              string           `json:"quantity"`
	Pallets               []string         `json:"pallets"`
	Cartons               []string         `json:"cartons"`
	Status                string           `json:"status"`
	ModelNo               string           `json:"modelNo"`
	VehicleNumber         string           `json:"vehicleNumber"`
	VehicleType           string           `json:"vehicleType"`
	PickUpTime            string           `json:"pickUpTime"`
	ValueOfGoods          string           `json:"valueOfGoods"`
	ContainerId           string           `json:"containerId"`
	MasterWayBillRef      []string         `json:"masterWayBillRef"`
	WayBillHistorys       []WayBillHistory `json:"wayBillHistorys"`
	Carrier               string           `json:"carrier"`
	Acl                   []string         `json:"acl"`
	CreatedBy             string           `json:"createdBy"`
	Custodian             string           `json:"custodian"`
	CreatedTimeStamp      string           `json:"createdTimeStamp"`
	UpdatedTimeStamp      string           `json:"updatedTimeStamp"`
}

type ShipmentIndex struct {
	ShipmentNumber string
	Status         string
	Acl            []string
}

type AllShipment struct {
	ShipmentIndexArr []ShipmentIndex
}

type AllShipmentDump struct {
	ShipmentIndexArr []string `json:"shipmentIndexArr"`
}

type Entity struct {
	EntityId        string  `json:"entityId"`
	EntityName      string  `json:"entityName"`
	EntityType      string  `json:"entityType"`
	EntityAddress   string  `json:"entityAddress"`
	EntityRegNumber string  `json:"entityRegNumber"`
	EntityCountry   string  `json:"entityCountry"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

//Will be avlable in the WorldStats as "ALL_ENTITIES"
type AllEntities struct {
	EntityArr []string `json:"entityArr"`
}

//Will be avlable in the WorldStats as "ASSET_MODEL_NAMES"
type AssetModelDetails struct {
	ModelNames []string `json:"modelNames"`
}

type WorkflowDetails struct {
	FromEntity  string   `json:"fromEntity"`
	ToEntity    string   `json:"toEntity"`
	Carrier     string   `json:"carrier"`
	EntityOrder []string `json:"entityOrder"`
}

//Will be available in the WorldStats as "ALL_WORKFLOWS"
type AllWorkflows struct {
	Workflows []WorkflowDetails `json:"workflows"`
}

/************** Arshad Start Code This new struct for AssetDetails , CartonDetails , PalletDetails  is added by Arshad as to incorporate new LLD published orginal structure
are not touched as of now to avoid break of any functionality devloped by Kartik 20/3/2017***************/

type BlockchainResponse struct {
	Err     string `json:"err"`
	ErrMsg  string `json:"errMsg"`
	Message string `json:"message"`
}

type AssetDetails struct {
	AssetSerialNo      string
	AssetModel         string
	AssetType          string
	AssetMake          string
	AssetCOO           string
	AssetMaufacture    string
	AssetStatus        string
	CreatedBy          string
	CreatedDate        string
	ModifiedBy         string
	ModifiedDate       string
	PalletSerialNumber string
	CartonSerialNumber string
	MshipmentNumber    string
	DcShipmentNumber   string
	MwayBillNumber     string
	DcWayBillNumber    string
	EwWayBillNumber    string
	MShipmentDate      string
	DcShipmentDate     string
	MWayBillDate       string
	DcWayBillDate      string
	EwWayBillDate      string
}

type CartonDetails struct {
	CartonSerialNo     string
	CartonModel        string
	CartonStatus       string
	CartonCreationDate string
	PalletSerialNumber string
	AssetsSerialNumber []string
	MshipmentNumber    string
	DcShipmentNumber   string
	MwayBillNumber     string
	DcWayBillNumber    string
	EwWayBillNumber    string
	Dimensions         string
	Weight             string
	MShipmentDate      string
	DcShipmentDate     string
	MWayBillDate       string
	DcWayBillDate      string
	EwWayBillDate      string
}

type PalletDetails struct {
	PalletSerialNo     string
	PalletModel        string
	PalletStatus       string
	CartonSerialNumber []string
	PalletCreationDate string
	AssetsSerialNumber []string
	MshipmentNumber    string
	DcShipmentNumber   string
	MwayBillNumber     string
	DcWayBillNumber    string
	EwWayBillNumber    string
	Dimensions         string
	Weight             string
	MShipmentDate      string
	DcShipmentDate     string
	MWayBillDate       string
	DcWayBillDate      string
	EwWayBillDate      string
}

/*
type CreatePalletDetailsResponse struct {
	Err     string `json:"err"`
	ErrMsg  string `json:"errMsg"`
	Message string `json:"message"`
}
type CreatePalletDetailsRequest struct {
	PalletSerialNo     string
	PalletModel        string
	PalletStatus       string
	CartonSerialNumber []string
	PalletCreationDate string
	AssetsSerialNumber []string
	MshipmentNumber    string
	DcShipmentNumber   string
	MwayBillNumber     string
	DcWayBillNumber    string
	EwWayBillNumber    string
	Dimensions         string
	Weight             string
	MShipmentDate      string
	DcShipmentDate     string
	MWayBillDate       string
	DcWayBillDate      string
	EwWayBillDate      string
}*/

//Will be avlable in the WorldStats as "ShipmentWayBillIndex"
type ShipmentWayBillIndex struct {
	ShipmentNumber []string
}

//Will be avlable in the WorldStats as "WayBillNumberIndex"
type WayBillNumberIndex struct {
	WayBillNumber []string
}

/*This is common struct across Shipment and Waybill*/
type ShipmentWayBill struct {
	WayBillNumber         string   `json:"wayBillNumber"`
	ShipmentNumber        string   `json:"shipmentNumber"`
	CountryFrom           string   `json:"countryFrom"`
	CountryTo             string   `json:"countryTo"`
	Consigner             string   `json:"consigner"`
	Consignee             string   `json:"consignee"`
	Custodian             string   `json:"custodian"`
	CustodianHistory      []string `json:"custodianHistory"`
	PersonConsigningGoods string   `json:"personConsigningGoods"`
	Comments              string   `json:"comments"`
	TpComments            string   `json:"tpComments"`
	VehicleNumber         string   `json:"vehicleNumber"`
	VehicleType           string   `json:"vehicleType"`
	PickupDate            string   `json:"pickupDate"`
	PalletsSerialNumber   []string `json:"palletsSerialNumber"`
	AddressOfConsigner    string   `json:"addressOfConsigner"`
	AddressOfConsignee    string   `json:"addressOfConsignee"`
	ConsignerRegNumber    string   `json:"consignerRegNumber"`
	Carrier               string   `json:"carrier"`
	VesselType            string   `json:"vesselType"`
	VesselNumber          string   `json:"vesselNumber"`
	ContainerNumber       string   `json:"containerNumber"`
	ServiceType           string   `json:"serviceType"`
	ShipmentModel         string   `json:"shipmentModel"`
	PalletsQuantity       string   `json:"palletsQuantity"`
	CartonsQuantity       string   `json:"cartonsQuantity"`
	AssetsQuantity        string   `json:"assetsQuantity"`
	ShipmentValue         string   `json:"shipmentValue"`
	EntityName            string   `json:"entityName"`
	ShipmentCreationDate  string   `json:"shipmentCreationDate"`
	EWWayBillNumber       string   `json:"eWWayBillNumber"`
	SupportiveDocuments   []string `json:"supportiveDocuments"`
	ShipmentCreatedBy     string   `json:"shipmentCreatedBy"`
	ShipmentModifiedDate  string   `json:"shipmentModifiedDate"`
	ShipmentModifiedBy    string   `json:"shipmentModifiedBy"`
	WayBillCreationDate   string   `json:"wayBillCreationDate"`
	WayBillCreatedBy      string   `json:"wayBillCreatedBy"`
	WayBillModifiedDate   string   `json:"wayBillModifiedDate"`
	WayBillModifiedBy     string   `json:"wayBillModifiedBy"`
}

type EWWayBill struct {
	EwWayBillNumber       string
	WayBillsNumber        []string
	ShipmentsNumber       []string
	CountryFrom           string
	CountryTo             string
	Consigner             string
	Consignee             string
	Custodian             string
	CustodianHistory      []string
	CustodianTime         string
	PersonConsigningGoods string
	Comments              string
	PalletsSerialNumber   []string
	AddressOfConsigner    string
	AddressOfConsignee    string
	ConsignerRegNumber    string
	VesselType            string
	VesselNumber          string
	ContainerNumber       string
	ServiceType           string
	SupportiveDocuments   []string
	EwWayBillCreationDate string
	EwWayBillCreatedBy    string
	EwWayBillModifiedDate string
	EwWayBillModifiedBy   string
}

type EntityWayBillMapping struct {
	WayBillsNumber []string
}
type CreateEntityWayBillMappingRequest struct {
	EntityName     string
	WayBillsNumber []string
}
type WayBillShipmentMapping struct {
	DCWayBillsNumber string
	DCShipmentNumber string
}
type EntityDetails struct {
	EntityName      string
	EntityType      string
	EntityAddress   string
	EntityRegNumber string
	EntityCountry   string
	Latitude        string
	Longitude       string
}

/************** Create Shipment Starts ***********************/
func CreateShipment(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Create Shipment", args[0])
	shipmentRequest := parseShipmentWayBillRequest(args[0])
	UpdatePalletCartonAssetByWayBill(stub, shipmentRequest, SHIPMENT, "")
	return saveShipmentWayBill(stub, shipmentRequest)
}

/************** Create Shipment Ends ************************/

/************** Create Way Bill Starts ***********************/
func CreateWayBill(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Create WayBill", args[0])

	wayBillRequest := parseShipmentWayBillRequest(args[0])

	shipmentDetails, err := fetchShipmentWayBillData(stub, wayBillRequest.ShipmentNumber)
	if err != nil {
		fmt.Println("Error while retrieveing the Pallet Details", err)
		return nil, err
	}
	shipmentDetails.WayBillNumber = wayBillRequest.WayBillNumber
	shipmentDetails.VehicleNumber = wayBillRequest.VehicleNumber
	shipmentDetails.VehicleType = wayBillRequest.VehicleType
	shipmentDetails.PickupDate = wayBillRequest.PickupDate
	shipmentDetails.Custodian = wayBillRequest.Custodian
	shipmentDetails.TpComments = wayBillRequest.TpComments
	shipmentDetails.WayBillCreationDate = wayBillRequest.WayBillCreationDate
	shipmentDetails.WayBillCreatedBy = wayBillRequest.WayBillCreatedBy

	UpdatePalletCartonAssetByWayBill(stub, wayBillRequest, WAYBILL, "")
	return saveShipmentWayBill(stub, shipmentDetails)
}

/************** Create Way Bill Ends ************************/

/************** Create Shipment Starts ***********************/
func CreateDCShipment(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering DC Create Shipment", args[0])
	shipmentRequest := parseShipmentWayBillRequest(args[0])
	UpdatePalletCartonAssetByWayBill(stub, shipmentRequest, DCSHIPMENT, "")
	return saveShipmentWayBill(stub, shipmentRequest)
}

/************** Create Shipment Ends ************************/

/************** Create Way Bill Starts ***********************/
func CreateDCWayBill(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Create WayBill", args[0])

	dcwayBillRequest := parseShipmentWayBillRequest(args[0])

	dcshipmentDetails, err := fetchShipmentWayBillData(stub, dcwayBillRequest.ShipmentNumber)

	if err != nil {
		fmt.Println("Error while retrieveing the Pallet Details", err)
		return nil, err
	}
	dcshipmentDetails.WayBillNumber = dcwayBillRequest.WayBillNumber
	dcshipmentDetails.VehicleNumber = dcwayBillRequest.VehicleNumber
	dcshipmentDetails.VehicleType = dcwayBillRequest.VehicleType
	dcshipmentDetails.PickupDate = dcwayBillRequest.PickupDate
	dcshipmentDetails.Custodian = dcwayBillRequest.Custodian
	dcshipmentDetails.TpComments = dcwayBillRequest.TpComments
	dcshipmentDetails.WayBillCreationDate = dcwayBillRequest.WayBillCreationDate
	dcshipmentDetails.WayBillCreatedBy = dcwayBillRequest.WayBillCreatedBy
	dcshipmentDetails.VehicleType = dcwayBillRequest.VehicleType
	dcshipmentDetails.EntityName = dcwayBillRequest.EntityName

	UpdatePalletCartonAssetByWayBill(stub, dcwayBillRequest, DCWAYBILL, "")
	UpdateEntityWayBillMapping(stub, dcshipmentDetails.EntityName, dcshipmentDetails.WayBillNumber)
	err = stub.PutState(dcshipmentDetails.WayBillNumber, []byte(dcwayBillRequest.ShipmentNumber))
	if err != nil {
		fmt.Println("Could not save WayBill to ledger", err)
		return nil, err
	}
	return saveShipmentWayBill(stub, dcshipmentDetails)
}

/************** Create Way Bill Ends ************************/

/************** Save Shipment WayBill Starts ************************/
func parseShipmentWayBillRequest(jsondata string) ShipmentWayBill {
	res := ShipmentWayBill{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}
func saveShipmentWayBill(stub shim.ChaincodeStubInterface, createShipmentWayBillRequest ShipmentWayBill) ([]byte, error) {

	shipmentWayBill := ShipmentWayBill{}
	shipmentWayBill.WayBillNumber = createShipmentWayBillRequest.WayBillNumber
	shipmentWayBill.ShipmentNumber = createShipmentWayBillRequest.ShipmentNumber
	shipmentWayBill.CountryFrom = createShipmentWayBillRequest.CountryFrom
	shipmentWayBill.CountryTo = createShipmentWayBillRequest.CountryTo
	shipmentWayBill.Consigner = createShipmentWayBillRequest.Consigner
	shipmentWayBill.Consignee = createShipmentWayBillRequest.Consignee
	shipmentWayBill.Custodian = createShipmentWayBillRequest.Custodian
	shipmentWayBill.CustodianHistory = createShipmentWayBillRequest.CustodianHistory
	shipmentWayBill.PersonConsigningGoods = createShipmentWayBillRequest.PersonConsigningGoods
	shipmentWayBill.Comments = createShipmentWayBillRequest.Comments
	shipmentWayBill.TpComments = createShipmentWayBillRequest.TpComments
	shipmentWayBill.VehicleNumber = createShipmentWayBillRequest.VehicleNumber
	shipmentWayBill.VehicleType = createShipmentWayBillRequest.VehicleType
	shipmentWayBill.PickupDate = createShipmentWayBillRequest.PickupDate
	shipmentWayBill.PalletsSerialNumber = createShipmentWayBillRequest.PalletsSerialNumber
	shipmentWayBill.AddressOfConsigner = createShipmentWayBillRequest.AddressOfConsigner
	shipmentWayBill.AddressOfConsignee = createShipmentWayBillRequest.AddressOfConsignee
	shipmentWayBill.ConsignerRegNumber = createShipmentWayBillRequest.ConsignerRegNumber
	shipmentWayBill.Carrier = createShipmentWayBillRequest.Carrier
	shipmentWayBill.VesselType = createShipmentWayBillRequest.VesselType
	shipmentWayBill.VesselNumber = createShipmentWayBillRequest.VesselNumber
	shipmentWayBill.ContainerNumber = createShipmentWayBillRequest.ContainerNumber
	shipmentWayBill.ServiceType = createShipmentWayBillRequest.ServiceType
	shipmentWayBill.ShipmentModel = createShipmentWayBillRequest.ShipmentModel
	shipmentWayBill.PalletsQuantity = createShipmentWayBillRequest.PalletsQuantity
	shipmentWayBill.CartonsQuantity = createShipmentWayBillRequest.CartonsQuantity
	shipmentWayBill.AssetsQuantity = createShipmentWayBillRequest.AssetsQuantity
	shipmentWayBill.ShipmentValue = createShipmentWayBillRequest.ShipmentValue
	shipmentWayBill.EntityName = createShipmentWayBillRequest.EntityName
	shipmentWayBill.ShipmentCreationDate = createShipmentWayBillRequest.ShipmentCreationDate
	shipmentWayBill.EWWayBillNumber = createShipmentWayBillRequest.EWWayBillNumber
	shipmentWayBill.SupportiveDocuments = createShipmentWayBillRequest.SupportiveDocuments
	shipmentWayBill.ShipmentCreatedBy = createShipmentWayBillRequest.ShipmentCreatedBy
	shipmentWayBill.ShipmentModifiedDate = createShipmentWayBillRequest.ShipmentModifiedDate
	shipmentWayBill.ShipmentModifiedBy = createShipmentWayBillRequest.ShipmentModifiedBy
	shipmentWayBill.WayBillCreationDate = createShipmentWayBillRequest.WayBillCreationDate
	shipmentWayBill.WayBillCreatedBy = createShipmentWayBillRequest.WayBillCreatedBy
	shipmentWayBill.WayBillModifiedDate = createShipmentWayBillRequest.WayBillModifiedDate
	shipmentWayBill.WayBillModifiedBy = createShipmentWayBillRequest.WayBillModifiedBy
	dataToStore, _ := json.Marshal(shipmentWayBill)

	err := stub.PutState(shipmentWayBill.ShipmentNumber, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save WayBill to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = shipmentWayBill.ShipmentNumber
	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Way Bill")
	return []byte(respString), nil

}

/************** Save Shipment WayBill Ends ************************/

/************** Get Shipment WayBill Starts ************************/

func ViewShipmentWayBill(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering ViewWayBill " + args[0])

	shipmentNo := args[0]

	wayBilldata, dataerr := fetchShipmentWayBillData(stub, shipmentNo)
	if dataerr == nil {

		dataToStore, _ := json.Marshal(wayBilldata)
		return []byte(dataToStore), nil

	}

	return nil, dataerr

}
func fetchShipmentWayBillData(stub shim.ChaincodeStubInterface, shipmentNo string) (ShipmentWayBill, error) {
	var shipmentWayBill ShipmentWayBill

	indexByte, err := stub.GetState(shipmentNo)
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

/************** End Shipment WayBill Ends ************************/

/************** Create Export Warehouse WayBill Starts ************************/
func CreateEWWayBill(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Export Warehouse WayBill ")

	ewWayBillRequest := parseEWWayBillRequest(args[0])
	lenOfArray := len(ewWayBillRequest.WayBillsNumber)
	for i := 0; i <= lenOfArray; i++ {
		wayBillShipmentMapping, err := fetchWayBillShipmentMappingData(stub, ewWayBillRequest.WayBillsNumber[i])
		dcShipmentNumber := wayBillShipmentMapping.DCShipmentNumber
		dcShipmentData, _ := fetchShipmentWayBillData(stub, dcShipmentNumber)
		UpdatePalletCartonAssetByWayBill(stub, dcShipmentData, EWWAYBILL, ewWayBillRequest.EwWayBillNumber)
		ewWayBillRequest.ShipmentsNumber = append(ewWayBillRequest.ShipmentsNumber, dcShipmentNumber)
		lenOfArray = len(dcShipmentData.PalletsSerialNumber)
		for j := 0; j <= lenOfArray; j++ {
			ewWayBillRequest.PalletsSerialNumber = append(ewWayBillRequest.ShipmentsNumber, dcShipmentData.PalletsSerialNumber[j])
		}
		if err != nil {
			fmt.Println("Could not retrive Export Warehouse WayBill ", err)
			return nil, err
		}
		wayBills, _ := fetchEntityWayBillMappingData(stub, ewWayBillRequest.Consigner)
		var tmpWayBillArray []string

		for k := 0; k < len(wayBills.WayBillsNumber); k++ {
			for j := 0; j < len(ewWayBillRequest.WayBillsNumber); j++ {
				if ewWayBillRequest.WayBillsNumber[j] != wayBills.WayBillsNumber[k] {
					tmpWayBillArray = append(tmpWayBillArray, wayBills.WayBillsNumber[k])
				}
			}
		}
		ewWayBillRequest.WayBillsNumber = tmpWayBillArray
	}

	return saveEWWayBill(stub, ewWayBillRequest)

}
func parseEWWayBillRequest(jsondata string) EWWayBill {
	res := EWWayBill{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}
func saveEWWayBill(stub shim.ChaincodeStubInterface, createEWWayBillRequest EWWayBill) ([]byte, error) {

	ewWayBill := EWWayBill{}
	ewWayBill.EwWayBillNumber = createEWWayBillRequest.EwWayBillNumber
	ewWayBill.WayBillsNumber = createEWWayBillRequest.WayBillsNumber
	ewWayBill.ShipmentsNumber = createEWWayBillRequest.ShipmentsNumber
	ewWayBill.CountryFrom = createEWWayBillRequest.CountryFrom
	ewWayBill.CountryTo = createEWWayBillRequest.CountryTo
	ewWayBill.Consigner = createEWWayBillRequest.Consigner
	ewWayBill.Consignee = createEWWayBillRequest.Consignee
	ewWayBill.Custodian = createEWWayBillRequest.Custodian
	ewWayBill.CustodianHistory = createEWWayBillRequest.CustodianHistory
	ewWayBill.CustodianTime = createEWWayBillRequest.CustodianTime
	ewWayBill.PersonConsigningGoods = createEWWayBillRequest.PersonConsigningGoods
	ewWayBill.Comments = createEWWayBillRequest.Comments
	ewWayBill.PalletsSerialNumber = createEWWayBillRequest.PalletsSerialNumber
	ewWayBill.AddressOfConsigner = createEWWayBillRequest.AddressOfConsigner
	ewWayBill.AddressOfConsignee = createEWWayBillRequest.AddressOfConsignee
	ewWayBill.ConsignerRegNumber = createEWWayBillRequest.ConsignerRegNumber
	ewWayBill.VesselType = createEWWayBillRequest.VesselType
	ewWayBill.VesselNumber = createEWWayBillRequest.VesselNumber
	ewWayBill.ContainerNumber = createEWWayBillRequest.ContainerNumber
	ewWayBill.ServiceType = createEWWayBillRequest.ServiceType
	ewWayBill.SupportiveDocuments = createEWWayBillRequest.SupportiveDocuments
	ewWayBill.EwWayBillCreationDate = createEWWayBillRequest.EwWayBillCreationDate
	ewWayBill.EwWayBillCreatedBy = createEWWayBillRequest.EwWayBillCreatedBy
	ewWayBill.EwWayBillModifiedDate = createEWWayBillRequest.EwWayBillModifiedDate
	ewWayBill.EwWayBillModifiedBy = createEWWayBillRequest.EwWayBillModifiedBy

	dataToStore, _ := json.Marshal(ewWayBill)

	err := stub.PutState(ewWayBill.EwWayBillNumber, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Export Warehouse Way Bill to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = ewWayBill.EwWayBillNumber

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Export Warehouse Way Bill")
	return []byte(respString), nil

}

/************** Create Export Warehouse WayBill Ends ************************/

/************** View Export Warehouse WayBill Starts ************************/
func ViewEWWayBill(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering ViewEWWayBill " + args[0])

	ewWayBillNumber := args[0]

	emWayBilldata, dataerr := fetchEWWayBillData(stub, ewWayBillNumber)
	if dataerr == nil {

		dataToStore, _ := json.Marshal(emWayBilldata)
		return []byte(dataToStore), nil

	}

	return nil, dataerr

}
func fetchEWWayBillData(stub shim.ChaincodeStubInterface, ewWayBillNumber string) (EWWayBill, error) {
	var ewWayBill EWWayBill

	indexByte, err := stub.GetState(ewWayBillNumber)
	if err != nil {
		fmt.Println("Could not retrive Export Warehouse WayBill ", err)
		return ewWayBill, err
	}

	if marshErr := json.Unmarshal(indexByte, &ewWayBill); marshErr != nil {
		fmt.Println("Could not retrieve Export Warehouse from ledger", marshErr)
		return ewWayBill, marshErr
	}

	return ewWayBill, nil

}

/************** View Export Warehouse WayBill Ends ************************/

/************** Create Entity WayBill Mapping Starts ************************/
func CreateEntityWayBillMapping(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Create Entity WayBill Mapping")
	entityWayBillMappingRequest := parseEntityWayBillMapping(args[0])

	return saveEntityWayBillMapping(stub, entityWayBillMappingRequest)

}
func parseEntityWayBillMapping(jsondata string) CreateEntityWayBillMappingRequest {
	res := CreateEntityWayBillMappingRequest{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}
func saveEntityWayBillMapping(stub shim.ChaincodeStubInterface, createEntityWayBillMappingRequest CreateEntityWayBillMappingRequest) ([]byte, error) {

	entityWayBillMapping := EntityWayBillMapping{}
	entityWayBillMapping.WayBillsNumber = createEntityWayBillMappingRequest.WayBillsNumber

	dataToStore, _ := json.Marshal(entityWayBillMapping)

	err := stub.PutState(createEntityWayBillMappingRequest.EntityName, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Entity WayBill Mapping to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = createEntityWayBillMappingRequest.EntityName

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Entity WayBill Mapping")
	return []byte(respString), nil

}

/************** Create Entity WayBill Mapping Ends ************************/

/************** Update Entity WayBill Mapping Starts ************************/
func UpdateEntityWayBillMapping(stub shim.ChaincodeStubInterface, entityName string, wayBillsNumber string) ([]byte, error) {
	fmt.Println("Entering Update Entity WayBill Mapping")
	entityWayBillMappingRequest := CreateEntityWayBillMappingRequest{}
	entityWayBillMapping, err := fetchEntityWayBillMappingData(stub, entityName)

	if err != nil {
		entityWayBillMappingRequest.EntityName = entityName
		entityWayBillMappingRequest.WayBillsNumber = append(entityWayBillMappingRequest.WayBillsNumber, wayBillsNumber)
		saveEntityWayBillMapping(stub, entityWayBillMappingRequest)
	} else {
		entityWayBillMappingRequest.WayBillsNumber = append(entityWayBillMapping.WayBillsNumber, wayBillsNumber)
		fmt.Println("Updated Entity", entityWayBillMappingRequest)
		dataToStore, _ := json.Marshal(entityWayBillMappingRequest)
		err := stub.PutState(entityName, []byte(dataToStore))
		if err != nil {
			fmt.Println("Could not save Entity WayBill Mapping to ledger", err)
			return nil, err
		}
	}
	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = entityName

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Entity WayBill Mapping")
	return []byte(respString), nil

}

/************** Create Assets Starts ************************/
/************** Get Entity WayBill Mapping Starts ************************/
func GetEntityWayBillMapping(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Get Entity WayBill Mapping")
	entityName := args[0]
	wayBillEntityMappingData, dataerr := fetchEntityWayBillMappingData(stub, entityName)
	if dataerr == nil {

		dataToStore, _ := json.Marshal(wayBillEntityMappingData)
		return []byte(dataToStore), nil

	}

	return nil, dataerr
}

func fetchEntityWayBillMappingData(stub shim.ChaincodeStubInterface, entityName string) (EntityWayBillMapping, error) {
	var entityWayBillMapping EntityWayBillMapping

	indexByte, err := stub.GetState(entityName)
	if err != nil {
		fmt.Println("Could not retrive Entity WayBill Mapping ", err)
		return entityWayBillMapping, err
	}

	if marshErr := json.Unmarshal(indexByte, &entityWayBillMapping); marshErr != nil {
		fmt.Println("Could not retrieve Entity WayBill Mapping from ledger", marshErr)
		return entityWayBillMapping, marshErr
	}

	return entityWayBillMapping, nil

}

/************** Get Entity Mapping Ends ************************/

/************** Create WayBill Shipment Mapping Starts ************************/

func CreateWayBillShipmentMapping(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Create WayBill Shipment Mapping")
	wayBillShipmentMapping := parseWayBillShipmentMapping(args[0])

	return saveWayBillShipmentMapping(stub, wayBillShipmentMapping)

}
func parseWayBillShipmentMapping(jsondata string) WayBillShipmentMapping {
	res := WayBillShipmentMapping{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}
func saveWayBillShipmentMapping(stub shim.ChaincodeStubInterface, craeteWayBillShipmentMappingRequest WayBillShipmentMapping) ([]byte, error) {

	wayBillShipmentMapping := WayBillShipmentMapping{}
	wayBillShipmentMapping.DCWayBillsNumber = craeteWayBillShipmentMappingRequest.DCWayBillsNumber
	wayBillShipmentMapping.DCShipmentNumber = craeteWayBillShipmentMappingRequest.DCShipmentNumber
	dataToStore, _ := json.Marshal(wayBillShipmentMapping)

	err := stub.PutState(wayBillShipmentMapping.DCWayBillsNumber, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save WayBill Shipment Mapping to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = wayBillShipmentMapping.DCWayBillsNumber

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Entity WayBill Mapping")
	return []byte(respString), nil

}

/************** Create WayBill Shipment Ends ************************/

/************** Get  WayBill Shipment Mapping Starts ************************/
func GetWayBillShipmentMapping(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Get Entity WayBill Mapping")
	wayBillNumber := args[0]
	wayBillShippingMappingData, dataerr := fetchEntityWayBillMappingData(stub, wayBillNumber)
	if dataerr == nil {

		dataToStore, _ := json.Marshal(wayBillShippingMappingData)
		return []byte(dataToStore), nil

	}

	return nil, dataerr
}

func fetchWayBillShipmentMappingData(stub shim.ChaincodeStubInterface, wayBillNumber string) (WayBillShipmentMapping, error) {
	var wayBillShipmentMapping WayBillShipmentMapping

	indexByte, err := stub.GetState(wayBillNumber)
	if err != nil {
		fmt.Println("Could not retrive WayBill Shipping Mapping ", err)
		return wayBillShipmentMapping, err
	}

	if marshErr := json.Unmarshal(indexByte, &wayBillShipmentMapping); marshErr != nil {
		fmt.Println("Could not retrieve Entity WayBill Mapping from ledger", marshErr)
		return wayBillShipmentMapping, marshErr
	}

	return wayBillShipmentMapping, nil

}

/************** Get  WayBill Shipment Mapping Ends ************************/

/************** Create Assets Starts ************************/

func CreateAsset(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Create Assets ")

	assetDetailsRequest := parseAssetRequest(args[0])

	return saveAssetDetails(stub, assetDetailsRequest)

}
func parseAssetRequest(jsondata string) AssetDetails {
	res := AssetDetails{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}
func saveAssetDetails(stub shim.ChaincodeStubInterface, createAssetDetailsRequest AssetDetails) ([]byte, error) {
	assetDetails := AssetDetails{}
	assetDetails.AssetSerialNo = createAssetDetailsRequest.AssetSerialNo
	assetDetails.AssetModel = createAssetDetailsRequest.AssetModel
	assetDetails.AssetType = createAssetDetailsRequest.AssetType
	assetDetails.AssetMake = createAssetDetailsRequest.AssetMake
	assetDetails.AssetCOO = createAssetDetailsRequest.AssetCOO
	assetDetails.AssetMaufacture = createAssetDetailsRequest.AssetMaufacture
	assetDetails.AssetStatus = createAssetDetailsRequest.AssetStatus
	assetDetails.CreatedBy = createAssetDetailsRequest.CreatedBy
	assetDetails.CreatedDate = createAssetDetailsRequest.CreatedDate
	assetDetails.ModifiedBy = createAssetDetailsRequest.ModifiedBy
	assetDetails.ModifiedDate = createAssetDetailsRequest.ModifiedDate
	assetDetails.PalletSerialNumber = createAssetDetailsRequest.PalletSerialNumber
	assetDetails.CartonSerialNumber = createAssetDetailsRequest.CartonSerialNumber
	assetDetails.MshipmentNumber = createAssetDetailsRequest.MshipmentNumber
	assetDetails.DcShipmentNumber = createAssetDetailsRequest.DcShipmentNumber
	assetDetails.MwayBillNumber = createAssetDetailsRequest.MwayBillNumber
	assetDetails.DcWayBillNumber = createAssetDetailsRequest.DcWayBillNumber
	assetDetails.EwWayBillNumber = createAssetDetailsRequest.EwWayBillNumber
	assetDetails.MShipmentDate = createAssetDetailsRequest.MShipmentDate
	assetDetails.DcShipmentDate = createAssetDetailsRequest.DcShipmentDate
	assetDetails.MWayBillDate = createAssetDetailsRequest.MWayBillDate
	assetDetails.DcWayBillDate = createAssetDetailsRequest.DcWayBillDate
	assetDetails.EwWayBillDate = createAssetDetailsRequest.EwWayBillDate

	dataToStore, _ := json.Marshal(assetDetails)

	err := stub.PutState(assetDetails.AssetSerialNo, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Assets Details to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = assetDetails.AssetSerialNo

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Asset Details")
	return []byte(respString), nil

}

/************** Create Assets Ends ************************/

/************** Create Carton Starts ************************/
func CreateCarton(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Create Cortons ")

	cartonDetailslRequest := parseCartonRequest(args[0])

	return saveCartonDetails(stub, cartonDetailslRequest)

}
func parseCartonRequest(jsondata string) CartonDetails {
	res := CartonDetails{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}
func saveCartonDetails(stub shim.ChaincodeStubInterface, createCartonDetailsRequest CartonDetails) ([]byte, error) {
	cartonDetails := CartonDetails{}
	cartonDetails.CartonSerialNo = createCartonDetailsRequest.CartonSerialNo
	cartonDetails.CartonModel = createCartonDetailsRequest.CartonModel
	cartonDetails.CartonStatus = createCartonDetailsRequest.CartonStatus
	cartonDetails.CartonCreationDate = createCartonDetailsRequest.CartonCreationDate
	cartonDetails.PalletSerialNumber = createCartonDetailsRequest.PalletSerialNumber
	cartonDetails.AssetsSerialNumber = createCartonDetailsRequest.AssetsSerialNumber
	cartonDetails.MshipmentNumber = createCartonDetailsRequest.MshipmentNumber
	cartonDetails.DcShipmentNumber = createCartonDetailsRequest.DcShipmentNumber
	cartonDetails.MwayBillNumber = createCartonDetailsRequest.MwayBillNumber
	cartonDetails.DcWayBillNumber = createCartonDetailsRequest.DcWayBillNumber
	cartonDetails.EwWayBillNumber = createCartonDetailsRequest.EwWayBillNumber
	cartonDetails.Dimensions = createCartonDetailsRequest.Dimensions
	cartonDetails.Weight = createCartonDetailsRequest.Weight
	cartonDetails.MShipmentDate = createCartonDetailsRequest.MShipmentDate
	cartonDetails.DcShipmentDate = createCartonDetailsRequest.DcShipmentDate
	cartonDetails.MWayBillDate = createCartonDetailsRequest.MWayBillDate
	cartonDetails.DcWayBillDate = createCartonDetailsRequest.DcWayBillDate
	cartonDetails.EwWayBillDate = createCartonDetailsRequest.EwWayBillDate
	dataToStore, _ := json.Marshal(cartonDetails)

	err := stub.PutState(cartonDetails.CartonSerialNo, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Carton Details to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = cartonDetails.CartonSerialNo

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Carton Details")
	return []byte(respString), nil

}

/************** Create Carton Ends ************************/
/************** Create Pallets Starts ************************/
func CreatePallet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Create Pallets ")

	palletDetailslRequest := parsePalletRequest(args[0])

	return savePalletDetails(stub, palletDetailslRequest)

}
func parsePalletRequest(jsondata string) PalletDetails {
	res := PalletDetails{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}
func savePalletDetails(stub shim.ChaincodeStubInterface, createPalletDetailsRequest PalletDetails) ([]byte, error) {
	palletDetails := PalletDetails{}
	palletDetails.PalletSerialNo = createPalletDetailsRequest.PalletSerialNo
	palletDetails.PalletModel = createPalletDetailsRequest.PalletModel
	palletDetails.PalletStatus = createPalletDetailsRequest.PalletStatus
	palletDetails.CartonSerialNumber = createPalletDetailsRequest.CartonSerialNumber
	palletDetails.PalletCreationDate = createPalletDetailsRequest.PalletCreationDate
	palletDetails.AssetsSerialNumber = createPalletDetailsRequest.AssetsSerialNumber
	palletDetails.MshipmentNumber = createPalletDetailsRequest.MshipmentNumber
	palletDetails.DcShipmentNumber = createPalletDetailsRequest.DcShipmentNumber
	palletDetails.MwayBillNumber = createPalletDetailsRequest.MwayBillNumber
	palletDetails.DcWayBillNumber = createPalletDetailsRequest.DcWayBillNumber
	palletDetails.EwWayBillNumber = createPalletDetailsRequest.EwWayBillNumber
	palletDetails.Dimensions = createPalletDetailsRequest.Dimensions
	palletDetails.Weight = createPalletDetailsRequest.Weight
	palletDetails.MShipmentDate = createPalletDetailsRequest.MShipmentDate
	palletDetails.DcShipmentDate = createPalletDetailsRequest.DcShipmentDate
	palletDetails.MWayBillDate = createPalletDetailsRequest.MWayBillDate
	palletDetails.DcWayBillDate = createPalletDetailsRequest.DcWayBillDate
	palletDetails.EwWayBillDate = createPalletDetailsRequest.EwWayBillDate
	dataToStore, _ := json.Marshal(palletDetails)

	err := stub.PutState(palletDetails.PalletSerialNo, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Pallet Details to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = palletDetails.PalletSerialNo

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Pallet Details")
	return []byte(respString), nil

}

/************** Create Pallets Ends ************************/
/************** View Asset Starts ************************/
func GetAsset(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering GetAsset " + args[0])

	assetSerialNo := args[0]

	assetData, dataerr := fetchAssetDetails(stub, assetSerialNo)
	if dataerr == nil {

		dataToStore, _ := json.Marshal(assetData)
		return []byte(dataToStore), nil

	}

	return nil, dataerr

}
func fetchAssetDetails(stub shim.ChaincodeStubInterface, assetSerialNo string) (AssetDetails, error) {
	var assetDetails AssetDetails

	indexByte, err := stub.GetState(assetSerialNo)
	if err != nil {
		fmt.Println("Could not retrive Asset Details ", err)
		return assetDetails, err
	}

	if marshErr := json.Unmarshal(indexByte, &assetDetails); marshErr != nil {
		fmt.Println("Could not retrieve Asset Details from ledger", marshErr)
		return assetDetails, marshErr
	}

	return assetDetails, nil

}

/************** View Asset Ends ************************/

/************** View Carton Starts ************************/
func GetCarton(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering GetPallet " + args[0])

	cartonSerialNo := args[0]

	cartonData, dataerr := fetchCartonDetails(stub, cartonSerialNo)
	if dataerr == nil {

		dataToStore, _ := json.Marshal(cartonData)
		return []byte(dataToStore), nil

	}

	return nil, dataerr

}
func fetchCartonDetails(stub shim.ChaincodeStubInterface, cartonSerialNo string) (CartonDetails, error) {
	var cartonDetails CartonDetails

	indexByte, err := stub.GetState(cartonSerialNo)
	if err != nil {
		fmt.Println("Could not retrive Carton Details ", err)
		return cartonDetails, err
	}

	if marshErr := json.Unmarshal(indexByte, &cartonDetails); marshErr != nil {
		fmt.Println("Could not retrieve Carton Details from ledger", marshErr)
		return cartonDetails, marshErr
	}

	return cartonDetails, nil

}

/************** View Carton Ends ************************/
/************** View Pallet Starts ************************/
func GetPallet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering GetPallet " + args[0])

	palletSerialNo := args[0]

	palletData, dataerr := fetchPalletDetails(stub, palletSerialNo)
	if dataerr == nil {

		dataToStore, _ := json.Marshal(palletData)
		return []byte(dataToStore), nil

	}

	return nil, dataerr

}
func fetchPalletDetails(stub shim.ChaincodeStubInterface, palletSerialNo string) (PalletDetails, error) {
	var palletDetails PalletDetails

	indexByte, err := stub.GetState(palletSerialNo)
	if err != nil {
		fmt.Println("Could not retrive Pallet Details ", err)
		return palletDetails, err
	}

	if marshErr := json.Unmarshal(indexByte, &palletDetails); marshErr != nil {
		fmt.Println("Could not retrieve Pallet Details from ledger", marshErr)
		return palletDetails, marshErr
	}

	return palletDetails, nil

}

/************** View Pallet Ends ************************/

/************** Update Asset Details Starts ************************/
func UpdateAssetDetails(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Update Asset Details")
	assetSerialNo := args[0]
	wayBillNumber := args[1]
	assetDetails, _ := fetchAssetDetails(stub, assetSerialNo)

	assetDetails.EwWayBillNumber = wayBillNumber

	fmt.Println("Updated Entity", assetDetails)
	dataToStore, _ := json.Marshal(assetDetails)
	err := stub.PutState(assetSerialNo, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Entity WayBill Mapping to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = assetSerialNo

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Entity WayBill Mapping")
	return []byte(respString), nil

}

/************** Update Asset Details Ends ************************/

/************** Update Carton Details Starts ************************/
func UpdateCartonDetails(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Update Carton Details")
	cartonSerialNo := args[0]
	wayBillNumber := args[1]
	cartonDetails, _ := fetchCartonDetails(stub, cartonSerialNo)
	cartonDetails.MwayBillNumber = wayBillNumber
	fmt.Println("Updated Entity", cartonDetails)
	dataToStore, _ := json.Marshal(cartonDetails)
	err := stub.PutState(cartonSerialNo, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Pallet Details to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = cartonSerialNo

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Carton Details")
	return []byte(respString), nil
}

/************** Update Carton Details Ends ************************/

/************** Update Pallet Details Starts ************************/
func UpdatePalletDetails(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Update Pallet Details")
	palletSerialNo := args[0]
	wayBillNumber := args[1]
	palletDetails, _ := fetchPalletDetails(stub, palletSerialNo)
	palletDetails.MwayBillNumber = wayBillNumber
	fmt.Println("Updated Entity", palletDetails)
	dataToStore, _ := json.Marshal(palletDetails)
	err := stub.PutState(palletSerialNo, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Pallet Details to ledger", err)
		return nil, err
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = palletSerialNo

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Pallet Details")
	return []byte(respString), nil

}

/************** Update Pallet Details Ends ************************/

/************** Update Pallet Details Starts ************************/
func UpdatePalletCartonAssetByWayBill(stub shim.ChaincodeStubInterface, wayBillRequest ShipmentWayBill, source string, ewWaybillId string) ([]byte, error) {
	fmt.Println("Entering Update Pallet Carton Asset Details")
	// Start Loop for Pallet Nos
	lenOfArray := len(wayBillRequest.PalletsSerialNumber)
	for i := 0; i < lenOfArray; i++ {

		palletData, err := fetchPalletDetails(stub, wayBillRequest.PalletsSerialNumber[i])

		if err != nil {
			fmt.Println("Error while retrieveing the Pallet Details", err)
			return nil, err
		}

		if source == SHIPMENT {
			palletData.MshipmentNumber = wayBillRequest.ShipmentNumber
		} else if source == WAYBILL {
			palletData.MwayBillNumber = wayBillRequest.WayBillNumber
		} else if source == DCSHIPMENT {
			palletData.DcShipmentNumber = wayBillRequest.ShipmentNumber
		} else if source == DCWAYBILL {
			palletData.DcWayBillNumber = wayBillRequest.WayBillNumber
		}
		savePalletDetails(stub, palletData)

		//Start Loop for Carton Nos
		lenOfArray = len(palletData.CartonSerialNumber)
		for i := 0; i < lenOfArray; i++ {

			cartonData, err := fetchCartonDetails(stub, palletData.CartonSerialNumber[i])

			if err != nil {
				fmt.Println("Error while retrieveing the Carton Details", err)
				return nil, err
			}
			if source == SHIPMENT {
				cartonData.MshipmentNumber = wayBillRequest.ShipmentNumber
			} else if source == WAYBILL {
				cartonData.MwayBillNumber = wayBillRequest.WayBillNumber
			} else if source == DCSHIPMENT {
				cartonData.DcShipmentNumber = wayBillRequest.ShipmentNumber
			} else if source == DCWAYBILL {
				cartonData.DcWayBillNumber = wayBillRequest.WayBillNumber
			}
			saveCartonDetails(stub, cartonData)
		} //End Loop for Carton Nos

		//Start Loop for Asset Nos
		lenOfArray = len(palletData.AssetsSerialNumber)
		for i := 0; i < lenOfArray; i++ {

			assetData, err := fetchAssetDetails(stub, palletData.AssetsSerialNumber[i])

			if err != nil {
				fmt.Println("Error while retrieveing the Asset Details", err)
				return nil, err
			}
			if source == SHIPMENT {
				assetData.MshipmentNumber = wayBillRequest.ShipmentNumber
			} else if source == WAYBILL {
				assetData.MwayBillNumber = wayBillRequest.WayBillNumber
			} else if source == DCSHIPMENT {
				assetData.DcShipmentNumber = wayBillRequest.ShipmentNumber
			} else if source == DCWAYBILL {
				assetData.DcWayBillNumber = wayBillRequest.WayBillNumber
			}
			saveAssetDetails(stub, assetData)
		} //End Loop for Asset Nos
	}

	resp := BlockchainResponse{}
	resp.Err = "000"
	resp.Message = ""

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved Pallet Carton Asset Details")
	return []byte(respString), nil
}

/************** Update Pallet Details Ends ************************/

/**************Arshad End new code as per LLD***************


/************** Create Shipment Starts ************************/
/**
	Expected Input is
	{
		"shipmentNumber"" : "123456",
		"personConsigningGoods" : "KarthikS",
		"consigner" : "HCL",
		"consignerAddress" : "Chennai",
		"consignee" : "HCL-AM",
		"consigneeAddress" : "Dallas",
		"consigneeRegNo" : "12122222222",
		"ModelNo" : "IA1a1222",
		"quantity" : "50",
		"pallets" : ["11111111","22222222","333333"],
		"status" : "intra",
		"notes" : "ha haha ha ha",
		"CreatedBy" : "KarthikSukumaram",
		"custodian" : "HCL",
		"createdTimeStamp" : "2017-03-02"
	}
**/

type CreateShipmentRequest struct {
	ShipmentNumber        string   `json:"shipmentNumber"`
	PersonConsigningGoods string   `json:"personConsigningGoods"`
	Consigner             string   `json:"consigner"`
	ConsignerAddress      string   `json:"consignerAddress"`
	Consignee             string   `json:"consignee"`
	ConsigneeAddress      string   `json:"consigneeAddress"`
	ConsigneeRegNo        string   `json:"consigneeRegNo"`
	ModelNo               string   `json:"modelNo"`
	Quantity              string   `json:"quantity"`
	Pallets               []string `json:"pallets"`
	Carrier               string   `json:"status"`
	Notes                 string   `json:"notes"`
	CreatedBy             string   `json:"createdBy"`
	Custodian             string   `json:"custodian"`
	CreatedTimeStamp      string   `json:"createdTimeStamp"`
	CallingEntityName     string   `json:"callingEntityName"`
}

type CreateShipmentResponse struct {
	Err     string `json:"err"`
	ErrMsg  string `json:"errMsg"`
	Message string `json:"message"`
}

/*
func CreateShipment(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering CreateShipment")

	shipmentRequest := parseCreateShipmentRequest(args[0])

	return processShipment(stub, shipmentRequest)

}*/

func processShipment(stub shim.ChaincodeStubInterface, shipmentRequest CreateShipmentRequest) ([]byte, error) {
	shipment := Shipment{}
	shipmentIndex := ShipmentIndex{}

	shipment.ShipmentNumber = shipmentRequest.ShipmentNumber
	shipment.PersonConsigningGoods = shipmentRequest.PersonConsigningGoods
	shipment.Consigner = shipmentRequest.Consigner
	shipment.ConsignerAddress = shipmentRequest.ConsignerAddress
	shipment.Consignee = shipmentRequest.Consignee
	shipment.ConsigneeAddress = shipmentRequest.ConsigneeAddress
	shipment.ConsigneeRegNo = shipmentRequest.ConsigneeRegNo
	shipment.ModelNo = shipmentRequest.ModelNo
	shipment.Quantity = shipmentRequest.Quantity
	shipment.Pallets = shipmentRequest.Pallets
	shipment.Carrier = shipmentRequest.Carrier
	shipment.CreatedBy = shipmentRequest.CreatedBy
	shipment.Custodian = shipmentRequest.Custodian
	shipment.CreatedTimeStamp = shipmentRequest.CreatedTimeStamp
	shipment.Status = "Created"

	var acl []string
	acl = append(acl, shipmentRequest.CallingEntityName) //TODO: Have to take the Entity name from the Certificate
	shipment.Acl = acl

	shipmentIndex.ShipmentNumber = shipmentRequest.ShipmentNumber
	shipmentIndex.Status = shipment.Status
	shipmentIndex.Acl = acl

	dataToStore, _ := json.Marshal(shipment)

	err := stub.PutState(shipment.ShipmentNumber, []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Shipment to ledger", err)
		return nil, err
	}

	addShipmentIndex(stub, shipmentIndex)

	resp := CreateShipmentResponse{}
	resp.Err = "000"
	resp.Message = shipment.ShipmentNumber

	respString, _ := json.Marshal(resp)

	fmt.Println("Successfully saved way bill")
	return []byte(respString), nil

}

func addShipmentIndex(stub shim.ChaincodeStubInterface, shipmentIndex ShipmentIndex) error {
	indexByte, err := stub.GetState("SHIPMENT_INDEX")
	if err != nil {
		fmt.Println("Could not retrive Shipment Index", err)
		return err
	}
	allShipmentIndex := AllShipment{}

	if marshErr := json.Unmarshal(indexByte, &allShipmentIndex); marshErr != nil {
		fmt.Println("Could not save Shipment to ledger", marshErr)
		return marshErr
	}

	allShipmentIndex.ShipmentIndexArr = append(allShipmentIndex.ShipmentIndexArr, shipmentIndex)
	dataToStore, _ := json.Marshal(allShipmentIndex)

	addErr := stub.PutState("SHIPMENT_INDEX", []byte(dataToStore))
	if addErr != nil {
		fmt.Println("Could not save Shipment to ledger", addErr)
		return addErr
	}

	return nil
}

func parseCreateShipmentRequest(jsondata string) CreateShipmentRequest {
	res := CreateShipmentRequest{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}

/************** Create Shipment Ends ************************/

/************** View Shipment Starts ************************/

type ViewShipmentRequest struct {
	CallingEntityName string `json:"callingEntityName"`
	ShipmentNumber    string `json:"shipmentNumber"`
}

func ViewShipment(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering ViewShipment " + args[0])

	/*request := parseViewShipmentRequest(args[0])

	shipmentData, dataerr := fetchShipmentData(stub, request.ShipmentNumber)
	if dataerr == nil {
		if hasPermission(shipmentData.Acl, request.CallingEntityName) {
			dataToStore, _ := json.Marshal(shipmentData)
			return []byte(dataToStore), nil
		} else {
			return []byte("{ \"errMsg\": \"No data found\" }"), nil
		}
	}*/

	return nil, nil

}

func parseViewShipmentRequest(jsondata string) ViewShipmentRequest {
	res := ViewShipmentRequest{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}

/************** View Shipment Ends ************************/

/************** Inbox Service Starts ************************/

/**
	Expected Input is
	{
		"callingEntityName" : "INTEL",
		"status" : "Created"
	}
**/

/*type InboxRequest struct {
	CallingEntityName string `json:"callingEntityName"`
	Status            string `json:"status"`
}

type InboxResponse struct {
	Data []Shipment `json:"data"`
}

func Inbox(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering Inbox " + args[0])

	request := parseInboxRequest(args[0])

	return fetchShipmentIndex(stub, request.CallingEntityName, request.Status)

}

func parseInboxRequest(jsondata string) InboxRequest {
	res := InboxRequest{}
	json.Unmarshal([]byte(jsondata), &res)
	fmt.Println(res)
	return res
}

func hasPermission(acl []string, currUser string) bool {
	lenOfArray := len(acl)

	for i := 0; i < lenOfArray; i++ {
		if acl[i] == currUser {
			return true
		}
	}

	return false
}

func fetchShipmentData(stub shim.ChaincodeStubInterface, shipmentNumber string) (Shipment, error) {
	var shipmentData Shipment

	indexByte, err := stub.GetState(shipmentNumber)
	if err != nil {
		fmt.Println("Could not retrive Shipment Index", err)
		return shipmentData, err
	}

	if marshErr := json.Unmarshal(indexByte, &shipmentData); marshErr != nil {
		fmt.Println("Could not save Shipment to ledger", marshErr)
		return shipmentData, marshErr
	}

	return shipmentData, nil

}

func fetchShipmentIndex(stub shim.ChaincodeStubInterface, callingEntityName string, status string) ([]byte, error) {
	allShipmentIndex := AllShipment{}
	var shipmentIndexArr []ShipmentIndex
	var tmpShipmentIndex ShipmentIndex
	var shipmentDataArr []Shipment
	resp := InboxResponse{}

	indexByte, err := stub.GetState("SHIPMENT_INDEX")
	if err != nil {
		fmt.Println("Could not retrive Shipment Index", err)
		return nil, err
	}

	if marshErr := json.Unmarshal(indexByte, &allShipmentIndex); marshErr != nil {
		fmt.Println("Could not save Shipment to ledger", marshErr)
		return nil, marshErr
	}

	shipmentIndexArr = allShipmentIndex.ShipmentIndexArr

	lenOfArray := len(shipmentIndexArr)

	for i := 0; i < lenOfArray; i++ {
		tmpShipmentIndex = shipmentIndexArr[i]
		if tmpShipmentIndex.Status == status {
			if hasPermission(tmpShipmentIndex.Acl, callingEntityName) {
				shipmentData, dataerr := fetchShipmentData(stub, tmpShipmentIndex.ShipmentNumber)
				if dataerr == nil {
					shipmentDataArr = append(shipmentDataArr, shipmentData)
				}
			}
		}
	}

	resp.Data = shipmentDataArr
	dataToStore, _ := json.Marshal(resp)

	return []byte(dataToStore), nil
}*/

/************** Inbox Service Ends ************************/

/************** Asset Search Service Starts ************************/

type SearchAssetRequest struct {
	CallingEntityName string `json:"callingEntityName"`
	AssetId           string `json:"assetId"`
}

type SearchAssetResponse struct {
	AssetId        string     `json:"assetId"`
	Modeltype      string     `json:"modeltype"`
	Color          string     `json:"color"`
	CartonId       string     `json:"cartonId"`
	PalletId       string     `json:"palletId"`
	ShipmentDetail []Shipment `json:"shipmentDetail"`
	ErrCode        string     `json:"errCode"`
	ErrMessage     string     `json:"errMessage"`
}

func parseAsset(stub shim.ChaincodeStubInterface, assetId string) (Asset, error) {
	var asset Asset

	assetBytes, err := stub.GetState(assetId)
	if err != nil {
		return asset, err
	} else {
		if marshErr := json.Unmarshal(assetBytes, &asset); marshErr != nil {
			fmt.Println("Could not Unmarshal Asset", marshErr)
			return asset, marshErr
		}
		return asset, nil
	}

}

func retrieveShipment(stub shim.ChaincodeStubInterface, shipmentId string) (Shipment, error) {
	var shipment Shipment

	shipmentBytes, err := stub.GetState(shipmentId)
	if err != nil {
		return shipment, err
	} else {
		if marshErr := json.Unmarshal(shipmentBytes, &shipment); marshErr != nil {
			fmt.Println("Could not Unmarshal Asset", marshErr)
			return shipment, marshErr
		}
		return shipment, nil
	}
}
func PrepareSearchAssetResponse(stub shim.ChaincodeStubInterface, asset Asset) ([]byte, error) {
	var resp SearchAssetResponse
	var shipmentArr []Shipment
	var tmpShipment Shipment
	var err error

	resp.AssetId = asset.AssetId
	resp.Modeltype = asset.Modeltype
	resp.Color = asset.Color
	resp.CartonId = asset.CartonId
	resp.PalletId = asset.PalletId

	lenOfArray := len(asset.ShipmentIds)

	for i := 0; i < lenOfArray; i++ {
		tmpShipment, err = retrieveShipment(stub, asset.ShipmentIds[i])
		if err != nil {
			fmt.Println("Error while retrieveing the Shipment Details", err)
			return nil, err
		} else {
			shipmentArr = append(shipmentArr, tmpShipment)
		}
	}

	resp.ShipmentDetail = shipmentArr
	return json.Marshal(resp)

}

func parseSearchAssetRequest(requestParam string) (SearchAssetRequest, error) {
	var request SearchAssetRequest

	if marshErr := json.Unmarshal([]byte(requestParam), &request); marshErr != nil {
		fmt.Println("Could not Unmarshal Asset", marshErr)
		return request, marshErr
	}
	return request, nil

}

func SearchAsset(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering SearchAsset " + args[0])
	var asset Asset
	var err error
	var request SearchAssetRequest
	var resp SearchAssetResponse

	request, err = parseSearchAssetRequest(args[0])
	if err != nil {
		resp.ErrCode = INVALID_INPUT_ERROR_CODE
		resp.ErrMessage = INVALID_INPUT_ERROR_MSG
		return json.Marshal(resp)
	}

	asset, err = parseAsset(stub, request.AssetId)

	if err != nil {
		resp.ErrCode = NODATA_ERROR_CODE
		resp.ErrMessage = NODATA_ERROR_MSG
		return json.Marshal(resp)
	}

	return PrepareSearchAssetResponse(stub, asset)

}

/************** Asset Search Service Ends ************************/

/************** Carton Search Service Starts ************************/

type SearchCartonRequest struct {
	CallingEntityName string `json:"callingEntityName"`
	CartonId          string `json:"cartonId"`
}

type SearchCartonResponse struct {
	CartonId       string     `json:"cartonId"`
	PalletId       string     `json:"palletId"`
	ShipmentDetail []Shipment `json:"shipmentDetail"`
	ErrCode        string     `json:"errCode"`
	ErrMessage     string     `json:"errMessage"`
}

func parseSearchCartonRequest(requestParam string) (SearchCartonRequest, error) {
	var request SearchCartonRequest

	if marshErr := json.Unmarshal([]byte(requestParam), &request); marshErr != nil {
		fmt.Println("Could not Unmarshal Asset", marshErr)
		return request, marshErr
	}
	return request, nil

}

func parseCarton(stub shim.ChaincodeStubInterface, cartonId string) (Carton, error) {
	var carton Carton

	cartonBytes, err := stub.GetState(cartonId)
	if err != nil {
		return carton, err
	} else {
		if marshErr := json.Unmarshal(cartonBytes, &carton); marshErr != nil {
			fmt.Println("Could not Unmarshal Asset", marshErr)
			return carton, marshErr
		}
		return carton, nil
	}

}

func PrepareSearchCartontResponse(stub shim.ChaincodeStubInterface, carton Carton) ([]byte, error) {
	var resp SearchCartonResponse
	var shipmentArr []Shipment
	var tmpShipment Shipment
	var err error

	resp.CartonId = carton.CartonId
	resp.PalletId = carton.PalletId

	lenOfArray := len(carton.ShipmentIds)

	for i := 0; i < lenOfArray; i++ {
		tmpShipment, err = retrieveShipment(stub, carton.ShipmentIds[i])
		if err != nil {
			fmt.Println("Error while retrieveing the Shipment Details", err)
			return nil, err
		}
		shipmentArr = append(shipmentArr, tmpShipment)
	}

	resp.ShipmentDetail = shipmentArr
	return json.Marshal(resp)

}

func SearchCarton(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering SearchCarton " + args[0])
	var carton Carton
	var err error
	var request SearchCartonRequest
	var resp SearchCartonResponse

	request, err = parseSearchCartonRequest(args[0])
	if err != nil {
		resp.ErrCode = INVALID_INPUT_ERROR_CODE
		resp.ErrMessage = INVALID_INPUT_ERROR_MSG
		return json.Marshal(resp)
	}

	carton, err = parseCarton(stub, request.CartonId)

	if err != nil {
		resp.ErrCode = NODATA_ERROR_CODE
		resp.ErrMessage = NODATA_ERROR_MSG
		return json.Marshal(resp)
	}

	return PrepareSearchCartontResponse(stub, carton)

}

/************** Carton Search Service Ends ************************/

/************** Pallet Search Service Starts ************************/

type SearchPalletRequest struct {
	CallingEntityName string `json:"callingEntityName"`
	PalletId          string `json:"palletId"`
}

type SearchPalletResponse struct {
	PalletId       string     `json:"palletId"`
	ShipmentDetail []Shipment `json:"shipmentDetail"`
	ErrCode        string     `json:"errCode"`
	ErrMessage     string     `json:"errMessage"`
}

func parseSearchPalletRequest(requestParam string) (SearchPalletRequest, error) {
	var request SearchPalletRequest

	if marshErr := json.Unmarshal([]byte(requestParam), &request); marshErr != nil {
		fmt.Println("Could not Unmarshal Asset", marshErr)
		return request, marshErr
	}
	return request, nil

}

func parsePallet(stub shim.ChaincodeStubInterface, palletId string) (Pallet, error) {

	var pallet Pallet

	palletBytes, err := stub.GetState(palletId)
	if err != nil {
		return pallet, err
	} else {
		if marshErr := json.Unmarshal(palletBytes, &pallet); marshErr != nil {
			fmt.Println("Could not Unmarshal Asset", marshErr)
			return pallet, marshErr
		}
		return pallet, nil
	}

}

func PrepareSearchPalletResponse(stub shim.ChaincodeStubInterface, pallet Pallet) ([]byte, error) {
	var resp SearchPalletResponse
	var shipmentArr []Shipment
	var tmpShipment Shipment
	var err error

	resp.PalletId = pallet.PalletId

	lenOfArray := len(pallet.ShipmentIds)

	for i := 0; i < lenOfArray; i++ {
		tmpShipment, err = retrieveShipment(stub, pallet.ShipmentIds[i])
		if err != nil {
			fmt.Println("Error while retrieveing the Shipment Details", err)
			return nil, err
		}
		shipmentArr = append(shipmentArr, tmpShipment)
	}

	resp.ShipmentDetail = shipmentArr
	return json.Marshal(resp)

}

func SearchPallet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering SearchPallet " + args[0])
	var pallet Pallet
	var err error
	var request SearchPalletRequest
	var resp SearchPalletResponse

	request, err = parseSearchPalletRequest(args[0])
	if err != nil {
		resp.ErrCode = INVALID_INPUT_ERROR_CODE
		resp.ErrMessage = INVALID_INPUT_ERROR_MSG
		return json.Marshal(resp)
	}

	pallet, err = parsePallet(stub, request.PalletId)

	if err != nil {
		resp.ErrCode = NODATA_ERROR_CODE
		resp.ErrMessage = NODATA_ERROR_MSG
		return json.Marshal(resp)
	}

	return PrepareSearchPalletResponse(stub, pallet)

}

/************** Pallet Search Service Ends ************************/

/************** Date Search Service Starts ************************/

type SearchDateRequest struct {
	CallingEntityName string `json:"callingEntityName"`
	StartDate         string `json:"startDate"`
	EndDate           string `json:"endDate"`
}

type SearchDateResponse struct {
	ShipmentDetail []Shipment `json:"shipmentDetail"`
}

func parseAllShipmentDump() (AllShipmentDump, error) {
	var dump AllShipmentDump

	if marshErr := json.Unmarshal([]byte("ALL_SHIPMENT_DUMP"), &dump); marshErr != nil {
		fmt.Println("Could not Unmarshal Asset", marshErr)
		return dump, marshErr
	}
	return dump, nil

}

func SearchDateRange(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	//var shipmentDump AllShipmentDump
	var err error
	var shipmentArr []Shipment
	var tmpShipment Shipment
	var resp SearchDateResponse

	/*shipmentDump, err = parseAllShipmentDump()
	if err != nil {
		return nil, err
	}*/

	lenOfArray := len(args)

	for i := 0; i < lenOfArray; i++ {
		tmpShipment, err = retrieveShipment(stub, args[i])
		if err != nil {
			fmt.Println("Error while retrieveing the Shipment Details", err)
			return nil, err
		}
		shipmentArr = append(shipmentArr, tmpShipment)
	}
	resp.ShipmentDetail = shipmentArr

	return json.Marshal(resp)

}

/************** Date Search Service Ends ************************/

/************** View Data for Key Starts ************************/

func ViewDataForKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering ViewDataForKey " + args[0])

	return stub.GetState(args[0])

}

/************** View Data for Key Ends ************************/

/************** DumpData Start ************************/

func DumpData(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Entering DumpData " + args[0] + "  " + args[1])

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		fmt.Println("Could not save the Data", err)
		return nil, err
	}

	return nil, nil
}

/************** DumpData Ends ************************/

func Initialize(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// Init resets all the things
func (t *B4SCChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Inside INIT for test chaincode")

	allShipment := AllShipment{}
	var tmpShipmentIndex []ShipmentIndex
	allShipment.ShipmentIndexArr = tmpShipmentIndex

	dataToStore, _ := json.Marshal(allShipment)

	err := stub.PutState("SHIPMENT_INDEX", []byte(dataToStore))
	if err != nil {
		fmt.Println("Could not save Shipment to ledger", err)
		return nil, err
	}

	return nil, nil
}

func (t *B4SCChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	/*if function == "Init" {
		return Init(stub, function, args)
	}else*/
	if function == "CreateShipment" {
		return CreateShipment(stub, args)
	} else if function == "DumpData" {
		return DumpData(stub, args)
	} else if function == "CreateShipment" {
		return CreateShipment(stub, args)
	} else if function == "CreateWayBill" {
		return CreateWayBill(stub, args)
	} else if function == "CreateDCShipment" {
		return CreateDCShipment(stub, args)
	} else if function == "CreateDCWayBill" {
		return CreateDCWayBill(stub, args)
	} else if function == "CreateEWWayBill" {
		return CreateEWWayBill(stub, args)
	} else if function == "CreateEntityWayBillMapping" {
		return nil, nil //CreateEntityWayBillMapping(stub, args)
	} else if function == "CreateAsset" {
		return CreateAsset(stub, args)
	} else if function == "CreateCarton" {
		return CreateCarton(stub, args)
	} else if function == "CreatePallet" {
		return CreatePallet(stub, args)
	} else if function == "UpdateAssetDetails" {
		return UpdateAssetDetails(stub, args)
	} else if function == "UpdateCartonDetails" {
		return UpdateCartonDetails(stub, args)
	} else if function == "UpdatePalletDetails" {
		return UpdatePalletDetails(stub, args)
	} else if function == "uploadComplianceDocument" {
		return uploadComplianceDocument(stub, args)
	} else if function == "UpdateEntityWayBillMapping" {
		return nil, nil //UpdateEntityWayBillMapping(stub, args)
	} else {
		return nil, errors.New("Invalid function name " + function)
	}
	//return nil, nil getComplianceDocumentByEntityName getAllComplianceDocument
}

func (t *B4SCChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "ViewShipment" {
		return ViewShipment(stub, args)
	} else if function == "ViewDataForKey" {
		return ViewDataForKey(stub, args)
	} else if function == "Inbox" {
		var inboxService InboxService
		return inboxService.Inbox(stub, args)
	} else if function == "SearchAsset" {
		return SearchAsset(stub, args)
	} else if function == "SearchCarton" {
		return SearchCarton(stub, args)
	} else if function == "SearchPallet" {
		return SearchPallet(stub, args)
	} else if function == "SearchDateRange" {
		return SearchDateRange(stub, args)
	} else if function == "ShipmentPageLoad" {
		var pageLoadService ShipmentPageLoadService
		return pageLoadService.ShipmentPageLoad(stub, args)
	} else if function == "ViewEWWayBill" {
		return nil, nil //ViewEWWayBill(stub, args)
	} else if function == "ViewEWWayBill" {
		return nil, nil //ViewEWWayBill(stub, args)
	} else if function == "GetEntityWayBillMapping" {
		return GetEntityWayBillMapping(stub, args)
	} else if function == "GetAsset" {
		return GetAsset(stub, args)
	} else if function == "GetPallet" {
		return GetPallet(stub, args)
	} else if function == "GetCarton" {
		return GetCarton(stub, args)
	} else if function == "ViewShipmentWayBill" {
		return ViewShipmentWayBill(stub, args)
	} else if function == "getComplianceDocumentByEntityName" {
		return getComplianceDocumentByEntityName(stub, args)
	} else if function == "getAllComplianceDocument" {
		return getAllComplianceDocument(stub, args)
	}
	return nil, errors.New("Invalid function name " + function)

}

func main() {
	Initialize(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	err := shim.Start(new(B4SCChaincode))
	if err != nil {
		fmt.Println("Could not start B4SCChaincode")
	} else {
		fmt.Println("B4SCChaincode successfully started")
	}
}
