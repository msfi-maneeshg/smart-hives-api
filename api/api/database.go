package api

import (
	"context"
	"fmt"
	"log"
	"smart-hives/api/common"
	"smart-hives/api/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//GetFarmerProfile :
func GetFarmerProfile(emailID, username string) (objProfile FarmerProfileDetails) {
	collection := database.Data.Collection(common.PROFILES)
	filterData := bson.D{
		{"email", emailID},
	}
	if username != "" {
		filterData = bson.D{
			{"username", username},
		}
	}

	err := collection.FindOne(context.TODO(), filterData).Decode(&objProfile)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Fatal(err)
	}

	return objProfile
}

//CreateNewProfile :
func CreateNewProfile(objProfile FarmerProfileDetails) (err error) {
	collection := database.Data.Collection(common.PROFILES)

	_, err = collection.InsertOne(context.TODO(), objProfile)
	if err != nil {
		return err
	}
	return nil
}

//UpdateUserPassword :
func UpdateUserPassword(objProfile FarmerProfileDetails) (err error) {
	collection := database.Data.Collection(common.PROFILES)

	opts := options.Update().SetUpsert(false)
	filter := bson.D{{"username", objProfile.Username}, {"email", objProfile.Email}}
	update := bson.D{{"$set", bson.D{{"password", objProfile.Password}}}}

	result, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("%v", "username/email not found")
	}
	return nil
}
