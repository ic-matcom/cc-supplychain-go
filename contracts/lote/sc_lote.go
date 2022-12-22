package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//Contrat for writing and reading lote from world-state
type LoteSmartContract struct {
	contractapi.Contract
}

const lote_index = "lote"

// Create and put new lote to world-state
func (sc *LoteSmartContract) CreateLote(ctx CustomTransactionContextInterface, key string, productID string, manufactureID string, priceamount string, pricecurrency string,
	units string, envTemp string, envTempMeasurement string, envHumidity string,
	components []string, currentLocationID string) error {

	err := ValidRole(ctx, "abac.manufacturer")

	if err != nil {
		return err
	}

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote != nil {
		return fmt.Errorf("Cannot create new lote in world state as key %s alredy exists", key)
	}

	//Verify if current location exists
	currentLocation, err := ctx.GetStub().GetState(currentLocationID)
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if currentLocation exists
	if currentLocation == nil {
		return fmt.Errorf("Location with key %s do not exists", key)
	}

	unit, _ := strconv.Atoi(units)
	//If units < 1 them return error
	if unit < 1 {
		return fmt.Errorf("Cannot create new lote in world state as key %s with less than 1 unit", key)
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

	//Create a new struct of lote
	lote := new(Lote)

	lote.DocType = "lote"
	lote.LoteID = key
	lote.ProductID = productID
	lote.ManufactureID = manufactureID
	lote.Advisor = advisor
	lote.OwnerID = ownerID
	lote.Price = Price{priceamount, pricecurrency}
	lote.Units = units
	lote.Certifications = make([]Certification, 0)
	lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
	lote.Components = components
	lote.CurrentLocationID = currentLocationID
	lote.FatherID = ""

	lote.LoteIsManufacturing()
	lote.LoteProductIsOk()

	//Obtain json encoding
	loteJson, _ := json.Marshal(lote)

	//put lote state in world state
	err = ctx.GetStub().PutState(key, []byte(loteJson))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//  This will enable very efficient state range queries based on composite keys matching indexName~product~manufacturer*
	loteIndexKey, err := ctx.GetStub().CreateCompositeKey(lote_index, []string{lote.LoteID, lote.ProductID, lote.ManufactureID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	//put composite key of lote in world statte
	err = ctx.GetStub().PutState(loteIndexKey, value)

	return nil
}

//If lote exists return true
func (sc *LoteSmartContract) ExistsLote(ctx CustomTransactionContextInterface, key string) bool {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return false
	}
	return currentLote != nil
}

//  inspection of the lote with the given key from world-state
func (sc *LoteSmartContract) LoteInspection(ctx CustomTransactionContextInterface, key string,
	issuer string, certificationtype string, result string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify lote state
	if lote.LoteState == 4 || lote.LoteState == 5 {
		return fmt.Errorf("Lote with key %s is selled or destroyed", key)
	}

	lote.Certifications = append(lote.Certifications, Certification{issuer, certificationtype, result})

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// Store the lote with the given key from world-state
func (sc *LoteSmartContract) StoreLote(ctx CustomTransactionContextInterface, key string,
	currentLocationID string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	//Verify if current location exists
	currentLocation, err := ctx.GetStub().GetState(currentLocationID)
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if currentLocation exists
	if currentLocation == nil {
		return fmt.Errorf("Location with key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify lote state
	if lote.LoteState != 1 {
		return fmt.Errorf("Lote with key %s is not transporting", key)
	}

	lote.LoteIsStored()

	lote.CurrentLocationID = currentLocationID
	lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
	lote.Advisor = newadvisor

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// Start transporting lote with the given key from world-state
func (sc *LoteSmartContract) TransportLote(ctx CustomTransactionContextInterface, key string,
	currentLocationID string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	//Verify if current location exists
	currentLocation, err := ctx.GetStub().GetState(currentLocationID)
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if currentLocation exists
	if currentLocation == nil {
		return fmt.Errorf("Location with key %s do not exists", currentLocationID)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise current lote")
	}

	//Verify lote state
	if lote.LoteState == 4 || lote.LoteState == 5 {
		return fmt.Errorf("Lote with key %s is selled or destroyed", key)
	}

	lote.LoteIsTransporting()

	lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
	lote.CurrentLocationID = currentLocationID
	lote.Advisor = newadvisor

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// Fork lote with the given key from world-state and create new lote with equals attribites
func (sc *LoteSmartContract) FractionateLote(ctx CustomTransactionContextInterface,
	key string, newLote1Units string, newLote2Units string, newLote1ID string, newLote2ID string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify lote state
	if lote.LoteState != 2 && lote.LoteState != 6 && lote.LoteState != 0 {
		return fmt.Errorf("Lote with key %s is not stored", key)
	}

	unitslote1, _ := strconv.Atoi(newLote1Units)
	unitslote2, _ := strconv.Atoi(newLote2Units)

	unit, _ := strconv.Atoi(lote.Units)
	if unit != unitslote1+unitslote2 {
		return fmt.Errorf("Lote with key %s dont have enough units, lote 1 = %s, lote 2 = %s, total = %s", key, newLote1Units, newLote2Units, strconv.Itoa(unit))
	}

	newLote := new(Lote)

	newLote.DocType = "lote"
	newLote.LoteID = newLote1ID
	newLote.ProductID = lote.ProductID
	newLote.ManufactureID = lote.ManufactureID
	newLote.OwnerID = lote.OwnerID
	newLote.Price = lote.Price
	newLote.Units = newLote1Units
	newLote.Certifications = lote.Certifications
	newLote.Environment = lote.Environment
	newLote.Components = lote.Components
	newLote.CurrentLocationID = lote.CurrentLocationID
	newLote.LoteState = lote.LoteState
	newLote.LoteProductState = lote.LoteProductState
	newLote.Advisor = advisor
	newLote.FatherID = key

	//Obtain json encoding
	newLoteJson, _ := json.Marshal(newLote)

	//put lote state in world state
	err = ctx.GetStub().PutState(newLote1ID, []byte(newLoteJson))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//  This will enable very efficient state range queries based on composite keys matching indexName~product~manufacturer*
	loteIndexKey, err := ctx.GetStub().CreateCompositeKey(lote_index, []string{newLote.LoteID, newLote.ProductID, newLote.ManufactureID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	//put composite key of lote in world statte
	err = ctx.GetStub().PutState(loteIndexKey, value)

	//Lote 2
	newLote.LoteID = newLote2ID
	newLote.Units = newLote2Units

	//Obtain json encoding
	newLoteJson, _ = json.Marshal(newLote)

	//put lote state in world state
	err = ctx.GetStub().PutState(newLote2ID, []byte(newLoteJson))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//  This will enable very efficient state range queries based on composite keys matching indexName~product~manufacturer*
	loteIndexKey, err = ctx.GetStub().CreateCompositeKey(lote_index, []string{newLote.LoteID, newLote.ProductID, newLote.ManufactureID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value = []byte{0x00}
	//put composite key of lote in world statte
	err = ctx.GetStub().PutState(loteIndexKey, value)

	return sc.LoteDestroyed(ctx, key)
}

// Transfer the lote with the given key from world-state
func (sc *LoteSmartContract) TransferLote(ctx CustomTransactionContextInterface, key string,
	currentLocationID string, newadvisor string, newowner string, priceamount string, pricecurrency string,
	envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	//Verify if current location exists
	currentLocation, err := ctx.GetStub().GetState(currentLocationID)
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if currentLocation exists
	if currentLocation == nil {
		return fmt.Errorf("Location with key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
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

	//Verify lote state
	if lote.LoteState != 1 {
		return fmt.Errorf("Lote with key %s is not transporting", key)
	}

	lote.LoteIsStored()

	lote.CurrentLocationID = currentLocationID
	lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
	lote.OwnerID = newowner
	lote.Price = Price{priceamount, pricecurrency}
	lote.Advisor = newadvisor

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// Store the lote with the given key from world-state
func (sc *LoteSmartContract) TransshipmentLote(ctx CustomTransactionContextInterface, key string,
	currentLocationID string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	//Verify if current location exists
	currentLocation, err := ctx.GetStub().GetState(currentLocationID)
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if currentLocation exists
	if currentLocation == nil {
		return fmt.Errorf("Location with key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify lote state
	if lote.LoteState != 1 {
		return fmt.Errorf("Lote with key %s is not transporting", key)
	}

	lote.CurrentLocationID = currentLocationID
	lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
	lote.Advisor = newadvisor

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// Return the lote with the given key from world-state
func (sc *LoteSmartContract) ReturnLote(ctx CustomTransactionContextInterface, key string,
	currentLocationID string, newadvisor string, newowner string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	//Verify if current location exists
	currentLocation, err := ctx.GetStub().GetState(currentLocationID)
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if currentLocation exists
	if currentLocation == nil {
		return fmt.Errorf("Location with key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
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

	//Verify lote state
	if lote.LoteState == 4 {
		lote.LoteIsStored()
		lote.LoteProductIsBroken()

		lote.CurrentLocationID = currentLocationID
		lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
		lote.OwnerID = newowner
		lote.Advisor = newadvisor

		//Obtain json encoding
		loteBytes, _ := json.Marshal(lote)

		err = ctx.GetStub().PutState(key, []byte(loteBytes))

		//put lote state in world state
		if err != nil {
			return errors.New("Unable to interact with world state")
		}
		return nil
	} else if lote.LoteState == 2 {
		lote.LoteIsTransporting()
		lote.LoteProductIsBroken()

		lote.CurrentLocationID = currentLocationID
		lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
		lote.OwnerID = newowner
		lote.Advisor = newadvisor

		//Obtain json encoding
		loteBytes, _ := json.Marshal(lote)

		err = ctx.GetStub().PutState(key, []byte(loteBytes))

		//put lote state in world state
		if err != nil {
			return errors.New("Unable to interact with world state")
		}
		return nil
	}
	return fmt.Errorf("Lote with key %s is not selled or stored", key)
}

// Put On reparation the lote with the given key from world-state
func (sc *LoteSmartContract) PutLoteOnReparation(ctx CustomTransactionContextInterface, key string,
	currentLocationID string, newadvisor string, newowner string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	//Verify if current location exists
	currentLocation, err := ctx.GetStub().GetState(currentLocationID)
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if currentLocation exists
	if currentLocation == nil {
		return fmt.Errorf("Location with key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
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

	//Verify lote state
	if lote.LoteState != 1 {
		return fmt.Errorf("Lote with key %s is not transporting", key)
	}
	if lote.LoteProductState != 1 {
		return fmt.Errorf("Lote with key %s is not broken", key)
	}

	lote.LoteIsRepairing()

	lote.CurrentLocationID = currentLocationID
	lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
	lote.OwnerID = newowner
	lote.Advisor = newadvisor

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// Start transporting lote with the given key from world-state
func (sc *LoteSmartContract) RepairLote(ctx CustomTransactionContextInterface, key string,
	currentLocationID string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	//Verify if current location exists
	currentLocation, err := ctx.GetStub().GetState(currentLocationID)
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if currentLocation exists
	if currentLocation == nil {
		return fmt.Errorf("Location with key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify lote state
	if lote.LoteState != 3 {
		return fmt.Errorf("Lote with key %s is not reparing", key)
	}

	lote.LoteIsTransporting()
	lote.LoteProductIsOk()

	lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
	lote.CurrentLocationID = currentLocationID
	lote.Advisor = newadvisor

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// Put On using the lote with the given key from world-state
func (sc *LoteSmartContract) PutLoteOnUsing(ctx CustomTransactionContextInterface, key string,
	currentLocationID string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	//Verify if current location exists
	currentLocation, err := ctx.GetStub().GetState(currentLocationID)
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if currentLocation exists
	if currentLocation == nil {
		return fmt.Errorf("Location with key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify lote state
	if lote.LoteState != 1 && lote.LoteState != 0 {
		return fmt.Errorf("Lote with key %s is not transporting or manufacturing", key)
	}
	if lote.LoteProductState == 1 {
		return fmt.Errorf("Lote with key %s is broken", key)
	}

	lote.LoteIsUsing()

	lote.CurrentLocationID = currentLocationID
	lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}
	lote.Advisor = newadvisor

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// update the lote price with the given key from world-state
func (sc *LoteSmartContract) UpdatePrice(ctx CustomTransactionContextInterface, key string, priceamount string, pricecurrency string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	lote.Price = Price{priceamount, pricecurrency}

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// update the lote advisor with the given key from world-state
func (sc *LoteSmartContract) UpdateAdvisor(ctx CustomTransactionContextInterface, key string, newAdvisor string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	lote.Advisor = advisor

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// update the lote environment with the given key from world-state
func (sc *LoteSmartContract) UpdateEnvironment(ctx CustomTransactionContextInterface, key string,
	envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	lote.Environment = Environment{envTemp, envTempMeasurement, envHumidity}

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// Sell lote with the given key from world-state
func (sc *LoteSmartContract) SellLote(ctx CustomTransactionContextInterface, key string, priceamount string, pricecurrency string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify lote state
	if lote.LoteState != 2 {
		return fmt.Errorf("Lote with key %s is not stored", key)
	}

	lote.Price = Price{priceamount, pricecurrency}
	lote.LoteIsSelled()

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// Destroy lote with the given key from world-state
func (sc *LoteSmartContract) LoteDestroyed(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return fmt.Errorf("Cannot update lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != lote.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify lote state
	if lote.LoteState == 5 {
		return fmt.Errorf("Lote with key %s is destroyed", key)
	}

	lote.LoteIsDestroyed()

	//Obtain json encoding
	loteBytes, _ := json.Marshal(lote)

	err = ctx.GetStub().PutState(key, []byte(loteBytes))

	//put lote state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//Get index entry
	loteIndexKey, err := ctx.GetStub().CreateCompositeKey(lote_index, []string{lote.LoteID, lote.ProductID, lote.ManufactureID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(loteIndexKey)
}

// Get the value of the lote with the given key from world-state
func (sc *LoteSmartContract) GetLote(ctx CustomTransactionContextInterface, key string) (*Lote, error) {
	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return nil, errors.New("Unable to interact with world state")
	}
	if currentLote == nil {
		return nil, fmt.Errorf("Cannot read lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	return lote, nil
}

// Get units of the lote with the given key from world-state
func (sc *LoteSmartContract) GetLoteUnits(ctx CustomTransactionContextInterface, key string) (string, error) {
	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return "", errors.New("Unable to interact with world state")
	}
	if currentLote == nil {
		return "", fmt.Errorf("Cannot read lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return "", fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	return lote.Units, nil
}

// GetLoteHistory returns the chain of custody for an lote since issuance.
func (sc *LoteSmartContract) GetLoteHistory(ctx CustomTransactionContextInterface, key string) ([]HistoryQueryResultLote, error) {

	//Get History of the lote with the given key
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResultLote
	//For each record in history takes the record value and add to the history
	father := ""
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var lote = new(Lote)
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, lote)
			if err != nil {
				return nil, err
			}
			father = lote.FatherID
		} else {
			lote.LoteID = key
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResultLote{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    lote,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	if father != "" {
		father_history, err := sc.GetLoteHistory(ctx, father)
		if err != nil {
			return nil, err
		}
		records = append(records, father_history...)
	}

	return records, nil
}

func (sc LoteSmartContract) GetLoteTraceability(ctx CustomTransactionContextInterface, key string) (*TraceabilityQueryResultLote, error) {

	response := TraceabilityQueryResultLote{}

	//Get History of the lote with the given key
	history, err := sc.GetLoteHistory(ctx, key)
	if err != nil {
		return nil, err
	}
	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return nil, errors.New("Unable to interact with world state")
	}
	if currentLote == nil {
		return nil, fmt.Errorf("Cannot read lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	components := lote.Components

	// Get components of the current lote
	if len(components) == 0 {
		response.TracedLoteID = key
		response.HistoryResultLote = history
		return &response, nil
	}

	var childsTraceability []TraceabilityQueryResultLote

	//get traceability of the each lote of the componets list
	for i := 0; i < len(components); i++ {
		child, err := sc.GetLoteTraceability(ctx, components[i])

		if err != nil {
			return nil, err
		}

		childsTraceability = append(childsTraceability, *child)
	}

	//Return traceability
	response.TracedLoteID = key
	response.HistoryResultLote = history
	response.ChildsHistoryResultLote = childsTraceability

	return &response, nil
}

// Get the value of the lote with the given key from world-state
func (sc *LoteSmartContract) GetLoteAdvisor(ctx CustomTransactionContextInterface, key string) (string, error) {
	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return "", errors.New("Unable to interact with world state")
	}
	if currentLote == nil {
		return "", fmt.Errorf("Cannot read lote in world state as key %s does not exist", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return "", fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	return lote.Advisor, nil
}

// Get the value of the lote with the given key from world-state
func (sc *LoteSmartContract) IsLoteUsing(ctx CustomTransactionContextInterface, key string) (bool, error) {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return false, errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return false, fmt.Errorf("Cannot create new lote in world state as key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return false, fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	return lote.LoteState == 6, nil
}

// Get the value of the lote with the given key from world-state
func (sc *LoteSmartContract) IsLoteReparing(ctx CustomTransactionContextInterface, key string) (bool, error) {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return false, errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return false, fmt.Errorf("Cannot create new lote in world state as key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return false, fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	return lote.LoteState == 3, nil
}

// Get the value of the lote with the given key from world-state
func (sc *LoteSmartContract) IsLoteStored(ctx CustomTransactionContextInterface, key string) (bool, error) {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return false, errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return false, fmt.Errorf("Cannot create new lote in world state as key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return false, fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	return lote.LoteState == 2, nil
}

// Get the value of the lote with the given key from world-state
func (sc *LoteSmartContract) IsLoteManufacturing(ctx CustomTransactionContextInterface, key string) (bool, error) {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return false, errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return false, fmt.Errorf("Cannot create new lote in world state as key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return false, fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	return lote.LoteState == 0, nil
}

// Get the value of the lote with the given key from world-state
func (sc *LoteSmartContract) IsLoteSelled(ctx CustomTransactionContextInterface, key string) (bool, error) {

	//Obtain current lote with the given key
	//currentLote := ctx.GetData()
	currentLote, err := ctx.GetStub().GetState(key)

	if err != nil {
		return false, errors.New("Unable to interact with world state")
	}
	//Verify if current lote exists
	if currentLote == nil {
		return false, fmt.Errorf("Cannot create new lote in world state as key %s do not exists", key)
	}

	lote := new(Lote)

	err = json.Unmarshal(currentLote, lote)

	if err != nil {
		return false, fmt.Errorf("Data retrieved from world state for key %s was not of type Lote", key)
	}

	return lote.LoteState == 6, nil
}

// GetEvaluateTransactions returns functions of SimpleContract not to be tagged as submit
func (sc *LoteSmartContract) GetEvaluateTransactions() []string {
	return []string{"GetLote", "GetLoteHistory", "GetLoteTraceability", "GetLoteAdvisor",
		"IsLoteUsing", "IsLoteReparing", "IsLoteStored", "IsLoteSelled"}
}
