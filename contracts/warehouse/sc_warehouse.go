package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//Contrat for writing and reading Warehouse from world-state
type WarehouseSmartContract struct {
	contractapi.Contract
}

const warehouse_index = "warehouse"

// Create and put new Warehouse to world-state
func (sc *WarehouseSmartContract) CreateWarehouse(ctx CustomTransactionContextInterface, key string,
	capacityUnits string, capacityUnitMeasurement string, capacityUsed string, country string,
	province string, lat string, lon string) error {

	err := ValidRole(ctx, "abac.administrator")

	if err != nil {
		return err
	}

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse != nil {
		return fmt.Errorf("Cannot create new warehouse in world state as key %s already exists", key)
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

	warehouse := new(Warehouse)

	warehouse.DocType = "warehouse"
	warehouse.WarehouseID = key
	warehouse.OwnerID = ownerID
	warehouse.AdvisorID = advisor
	warehouse.Capacity = Capacity{capacityUnits, capacityUnitMeasurement, capacityUsed}
	warehouse.Location = Location{country, province, lat, lon}
	warehouse.Certifications = make([]Certification, 0)
	warehouse.Storing = make([]string, 0)

	warehouse.WarehouseIsWorking()

	//Obtain json encoding
	WarehouseJson, _ := json.Marshal(warehouse)

	//put manufacture state in world state
	err = ctx.GetStub().PutState(key, []byte(WarehouseJson))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//  This will enable very efficient state range queries based on composite keys matching indexName~product~warehouse*
	warehouseIndexKey, err := ctx.GetStub().CreateCompositeKey(warehouse_index, []string{warehouse.WarehouseID, warehouse.OwnerID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	//put composite key of manufacture in world statte
	err = ctx.GetStub().PutState(warehouseIndexKey, value)

	return nil
}

// inspect the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseInspection(ctx CustomTransactionContextInterface, key string,
	issuer string, certificationtype string, result string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState == 2 {
		return fmt.Errorf("Warehouse with key %s is not working", key)
	}

	warehouse.Certifications = append(warehouse.Certifications, Certification{issuer, certificationtype, result})

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put warehouse state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// inspect lote in the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseLoteInspection(ctx CustomTransactionContextInterface, key string,
	loteKey string, issuer string, certificationtype string, result string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 0 {
		return fmt.Errorf("Warehouse with key %s is not available", key)
	}

	//Instance smart contract
	loteContract := new(LoteSmartContract)

	//Verify authorization
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
		warehouseBytes, _ := json.Marshal(warehouse)

		//put warehouse state in world state
		err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

		if err != nil {
			return errors.New("Unable to interact with world state")
		}
	*/
	return nil
}

// store lote in the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseStoreLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, capacityUsed string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 0 {
		return fmt.Errorf("Warehouse with key %s is not working", key)
	}

	//Instance smart contract
	loteContract := new(LoteSmartContract)

	//Verify authorization
	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}
	//Verify lote is stored
	response, err := loteContract.IsLoteStored(ctx, loteKey)
	if !response {
		return fmt.Errorf("lote is not stored")
	}
	if err != nil {
		return err
	}

	//Add lote to trasnport shipment

	warehouse.Storing = append(warehouse.Storing, loteKey)

	warehouse.Capacity = Capacity{warehouse.Capacity.Units, warehouse.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put warehouse state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// recieve lote selled return in the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseRecieveSelledReturnLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, nextLocation string, newadvisor string, newOwner string, envTemp string,
	envTempMeasurement string, envHumidity string, capacityUsed string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 0 {
		return fmt.Errorf("Warehouse with key %s is not working", key)
	}

	//Instance smart contract
	loteContract := new(LoteSmartContract)

	//Verify authorization
	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}
	//Verify lote is selled
	response, err := loteContract.IsLoteSelled(ctx, loteKey)
	if !response {
		return fmt.Errorf("lote is not selled")
	}
	if err != nil {
		return err
	}

	err = loteContract.ReturnLote(ctx, loteKey, nextLocation, newadvisor, newOwner, envTemp, envTempMeasurement, envHumidity)
	if err != nil {
		return err
	}

	//Add lote to trasnport shipment

	warehouse.Storing = append(warehouse.Storing, loteKey)

	warehouse.Capacity = Capacity{warehouse.Capacity.Units, warehouse.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put warehouse state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// sell lote of the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseSellLote(ctx CustomTransactionContextInterface, key string, loteKey string,
	priceamount string, pricecurrency string, capacityUsed string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 0 {
		return fmt.Errorf("Warehouse with key %s is not working", key)
	}

	//Instance smart contract
	loteContract := new(LoteSmartContract)

	//Verify authorization
	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	index_finded := -1
	for i := 0; i < len(warehouse.Storing); i++ {
		if loteKey == warehouse.Storing[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.SellLote(ctx, loteKey, priceamount, pricecurrency)

	if err != nil {
		return err
	}

	warehouse.Storing = append(warehouse.Storing[:index_finded], warehouse.Storing[index_finded+1:]...)

	warehouse.Capacity = Capacity{warehouse.Capacity.Units, warehouse.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put warehouse state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// transport lote of the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseTransportLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, nextLocation string, newadvisor string, envTemp string,
	envTempMeasurement string, envHumidity string, capacityUsed string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 0 {
		return fmt.Errorf("Warehouse with key %s is not working", key)
	}

	//Instance smart contract
	loteContract := new(LoteSmartContract)

	//Verify authorization
	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	index_finded := -1
	for i := 0; i < len(warehouse.Storing); i++ {
		if loteKey == warehouse.Storing[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.TransportLote(ctx, loteKey, nextLocation, newadvisor, envTemp, envTempMeasurement, envHumidity)

	if err != nil {
		return err
	}

	warehouse.Storing = append(warehouse.Storing[:index_finded], warehouse.Storing[index_finded+1:]...)

	warehouse.Capacity = Capacity{warehouse.Capacity.Units, warehouse.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put warehouse state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// return lote of the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseReturnLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, nextLocation string, newadvisor string, newOwner string, envTemp string,
	envTempMeasurement string, envHumidity string, capacityUsed string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 0 {
		return fmt.Errorf("Warehouse with key %s is not working", key)
	}

	//Instance smart contract
	loteContract := new(LoteSmartContract)

	//Verify authorization
	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	index_finded := -1
	for i := 0; i < len(warehouse.Storing); i++ {
		if loteKey == warehouse.Storing[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.ReturnLote(ctx, loteKey, nextLocation, newadvisor, newOwner, envTemp, envTempMeasurement, envHumidity)

	if err != nil {
		return err
	}

	warehouse.Storing = append(warehouse.Storing[:index_finded], warehouse.Storing[index_finded+1:]...)

	warehouse.Capacity = Capacity{warehouse.Capacity.Units, warehouse.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put warehouse state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// destroy lote of the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseDestroyLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, capacityUsed string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 0 {
		return fmt.Errorf("Warehouse with key %s is not working", key)
	}

	//Instance smart contract
	loteContract := new(LoteSmartContract)

	//Verify authorization
	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	index_finded := -1
	for i := 0; i < len(warehouse.Storing); i++ {
		if loteKey == warehouse.Storing[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.LoteDestroyed(ctx, loteKey)

	if err != nil {
		return err
	}

	warehouse.Storing = append(warehouse.Storing[:index_finded], warehouse.Storing[index_finded+1:]...)

	warehouse.Capacity = Capacity{warehouse.Capacity.Units, warehouse.Capacity.UnitMeasurement, capacityUsed}

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put warehouse state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// update advisor of the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseUpdateAdvisor(ctx CustomTransactionContextInterface, key string, newadvisor string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 0 {
		return fmt.Errorf("Warehouse with key %s is not available", key)
	}

	warehouse.AdvisorID = newadvisor

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put manufacture state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// the Warehouse is available with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseAvailable(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 1 {
		return fmt.Errorf("Warehouse with key %s is not non available", key)
	}

	warehouse.WarehouseIsWorking()

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put manufacture state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// the Warehouse is non available with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseNonAvailable(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 0 {
		return fmt.Errorf("Warehouse with key %s is not available", key)
	}

	warehouse.WarehouseIsNonAvailable()
	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put manufacture state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Put on discontinued the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) WarehouseDestroyed(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current warehouse with the given key
	currentWarehouse := ctx.GetData()

	//Verify if current warehouse exists
	if currentWarehouse == nil {
		return fmt.Errorf("Cannot update warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != warehouse.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify warehouse state
	if warehouse.WarehouseState != 1 {
		return fmt.Errorf("Warehouse with key %s is not non available", key)
	}

	warehouse.WarehouseIsDestroyed()

	//Obtain json encoding
	warehouseBytes, _ := json.Marshal(warehouse)

	//put warehouse state in world state
	err = ctx.GetStub().PutState(key, []byte(warehouseBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//Get index entry
	warehouseIndexKey, err := ctx.GetStub().CreateCompositeKey(warehouse_index, []string{warehouse.WarehouseID, warehouse.OwnerID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(warehouseIndexKey)
}

// GetWarehouseHistory returns the chain of custody for an warehouse since issuance.
func (sc *WarehouseSmartContract) GetWarehouseHistory(ctx contractapi.TransactionContextInterface, key string) ([]HistoryQueryResultWarehouse, error) {
	log.Printf("GetWarehouseHistory: ID %v", key)

	//Get History of the warehouse with the given key
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResultWarehouse
	//For each record in history takes the record value and add to the history
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var warehouse = new(Warehouse)
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, warehouse)
			if err != nil {
				return nil, err
			}
		} else {
			warehouse.WarehouseID = key
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResultWarehouse{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    warehouse,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// Get the value of the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) GetWarehouse(ctx CustomTransactionContextInterface, key string) (*Warehouse, error) {

	currentWarehouse := ctx.GetData()

	if currentWarehouse == nil {
		return nil, fmt.Errorf("Cannot read warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	return warehouse, nil
}

// Get lotes stored in the Warehouse with the given key from world-state
func (sc *WarehouseSmartContract) GetWarehouseLoteStored(ctx CustomTransactionContextInterface, key string) ([]string, error) {

	currentWarehouse := ctx.GetData()

	if currentWarehouse == nil {
		return nil, fmt.Errorf("Cannot read warehouse in world state as key %s does not exist", key)
	}

	warehouse := new(Warehouse)

	err := json.Unmarshal(currentWarehouse, warehouse)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Warehouse", key)
	}

	return warehouse.Storing, nil
}

// GetEvaluateTransactions returns functions of SimpleContract not to be tagged as submit
func (sc *WarehouseSmartContract) GetEvaluateTransactions() []string {
	return []string{"GetWarehouse", "GetWarehouseLoteStored", "GetWarehouseHistory"}
}
