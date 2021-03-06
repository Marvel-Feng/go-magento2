package main

import (
	"log"

	"github.com/hermsi1337/go-magento2/pkg/magento2/api"
	"github.com/hermsi1337/go-magento2/pkg/magento2/products/attributeset"
)

func main() {
	// initiate storeconfig
	storeConfig := &api.StoreConfig{
		Scheme:    "https",
		HostName:  "magento2.hermsi.localhost",
		StoreCode: "default",
	}
	// initiate bearer payload
	bearerToken := "yd1o9zs1hb1qxnn8ek68eu8nwqjg5hrv"

	// create a new apiClient
	apiClient, err := api.NewAPIClientFromIntegration(storeConfig, bearerToken)
	if err != nil {
		panic(err)
	}
	log.Printf("Obtained client: '%v'", apiClient)

	// define your atrribute-set
	set := attributeset.AttributeSet{
		AttributeSetName: "foo2",
		SortOrder:        2,
	}

	// "Skeletonid" indicates the creation of the attribute set on the default attribute set that in Magento always has id = 4
	skeletonID := 4

	// create atrribute-set on remote
	mAttributeSet, err := attributeset.CreateAttributeSet(set, skeletonID, apiClient)
	if err != nil {
		panic(err)
	}

	// here you go
	log.Printf("Created attribute-set: %+v", mAttributeSet)
	log.Printf("Detailed attribute-set: %+v", mAttributeSet.AttributeSet)
	log.Printf("Groups of attribute-set: %+v", mAttributeSet.AttributeSetGroups)
}
