package api

import (
	"context"
	"log"
	"smart-hives/api/common"
	"smart-hives/api/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetFarmerProfile(emailID string) (objProfile FarmerProfileDetails) {
	collection := database.Data.Collection(common.PROFILES)
	filterData := bson.D{
		{"email", emailID},
	}

	err := collection.FindOne(context.TODO(), filterData).Decode(&objProfile)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Fatal(err)
	}

	return objProfile
}

func CreateNewProfile(objProfile FarmerProfileDetails) (err error) {
	collection := database.Data.Collection(common.PROFILES)

	_, err = collection.InsertOne(context.TODO(), objProfile)
	if err != nil {
		return err
	}
	return nil
}
