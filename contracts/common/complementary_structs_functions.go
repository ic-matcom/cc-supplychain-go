package main

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Price struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type Environment struct {
	Temperature            string `json:"temperature"`
	TemperatureMeasurement string `json:"temperaturemeasurement"`
	Humidity               string `json:"humidity"`
	//Other details like lighting, preassure, etc.
}

type Brand struct {
	Name string `json:"brandname"`
	Logo string `json:"brandlogo"`
}

type NetContent struct {
	Weight             string `json:"weight"`
	WeightMeasurement  string `json:"weightmeasurement"`
	Volumen            string `json:"volumen"`
	VolumenMeasurement string `json:"volumenmeasurement"`
	Units              string `json:"units"`
}

type Dimention struct {
	Width       string `json:"width"`
	Height      string `json:"height"`
	Deep        string `json:"deep"`
	Measurement string `json:"measurement"`
}

type Location struct {
	Country   string `json:"country"`
	Province  string `json:"province"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type Capacity struct {
	Units           string `json:"units"`
	UnitMeasurement string `json:"unitmeasurement"`
	Used            string `json:"used"`
}

type Transporting struct {
	Destinity  Location `json:"destinity"`
	FinishTime string   `json:"finishtime"`
}

type Certification struct {
	Issuer            string `json:"issuer"`
	CertificationType string `json:"certificationtype"`
	Result            string `json:"result"`
}

// GetSubmittingClientIdentity returns the name and issuer of the identity that
// invokes the smart contract. This function base64 decodes the identity string
// before returning the value to the client or smart contract.
func GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil
}

// GetSubmittingClientOrg returns the id of the org of the identity that
// invokes the smart contract. This function base64 decodes the identity string
// before returning the value to the client or smart contract.
func GetSubmittingClientOrg(ctx contractapi.TransactionContextInterface) (string, error) {
	//CAMBIOS AQUI
	/*
		b64OrgID, err := ctx.GetClientIdentity().GetMSPID()
		if err != nil {
			return "", fmt.Errorf("failed getting client's orgID: %v", err)
		}
		decodeID, err := base64.StdEncoding.DecodeString(b64OrgID)
		if err != nil {
			return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
		}
		return string(decodeID), nil
	*/
	return "", nil
}

// verifyClientOrgMatchesPeerOrg checks that the client is from the same org as the peer
func VerifyClientOrgMatchesPeerOrg(clientOrgID string) error {
	//CAMBIOS AQUI
	/*
		peerb64OrgID, err := shim.GetMSPID()
		if err != nil {
			return fmt.Errorf("failed getting peer's orgID: %v", err)
		}
		peerOrgID, err := base64.StdEncoding.DecodeString(peerb64OrgID)
		if err != nil {
			return fmt.Errorf("failed to base64 decode orgID: %v", err)
		}
		if clientOrgID != string(peerOrgID) {
			return fmt.Errorf("client from org %s is not authorized to read or write private data from an org %s peer",
				clientOrgID,
				string(peerOrgID),
			)
		}
	*/
	return nil
}

func ValidRole(cxt contractapi.TransactionContextInterface, role string) error {
	//CAMBIOS AQUI
	/*
		// Demonstrate the use of Attribute-Based Access Control (ABAC) by checking
		// to see if the caller has the role attribute with a value of true;
		// if not, return an error.
		err := ctx.GetClientIdentity().AssertAttributeValue(role, "true")
		if err != nil {
			return fmt.Errorf("submitting client not authorized to create asset, does not have abac.manufacturer role")
		}
	*/
	return nil
}
