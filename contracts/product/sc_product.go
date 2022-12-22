package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const product_index = "product"

//Contrat for writing and reading product from world-state
type ProductSmartContract struct {
	contractapi.Contract
}

// Create and put new product to world-state
func (sc *ProductSmartContract) CreateProduct(ctx CustomTransactionContextInterface, key string,
	productName string, description string, brandname string, brandlogo string, core string, variety string,
	productImage string, Weight string, WeightMeasurement string, Volumen string, VolumenMeasurement string,
	Units string, PackDimentionWidth string, PackDimentionHeight string, PackDimentionDeep string, PackDimentionMeasurement string,
	DisplaySpaceWidth string, DisplaySpaceHeight string, DisplaySpaceDeep string, DisplaySpaceMeasurement string,
	packageType string, manufactureDetails string, manufacturer string, components []string) error {

	err := ValidRole(ctx, "abac.manufacturer_administrator")

	if err != nil {
		return err
	}

	//Obtain current product with the given key
	//currentProduct := ctx.GetData()
	currentProduct, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if currentProduct != nil {
		return fmt.Errorf("Cannot create new product in world state as key %s already exists", key)
	}

	// Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	//Create a new struct of product
	product := new(Product)

	product.DocType = "product"
	product.ProductID = key
	product.ProductName = productName
	product.Description = description
	product.Brand = Brand{brandname, brandlogo}
	product.Core = core
	product.Variety = variety
	product.ProductImage = productImage
	product.NetContent = NetContent{Weight, WeightMeasurement, Volumen, VolumenMeasurement, Units}
	product.WidthAndHeight = Dimention{PackDimentionWidth, PackDimentionHeight, PackDimentionDeep, PackDimentionMeasurement}
	product.DisplaySpace = Dimention{DisplaySpaceWidth, DisplaySpaceHeight, DisplaySpaceDeep, DisplaySpaceMeasurement}
	product.PackageType = packageType
	product.Certifications = make([]Certification, 0)
	product.ManufactureDetails = manufactureDetails
	product.Manufacturer = manufacturer
	product.Components = components
	product.Advisor = advisor

	product.ProductIsOnTesting()

	//Obtain json encoding
	productJson, _ := json.Marshal(product)

	//put product state in world state
	err = ctx.GetStub().PutState(key, []byte(productJson))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//  This will enable very efficient state range queries based on composite keys matching indexName~productname~variety*
	productIndexKey, err := ctx.GetStub().CreateCompositeKey(product_index, []string{product.ProductID, product.ProductName, product.Variety})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	//put composite key of product in world statte
	err = ctx.GetStub().PutState(productIndexKey, value)

	return nil
}

//If return true then product with the given key exists
func (sc *ProductSmartContract) ExistsProduct(ctx CustomTransactionContextInterface, key string) bool {
	//Obtain current product with the given key
	//currentProduct := ctx.GetData()
	currentProduct, err := ctx.GetStub().GetState(key)

	if err != nil {
		return false
	}

	return currentProduct != nil
}

// Add certification to the product with the givrn key from world-state
func (sc *ProductSmartContract) Certify(ctx CustomTransactionContextInterface, key string,
	issuer string, certificationtype string, result string) error {

	//Obtain current product with the given key
	//currentProduct := ctx.GetData()
	currentProduct, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//Verify if current product exists
	if currentProduct == nil {
		return fmt.Errorf("Cannot update product in world state as key %s does not exist", key)
	}

	product := new(Product)

	err = json.Unmarshal(currentProduct, product)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Product", key)
	}

	//Verify product state

	if product.ProductState != 0 {
		return fmt.Errorf("Product with key %s is not on testing", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != product.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	//Add certification to products certifications

	product.Certifications = append(product.Certifications, Certification{issuer, certificationtype, result})

	//Obtain json encoding
	productBytes, _ := json.Marshal(product)

	//put product state in world state
	err = ctx.GetStub().PutState(key, []byte(productBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Put on production the product with the given key from world-state
func (sc *ProductSmartContract) ProductOnProduction(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current product with the given key
	//currentProduct := ctx.GetData()
	currentProduct, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if currentProduct == nil {
		return fmt.Errorf("Cannot update product in world state as key %s does not exist", key)
	}

	product := new(Product)

	err = json.Unmarshal(currentProduct, product)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Product", key)
	}

	//Verify product state
	if product.ProductState != 0 {
		return fmt.Errorf("Product with key %s is not on testing", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != product.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	product.ProductIsOnProduction()

	//Obtain json encoding
	productBytes, _ := json.Marshal(product)

	//put product state in world state
	err = ctx.GetStub().PutState(key, []byte(productBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Put on discontinued the product with the given key from world-state
func (sc *ProductSmartContract) ProductDiscontinued(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current product with the given key
	//currentProduct := ctx.GetData()
	currentProduct, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if currentProduct == nil {
		return fmt.Errorf("Cannot update product in world state as key %s does not exist", key)
	}

	product := new(Product)

	err = json.Unmarshal(currentProduct, product)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Product", key)
	}

	//Verify product state
	if product.ProductState != 1 {
		return fmt.Errorf("Product with key %s is not on production", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != product.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	product.ProductIsDiscontinued()

	//Obtain json encoding
	productBytes, _ := json.Marshal(product)

	//put product state in world state
	err = ctx.GetStub().PutState(key, []byte(productBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Put on continued the product with the given key from world-state
func (sc *ProductSmartContract) ProductContinued(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current product with the given key
	//currentProduct := ctx.GetData()
	currentProduct, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if currentProduct == nil {
		return fmt.Errorf("Cannot update product in world state as key %s does not exist", key)
	}

	product := new(Product)

	err = json.Unmarshal(currentProduct, product)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Product", key)
	}

	//Verify product state
	if product.ProductState != 2 {
		return fmt.Errorf("Product with key %s is not discontinued", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != product.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	product.ProductIsOnProduction()

	//Obtain json encoding
	productBytes, _ := json.Marshal(product)

	//put product state in world state
	err = ctx.GetStub().PutState(key, []byte(productBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Put on destroyed the product with the given key from world-state
func (sc *ProductSmartContract) ProductDestroyed(ctx CustomTransactionContextInterface, key string) error {

	//Obtain current product with the given key
	//currentProduct := ctx.GetData()
	currentProduct, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if currentProduct == nil {
		return fmt.Errorf("Cannot update product in world state as key %s does not exist", key)
	}

	product := new(Product)

	err = json.Unmarshal(currentProduct, product)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type Product", key)
	}

	//Verify product state
	if product.ProductState != 2 {
		return fmt.Errorf("Product with key %s is not discontinued", key)
	}

	//Get ID of submitting client identity
	advisor, err := GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if advisor != product.Advisor {
		return fmt.Errorf("submitting client not authorized to update asset, does not advise asset")
	}

	product.ProductIsDiscontinued()

	//Obtain json encoding
	productBytes, _ := json.Marshal(product)

	//put product state in world state
	err = ctx.GetStub().PutState(key, []byte(productBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	//Get index entry
	productIndexKey, err := ctx.GetStub().CreateCompositeKey(product_index, []string{product.ProductID, product.ProductName, product.Variety})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(productIndexKey)
}

// GetProductHistory returns the chain of custody for an product since issuance.
func (sc *ProductSmartContract) GetProductHistory(ctx contractapi.TransactionContextInterface, key string) ([]HistoryQueryResultProduct, error) {
	log.Printf("GetProductHistory: ID %v", key)

	//Get History of the product with the given key
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResultProduct
	//For each record in history takes the record value and add to the history
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product = new(Product)
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, product)
			if err != nil {
				return nil, err
			}
		} else {
			product.ProductID = key
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResultProduct{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    product,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// Get the value of the product with the given key from world-state
func (sc *ProductSmartContract) GetProduct(ctx CustomTransactionContextInterface, key string) (*Product, error) {

	//currentProduct := ctx.GetData()
	currentProduct, err := ctx.GetStub().GetState(key)

	if err != nil {
		return nil, errors.New("Unable to interact with world state")
	}

	if currentProduct == nil {
		return nil, fmt.Errorf("Cannot read product in world state as key %s does not exist", key)
	}

	product := new(Product)

	err = json.Unmarshal(currentProduct, product)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Product", key)
	}

	return product, nil
}

// Get the value of the product with the given key from world-state
func (sc *ProductSmartContract) GetProductAdvisor(ctx CustomTransactionContextInterface, key string) (string, error) {

	//currentProduct := ctx.GetData()
	currentProduct, err := ctx.GetStub().GetState(key)

	if err != nil {
		return "", errors.New("Unable to interact with world state")
	}

	if currentProduct == nil {
		return "", fmt.Errorf("Cannot read product in world state as key %s does not exist", key)
	}

	product := new(Product)

	err = json.Unmarshal(currentProduct, product)

	if err != nil {
		return "", fmt.Errorf("Data retrieved from world state for key %s was not of type Product", key)
	}

	return product.Advisor, nil
}

// GetEvaluateTransactions returns functions of SimpleContract not to be tagged as submit
func (sc *ProductSmartContract) GetEvaluateTransactions() []string {
	return []string{"GetProduct", "GetProductAdvisor", "GetProductHistory"}
}
