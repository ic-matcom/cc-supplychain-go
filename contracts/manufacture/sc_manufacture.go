package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//Contrat for writing and reading Manufacture from world-state
type ManufactureSmartContract struct {
	contractapi.Contract
}

const manufacture_index = "manufacture"

// Create and put new manufacture to world-state
func (sc *ManufactureSmartContract) CreateManufacture(ctx CustomTransactionContextInterface,
	key string, country string, province string, latitude string, longitude string) error {

	err := ValidRole(ctx, "abac.administrator")

	if err != nil {
		return err
	}

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture != nil {
		return fmt.Errorf("Cannot create new manufacture in world state as key %s already exists", key)
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

	manufacture := new(Manufacture)

	manufacture.DocType = "manufacture"
	manufacture.ManufactureID = key
	manufacture.OwnerID = ownerID
	manufacture.AdvisorID = advisor
	manufacture.Location = Location{Country: country, Province: province, Latitude: latitude, Longitude: longitude}
	manufacture.Certifications = make([]Certification, 0)
	manufacture.Production = make([]string, 0)
	manufacture.ToRepair = make([]string, 0)
	manufacture.ToUse = make([]string, 0)

	manufacture.ManufactureIsOnProduction()

	//Obtain json encoding
	manufactureJson, _ := json.Marshal(manufacture)

	//put manufacture state in world state
	err = ctx.GetStub().PutState(key, []byte(manufactureJson))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//  This will enable very efficient state range queries based on composite keys matching indexName~product~manufacturer*
	manufactureIndexKey, err := ctx.GetStub().CreateCompositeKey(manufacture_index, []string{manufacture.ManufactureID, manufacture.OwnerID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	//put composite key of manufacture in world statte
	err = ctx.GetStub().PutState(manufactureIndexKey, value)

	return nil
}

// Inspect the manufacture with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureInspection(ctx CustomTransactionContextInterface,
	key string, issuer string, certificationtype string, result string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState == 2 {
		return fmt.Errorf("Manufacture with key %s is destroyed", key)
	}

	certification := Certification{Issuer: issuer, CertificationType: certificationtype, Result: result}

	//Add certification to manufacture certifications

	manufacture.Certifications = append(manufacture.Certifications, certification)

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	//put manufacture state in world state
	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// create product with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureCreateProduct(ctx CustomTransactionContextInterface, key string,
	productKey string, productName string, description string, brandname string, brandlogo string, core string,
	variety string, productImage string, Weight string, WeightMeasurement string, Volumen string, VolumenMeasurement string,
	Units string, PackDimentionWidth string, PackDimentionHeight string, PackDimentionDeep string, PackDimentionMeasurement string,
	DisplaySpaceWidth string, DisplaySpaceHeight string, DisplaySpaceDeep string, DisplaySpaceMeasurement string,
	packageType string, manufactureDetails string, components []string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	productContract := new(ProductSmartContract)

	for i := 0; i < len(components); i++ {
		if !productContract.ExistsProduct(ctx, components[i]) {
			return fmt.Errorf("component  %s do not exists", components[i])
		}
	}

	err = productContract.CreateProduct(ctx, productKey, productName, description,
		brandname, brandlogo, core, variety, productImage, Weight, WeightMeasurement, Volumen,
		VolumenMeasurement, Units, PackDimentionWidth, PackDimentionHeight, PackDimentionDeep,
		PackDimentionMeasurement, DisplaySpaceWidth, DisplaySpaceHeight, DisplaySpaceDeep,
		DisplaySpaceMeasurement, packageType, manufactureDetails, key, components)

	if err != nil {
		return err
	}

	/* //Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	} */

	return nil
}

// Certify product with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureCertifyProduct(ctx CustomTransactionContextInterface, key string,
	productKey string, issuer string, certificationtype string, result string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	productContract := new(ProductSmartContract)

	//Verify authorization
	productAdvisor, err := productContract.GetProductAdvisor(ctx, productKey)
	if err != nil {
		return err
	}
	if advisor != productAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	err = productContract.Certify(ctx, productKey, issuer, certificationtype, result)

	if err != nil {
		return err
	}

	/* //Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	} */

	return nil
}

// start production product with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureStartProductionProduct(ctx CustomTransactionContextInterface, key string, productKey string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	productContract := new(ProductSmartContract)

	//Verify authorization
	productAdvisor, err := productContract.GetProductAdvisor(ctx, productKey)
	if err != nil {
		return err
	}
	if advisor != productAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	err = productContract.ProductOnProduction(ctx, productKey)

	if err != nil {
		return err
	}

	/* //Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	} */

	return nil
}

// discontinue product with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureDiscontinuetionProduct(ctx CustomTransactionContextInterface, key string, productKey string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	productContract := new(ProductSmartContract)

	err = productContract.ProductDiscontinued(ctx, productKey)

	//Verify authorization
	productAdvisor, err := productContract.GetProductAdvisor(ctx, productKey)
	if err != nil {
		return err
	}
	if advisor != productAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	if err != nil {
		return err
	}

	/* //Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	} */

	return nil
}

// continue product with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureContinueProduct(ctx CustomTransactionContextInterface, key string, productKey string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	productContract := new(ProductSmartContract)

	//Verify authorization
	productAdvisor, err := productContract.GetProductAdvisor(ctx, productKey)
	if err != nil {
		return err
	}
	if advisor != productAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	err = productContract.ProductContinued(ctx, productKey)

	if err != nil {
		return err
	}

	/* //Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	} */

	return nil
}

// Destroy product with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureDestroyProduct(ctx CustomTransactionContextInterface, key string, productKey string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	productContract := new(ProductSmartContract)

	//Verify authorization
	productAdvisor, err := productContract.GetProductAdvisor(ctx, productKey)
	if err != nil {
		return err
	}
	if advisor != productAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	err = productContract.ProductDestroyed(ctx, productKey)

	if err != nil {
		return err
	}

	/* //Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	} */

	return nil
}

// create lote with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureCreateLote(ctx CustomTransactionContextInterface,
	key string, loteKey string, productID string,
	priceamount string, pricecurrency string, units string, envTemp string,
	envTempMeasurement string, envHumidity string, components []string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
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

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	//Instance smart contracts
	productContract := new(ProductSmartContract)
	loteContract := new(LoteSmartContract)

	//Verify that the product exists
	if !productContract.ExistsProduct(ctx, productID) {
		return fmt.Errorf("product do not exists")
	}

	list_components_id := []string{}

	//Verify that exists the components and units are equal
	for i := 0; i < len(components); i++ {
		if components[i] == "" {
			continue
		}
		currentComponent_i := strings.Split(components[i], ":")
		currentComponent_i_id := currentComponent_i[0]
		currentComponent_i_units, _ := strconv.Atoi(currentComponent_i[1])

		if !loteContract.ExistsLote(ctx, currentComponent_i_id) {
			return fmt.Errorf("component do not exists")
		}

		//Verify componets from To Use array
		finded := false
		for j := 0; j < len(manufacture.ToUse); j++ {
			if currentComponent_i_id == manufacture.ToUse[j] {
				finded = true
			}
		}
		if !finded {
			return fmt.Errorf("component do not finded in lote whit key %s for use", currentComponent_i_id)
		}

		currentLoteUnits, err := loteContract.GetLoteUnits(ctx, currentComponent_i_id)
		if err != nil {
			return err
		}

		currentLoteUnits_int, _ := strconv.Atoi(currentLoteUnits)
		if currentLoteUnits_int < currentComponent_i_units {
			return fmt.Errorf("It is not enough lote units")
		}

		if currentLoteUnits_int != currentComponent_i_units {
			return fmt.Errorf("units in lote do not match with request, fractionate first")
		}
		list_components_id = append(list_components_id, currentComponent_i_id)
	}

	//Destroy the components
	for i := 0; i < len(list_components_id); i++ {
		if list_components_id[i] == "" {
			continue
		}
		err = loteContract.LoteDestroyed(ctx, list_components_id[i])
		if err != nil {
			return err
		}
		//Delete componets from To Use array
		for j := 0; j < len(manufacture.ToUse); j++ {
			if list_components_id[i] == manufacture.ToUse[j] {
				manufacture.ToUse = append(manufacture.ToUse[:j], manufacture.ToUse[j+1:]...)
			}
		}
	}

	err = loteContract.CreateLote(ctx, loteKey, productID, key, pricecurrency,
		priceamount, units, envTemp, envTempMeasurement, envHumidity, list_components_id, key)

	if err != nil {
		return err
	}

	manufacture.Production = append(manufacture.Production, loteKey)

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

func (sc *ManufactureSmartContract) ManufactureFractionateLoteForUse(ctx CustomTransactionContextInterface,
	key string, components string, newComponentNames string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
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

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	loteContract := new(LoteSmartContract)
	list_components := []string{}
	if components != "" {
		list_components = strings.Split(components, ",")
	}

	list_componentNames := []string{}
	if newComponentNames != "" {
		list_componentNames = strings.Split(newComponentNames, ",")
	}

	//Verify that exists the components
	for i := 0; i < len(list_components); i++ {
		if list_components[i] == "" {
			continue
		}
		currentComponent_i := strings.Split(list_components[i], ":")
		currentComponent_i_id := currentComponent_i[0]
		currentComponent_i_units_str := currentComponent_i[1]
		currentComponent_i_units, _ := strconv.Atoi(currentComponent_i[1])

		if !loteContract.ExistsLote(ctx, currentComponent_i_id) {
			return fmt.Errorf("component do not exists")
		}

		//Verify componets from To Use array
		finded := -1
		for j := 0; j < len(manufacture.ToUse); j++ {
			if currentComponent_i_id == manufacture.ToUse[j] {
				finded = j
			}
		}
		if finded == -1 {
			return fmt.Errorf("component do not finded in lote whit key %s for use", currentComponent_i_id)
		}

		currentLoteUnits, err := loteContract.GetLoteUnits(ctx, currentComponent_i_id)
		if err != nil {
			return err
		}

		currentLoteUnits_int, _ := strconv.Atoi(currentLoteUnits)
		if currentLoteUnits_int < currentComponent_i_units {
			return fmt.Errorf("It is not enough lote units")
		}

		if currentLoteUnits_int > currentComponent_i_units {
			new_component_units := strconv.Itoa(currentLoteUnits_int - currentComponent_i_units)
			err = loteContract.FractionateLote(ctx, currentComponent_i_id, currentComponent_i_units_str, new_component_units, list_componentNames[2*i], list_componentNames[2*i+1])
			if err != nil {
				return err
			}

			manufacture.ToUse = append(manufacture.ToUse[:finded], manufacture.ToUse[finded+1:]...)
			manufacture.ToUse = append(manufacture.ToUse, list_componentNames[2*i], list_componentNames[2*i+1])
		}
	}

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// put lote in using with the given key from world-state
func (sc *ManufactureSmartContract) ManufacturePutLoteInUsing(ctx CustomTransactionContextInterface, key string, loteKey string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
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

	//Verify lote is using
	response, err := loteContract.IsLoteUsing(ctx, loteKey)
	if err != nil {
		return err
	}
	if !response {
		return fmt.Errorf("lote with key %s is not using", loteKey)
	}

	//Add lote to use list

	manufacture.ToUse = append(manufacture.ToUse, loteKey)

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

func (sc *ManufactureSmartContract) ManufacturePutLoteFromProductionToUsing(ctx CustomTransactionContextInterface, key string, loteKey string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	//Verify if lote is manufacturing
	index_finded := -1
	for i := 0; i < len(manufacture.Production); i++ {
		if loteKey == manufacture.Production[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
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

	//Verify lote is manufacturing
	response, err := loteContract.IsLoteManufacturing(ctx, loteKey)
	if err != nil {
		return err
	}
	if !response {
		return fmt.Errorf("lote with key %s is not manufacturing", loteKey)
	}

	err = loteContract.PutLoteOnUsing(ctx, loteKey, manufacture.ManufactureID, manufacture.AdvisorID, envTemp, envTempMeasurement, envHumidity)
	if err != nil {
		return err
	}

	//Remove from production list
	manufacture.Production = append(manufacture.Production[:index_finded], manufacture.Production[index_finded+1:]...)

	//Add lote to touse list

	manufacture.ToUse = append(manufacture.ToUse, loteKey)

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// put lote in reparing with the given key from world-state
func (sc *ManufactureSmartContract) ManufacturePutLoteInReparing(ctx CustomTransactionContextInterface, key string, loteKey string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
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

	//Verify lote is reparing
	response, err := loteContract.IsLoteReparing(ctx, loteKey)
	if !response {
		return fmt.Errorf("lote is not reparing")
	}
	if err != nil {
		return err
	}

	//Add lote to trasnport shipment

	manufacture.ToRepair = append(manufacture.ToRepair, loteKey)

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// repair lote with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureRepairLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, newLocation string, newAdvisor string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
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
	for i := 0; i < len(manufacture.ToRepair); i++ {
		if loteKey == manufacture.ToRepair[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.RepairLote(ctx, loteKey, newLocation, newAdvisor, envTemp, envTempMeasurement, envHumidity)

	if err != nil {
		return err
	}

	manufacture.ToRepair = append(manufacture.ToRepair[:index_finded], manufacture.ToRepair[index_finded+1:]...)

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// certify lote with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureInspectionLote(ctx CustomTransactionContextInterface, key string,
	loteKey string, issuer string, certificationtype string, result string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
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

	/* //Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	*/
	return nil
}

// destroy lote with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureDestroyLote(ctx CustomTransactionContextInterface, key string, loteKey string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
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
	for i := 0; i < len(manufacture.ToRepair); i++ {
		if loteKey == manufacture.ToRepair[i] {
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

	manufacture.ToRepair = append(manufacture.ToRepair[:index_finded], manufacture.ToRepair[index_finded+1:]...)

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

// transport lote with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureTransportLote(ctx CustomTransactionContextInterface, key string, loteKey string,
	currentLocation string, newadvisor string, envTemp string, envTempMeasurement string, envHumidity string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise manufacture")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	//Instance smart contract
	loteContract := new(LoteSmartContract)

	//Verify authorization
	loteAdvisor, err := loteContract.GetLoteAdvisor(ctx, loteKey)
	if err != nil {
		return err
	}
	if advisor != loteAdvisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise lote")
	}
	index_finded := -1
	for i := 0; i < len(manufacture.Production); i++ {
		if loteKey == manufacture.Production[i] {
			index_finded = i
			continue
		}
	}
	if index_finded == -1 {
		return fmt.Errorf("component do not finded in lotes for use")
	}

	err = loteContract.TransportLote(ctx, loteKey, currentLocation, newadvisor, envTemp, envTempMeasurement, envHumidity)

	if err != nil {
		return err
	}

	manufacture.Production = append(manufacture.Production[:index_finded], manufacture.Production[index_finded+1:]...)

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// update advisor of the manufacture with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureUpdateAdvisor(ctx CustomTransactionContextInterface,
	key string, newAdvisor string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	manufacture.AdvisorID = newAdvisor

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	//put manufacture state in world state
	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// manufacture is broken with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureBroken(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 0 {
		return fmt.Errorf("Manufacture with key %s is not working", key)
	}

	manufacture.ManufactureIsBrokn()

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// manufacture is ready to production with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureReadyToProduce(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 1 {
		return fmt.Errorf("Manufacture with key %s is not broken", key)
	}

	manufacture.ManufactureIsOnProduction()

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	//put manufacture state in world state
	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// destroy manufacture with the given key from world-state
func (sc *ManufactureSmartContract) ManufactureDestroyed(ctx CustomTransactionContextInterface,
	key string) error {

	//Obtain current manufacture with the given key
	currentManufacture := ctx.GetData()

	//Verify if current manufacture exists
	if currentManufacture == nil {
		return fmt.Errorf("Cannot update manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	if advisor != manufacture.AdvisorID {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Verify manufacture state
	if manufacture.ManufactureState != 1 {
		return fmt.Errorf("Manufacture with key %s is not broken", key)
	}

	manufacture.ManufactureIsDestroyed()

	//Obtain json encoding
	manufactureBytes, _ := json.Marshal(manufacture)

	//put manufacture state in world state
	err = ctx.GetStub().PutState(key, []byte(manufactureBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//Get index entry
	manufactureIndexKey, err := ctx.GetStub().CreateCompositeKey(manufacture_index, []string{manufacture.ManufactureID, manufacture.OwnerID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(manufactureIndexKey)
}

// GetManufactureHistory returns the chain of custody for an manufacture since issuance.
func (sc *ManufactureSmartContract) GetManufactureHistory(ctx contractapi.TransactionContextInterface, key string) ([]HistoryQueryResultManufacture, error) {
	log.Printf("GetManufactoryHistory: ID %v", key)

	//Get History of the manufacture with the given key
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResultManufacture
	//For each record in history takes the record value and add to the history
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var manufacture = new(Manufacture)
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, manufacture)
			if err != nil {
				return nil, err
			}
		} else {
			manufacture.ManufactureID = key
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResultManufacture{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    manufacture,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// Get the value of the manufacture with the given key from world-state
func (sc *ManufactureSmartContract) GetManufacture(ctx CustomTransactionContextInterface, key string) (*Manufacture, error) {

	currentManufacture := ctx.GetData()

	if currentManufacture == nil {
		return nil, fmt.Errorf("Cannot read manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)
	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	return manufacture, nil
}

// Get lote repairing in the manufacture with the given key from world-state
func (sc *ManufactureSmartContract) GetManufactureRepairingLotes(ctx CustomTransactionContextInterface, key string) ([]string, error) {

	currentManufacture := ctx.GetData()

	if currentManufacture == nil {
		return nil, fmt.Errorf("Cannot read manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	return manufacture.ToRepair, nil
}

// Get lote producing in the manufacture with the given key from world-state
func (sc *ManufactureSmartContract) GetManufactureProducingLotes(ctx CustomTransactionContextInterface, key string) ([]string, error) {

	currentManufacture := ctx.GetData()

	if currentManufacture == nil {
		return nil, fmt.Errorf("Cannot read manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	return manufacture.Production, nil
}

// Get lote using in the manufacture with the given key from world-state
func (sc *ManufactureSmartContract) GetManufactureUsingLotes(ctx CustomTransactionContextInterface, key string) ([]string, error) {

	currentManufacture := ctx.GetData()

	if currentManufacture == nil {
		return nil, fmt.Errorf("Cannot read manufacture in world state as key %s does not exist", key)
	}

	manufacture := new(Manufacture)

	err := json.Unmarshal(currentManufacture, manufacture)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Manufacturer", key)
	}

	return manufacture.ToUse, nil
}

// GetEvaluateTransactions returns functions of SimpleContract not to be tagged as submit
func (sc *ManufactureSmartContract) GetEvaluateTransactions() []string {
	return []string{"GetManufacture", "GetManufactureRepairingLotes", "GetManufactureProducingLotes", "GetManufactureUsingLotes", "GetManufactureHistory"}
}
