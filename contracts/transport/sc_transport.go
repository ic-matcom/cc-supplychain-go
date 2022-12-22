package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//Contrat for writing and reading Transport from world-state
type TransportSmartContract struct {
	contractapi.Contract
}

const transport_index = "transport"

// Create and put new Transport to world-state
func (sc *TransportSmartContract) CreateTransport(ctx CustomTransactionContextInterface, key string,
	capacityUnits string, capacityUnitMeasurement string, capacityUsed string,
	country string, province string, lat string, lon string, transportType string) error {

	err := ValidRole(ctx, "abac.administrator")

	if err != nil {
		return err
	}

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport != nil {
		return fmt.Errorf("Cannot create new transport in world state as key %s alredy exists", key)
	}

	// Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	//Get ID of submitting client's org
	ownerID, err := GetSubmittingClientOrg(ctx)
	if err != nil {
		return err
	}

	// Verify that this client belongs to the peer's org
	err = VerifyClientOrgMatchesPeerOrg(ownerID)
	if err != nil {
		return err
	}

	transport := new(Transport)

	transport.DocType = "transport"
	transport.TransportID = key
	transport.AdvisorID = advisor
	transport.OwnerID = ownerID
	transport.Capacity = Capacity{transport.Capacity.Units, transport.Capacity.UnitMeasurement, capacityUsed}
	transport.Location = Location{country, province, lat, lon}
	transport.Certifications = make([]Certification, 0)
	transport.Shipment = make([]string, 0)
	transport.TransportType = transportType

	transport.TransportIsAvailable()

	//Obtain json encoding
	transportJson, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportJson))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//  This will enable very efficient state range queries based on composite keys matching indexName~product~transportr*
	transportIndexKey, err := ctx.GetStub().CreateCompositeKey(transport_index, []string{transport.TransportID, transport.OwnerID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	//put composite key of transport in world statte
	err = ctx.GetStub().PutState(transportIndexKey, value)

	return nil
}

// Inspect the Transport with the given key from world-state
func (sc *TransportSmartContract) TransportInspection(ctx CustomTransactionContextInterface, key string,
	issuer string, certificationtype string, result string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 0 {
		return fmt.Errorf("Transport with key %s is not available", key)
	}

	//Add certification to trasnport certifications

	transport.Certifications = append(transport.Certifications, Certification{issuer, certificationtype, result})

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Change Advisor of the the Transport with the given key from world-state
func (sc *TransportSmartContract) UpdateTransportAdvisor(ctx CustomTransactionContextInterface, key string, newAdvisor string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify trasnport state
	if transport.TransportState != 4 {
		return fmt.Errorf("Transport with key %s is not available", key)
	}

	transport.AdvisorID = newAdvisor

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Load lotes on the Transport with the given key from world-state
func (sc *TransportSmartContract) LoadLoteInTransport(ctx CustomTransactionContextInterface, key string,
	capacityUsed string, lotes string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 0 {
		return fmt.Errorf("Transport with key %s is not available", key)
	}

	//Verify Lote advisor
	//Instance smart contract
	loteContract := new(LoteSmartContract)

	if lotes == "" {
		return fmt.Errorf("Does not exist lotes to load")
	}
	lote := strings.Split(lotes, ",")

	for i := 0; i < len(lote); i++ {
		//Verify authorization
		loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, lote[i])
		if err != nil {
			return err
		}
		if advisor != loteAdvisor {
			return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
		}
	}

	transport.Capacity = Capacity{transport.Capacity.Units, transport.Capacity.UnitMeasurement, capacityUsed}

	//Add lote to transport shipment

	transport.Shipment = append(transport.Shipment, lote...)

	transport.TransportIsLoading()

	transportBytes, _ := json.Marshal(transport)

	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Start delivery of the Transport with the given key from world-state
func (sc *TransportSmartContract) StartTransportDelivery(ctx CustomTransactionContextInterface, key string,
	destCountry string, destProvince string, destLat string, desLon string, finishTime string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 1 {
		return fmt.Errorf("Transport with key %s is not loading", key)
	}

	transport.Transporting = Transporting{Location{destCountry, destProvince, destLat, desLon}, finishTime}
	transport.TransportIsDelivering()

	transportBytes, _ := json.Marshal(transport)

	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Inspect a lote in Transport with the given key from world-state
func (sc *TransportSmartContract) TransportInspectionLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, issuer string, certificationtype string, result string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 2 {
		return fmt.Errorf("Transport with key %s is not delivering", key)
	}

	//Verify Lote advisor
	//Instance smart contract
	loteContract := new(LoteSmartContract)

	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	err = loteContract.LoteInspection(ctx, loteKey, issuer, certificationtype, result)

	if err != nil {
		return err
	}
	/*
		//Obtain json encoding
		transportBytes, _ := json.Marshal(transport)

		//put transport state in world state
		err = ctx.GetStub().PutState(key, []byte(transportBytes))

		if err != nil {
			return errors.New("Unable to interact with world state")
		}
	*/
	return nil
}

// Deliver a lote from Transport to a store with the given key from world-state
func (sc *TransportSmartContract) TransportDeliverStoreLote(ctx CustomTransactionContextInterface, key string, loteKey string,
	nextLocation string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string,
	capacityUsed string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 2 {
		return fmt.Errorf("Transport with key %s is not delivering", key)
	}

	//Verify Lote advisor
	//Instance smart contract
	loteContract := new(LoteSmartContract)

	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	index_finded := -1
	for i := 0; i < len(transport.Shipment); i++ {
		if loteKey == transport.Shipment[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.StoreLote(ctx, loteKey, nextLocation, newadvisor, envTemp, envTempMeasurement, envHumidity)

	if err != nil {
		return err
	}

	transport.Shipment = append(transport.Shipment[:index_finded], transport.Shipment[index_finded+1:]...)

	if len(transport.Shipment) == 0 {
		transport.TransportIsAvailable()
	}

	transport.Capacity = Capacity{transport.Capacity.Units, transport.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Deliver a lote for reparing from Transport to a manufacture with the given key from world-state
func (sc *TransportSmartContract) TransportDeliverRepairingLote(ctx CustomTransactionContextInterface, key string, loteKey string,
	newOwner string, nextLocation string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string,
	capacityUsed string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 2 {
		return fmt.Errorf("Transport with key %s is not delivering", key)
	}

	//Verify Lote advisor
	//Instance smart contract
	loteContract := new(LoteSmartContract)

	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	index_finded := -1
	for i := 0; i < len(transport.Shipment); i++ {
		if loteKey == transport.Shipment[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.PutLoteOnReparation(ctx, loteKey, nextLocation, newadvisor, newOwner, envTemp, envTempMeasurement, envHumidity)

	if err != nil {
		return err
	}

	transport.Shipment = append(transport.Shipment[:index_finded], transport.Shipment[index_finded+1:]...)

	if len(transport.Shipment) == 0 {
		transport.TransportIsAvailable()
	}

	transport.Capacity = Capacity{transport.Capacity.Units, transport.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Deliver a lote for using from Transport to a manufacture with the given key from world-state
func (sc *TransportSmartContract) TransportDeliverUsingLote(ctx CustomTransactionContextInterface, key string, loteKey string,
	nextLocation string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string, capacityUsed string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 2 {
		return fmt.Errorf("Transport with key %s is not delivering", key)
	}

	//Verify Lote advisor
	//Instance smart contract
	loteContract := new(LoteSmartContract)

	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	index_finded := -1
	for i := 0; i < len(transport.Shipment); i++ {
		if loteKey == transport.Shipment[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.PutLoteOnUsing(ctx, loteKey, nextLocation, newadvisor, envTemp, envTempMeasurement, envHumidity)

	if err != nil {
		return err
	}

	transport.Shipment = append(transport.Shipment[:index_finded], transport.Shipment[index_finded+1:]...)

	if len(transport.Shipment) == 0 {
		transport.TransportIsAvailable()
	}

	transport.Capacity = Capacity{transport.Capacity.Units, transport.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Deliver a lote from Transport to a store transfering for reparing with the given key from world-state
func (sc *TransportSmartContract) TransportDeliverTransferLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, newOwner string, nextLocation string, priceamount string, pricecurrency string, newadvisor string,
	envTemp string, envTempMeasurement string, envHumidity string, capacityUsed string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 2 {
		return fmt.Errorf("Transport with key %s is not delivering", key)
	}

	//Verify Lote advisor
	//Instance smart contract
	loteContract := new(LoteSmartContract)

	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	index_finded := -1
	for i := 0; i < len(transport.Shipment); i++ {
		if loteKey == transport.Shipment[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.TransferLote(ctx, loteKey, nextLocation, newadvisor, newOwner, priceamount, pricecurrency, envTemp, envTempMeasurement, envHumidity)

	if err != nil {
		return err
	}

	transport.Shipment = append(transport.Shipment[:index_finded], transport.Shipment[index_finded+1:]...)

	if len(transport.Shipment) == 0 {
		transport.TransportIsAvailable()
	}

	transport.Capacity = Capacity{transport.Capacity.Units, transport.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Deliver a lote from Transport to a other transport with the given key from world-state
func (sc *TransportSmartContract) TransportDeliverTransshipmentLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, nextLocation string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string,
	capacityUsed string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 2 {
		return fmt.Errorf("Transport with key %s is not delivering", key)
	}

	//Verify Lote advisor
	//Instance smart contract
	loteContract := new(LoteSmartContract)

	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	index_finded := -1
	for i := 0; i < len(transport.Shipment); i++ {
		if loteKey == transport.Shipment[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.TransshipmentLote(ctx, loteKey, nextLocation, newadvisor, envTemp, envTempMeasurement, envHumidity)

	if err != nil {
		return err
	}

	transport.Shipment = append(transport.Shipment[:index_finded], transport.Shipment[index_finded+1:]...)

	if len(transport.Shipment) == 0 {
		transport.TransportIsAvailable()
	}

	transport.Capacity = Capacity{transport.Capacity.Units, transport.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Is non available the Transport with the given key from world-state
func (sc *TransportSmartContract) TransportIsNonAvailable(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 0 {
		return fmt.Errorf("Transport with key %s is not destroyed", key)
	}

	transport.TransportIsNonAvailable()

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Is available the Transport with the given key from world-state
func (sc *TransportSmartContract) TransportIsAvailable(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 3 {
		return fmt.Errorf("Transport with key %s is not a non available transport", key)
	}

	transport.TransportIsAvailable()

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// update location of the Transport with the given key from world-state
func (sc *TransportSmartContract) TransportUpdateLocation(ctx CustomTransactionContextInterface, key string,
	country string, province string, lat string, lon string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState != 4 {
		return fmt.Errorf("Transport with key %s is not destroyed", key)
	}

	transport.Location = Location{country, province, lat, lon}

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	//put transport state in world state
	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Destroy Transport with the given key from world-state
func (sc *TransportSmartContract) TransportDestroyed(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current transport with the given key
	currentTransport := ctx.GetData()

	//Verify if current transport exists
	if currentTransport == nil {
		return fmt.Errorf("Cannot update transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != transport.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify transport state
	if transport.TransportState == 3 {
		return fmt.Errorf("Transport with key %s is not non available", key)
	}

	transport.TransportIsDestroyed()

	//Obtain json encoding
	transportBytes, _ := json.Marshal(transport)

	err = ctx.GetStub().PutState(key, []byte(transportBytes))

	//put transport state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//Get index entry
	transportIndexKey, err := ctx.GetStub().CreateCompositeKey(transport_index, []string{transport.TransportID, transport.OwnerID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(transportIndexKey)
}

// GetTransportHistory returns the chain of custody for an transport since issuance.
func (sc *TransportSmartContract) GetTransportHistory(ctx contractapi.TransactionContextInterface, key string) ([]HistoryQueryResultTransport, error) {
	log.Printf("GetTransportHistory: ID %v", key)

	//Get History of the transport with the given key
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResultTransport
	//For each record in history takes the record value and add to the history
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var transport = new(Transport)
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, transport)
			if err != nil {
				return nil, err
			}
		} else {
			transport.TransportID = key
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResultTransport{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    transport,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// Get the value of the Transport with the given key from world-state
func (sc *TransportSmartContract) GetTransport(ctx CustomTransactionContextInterface, key string) (*Transport, error) {

	currentTransport := ctx.GetData()

	if currentTransport == nil {
		return nil, fmt.Errorf("Cannot read transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	return transport, nil
}

// Get lotes delivering with the Transport with the given key from world-state
func (sc *TransportSmartContract) GetTransportDeliveringLotes(ctx CustomTransactionContextInterface, key string) ([]string, error) {

	currentTransport := ctx.GetData()

	if currentTransport == nil {
		return nil, fmt.Errorf("Cannot read transport in world state as key %s does not exist", key)
	}

	transport := new(Transport)

	err := json.Unmarshal(currentTransport, transport)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Transport", key)
	}

	return transport.Shipment, nil
}

// GetEvaluateTransactions returns functions of SimpleContract not to be tagged as submit
func (sc *TransportSmartContract) GetEvaluateTransactions() []string {
	return []string{"GetTransport", "GetTransportDeliveringLotes", "GetTransportHistory"}
}
