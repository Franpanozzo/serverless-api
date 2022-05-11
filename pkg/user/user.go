package user

import (
	"encoding/json"
	"errors"
	"serverless-api/pkg/validators"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

//Aca estarian los models si hablamos de MVC

var (
	ErrorFailedToFetchRecord     = "failed to fetch error"
	ErrorFailedToUnmarshalRecord = "failed to unmarhsal error"
	ErrorInvalidUserData         = "invalid user data"
	ErrorInvalidEmail            = "invalid email"
	ErrorCouldNotMarshallItem    = "could not marshal item"
	ErrorCouldNotDeleteItem      = "could not delete item"
	ErrorCouldNotDynamoPutItem   = "could not dynamo put iden"
	ErrorUserAlreadyExists       = "user.User already exists"
	ErrorUserDoesNotExists       = "user.User does no exist"
)

type User struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func FetchUser(email, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tableName),
	}

	result, err := dynaClient.GetItem(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	item := new(User)                                       // Lo mismo que item := &User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, item) // Pasamos a structs un arreglo de jsons
	if err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}

	return item, nil
}

func FetchUsers(tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*[]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := dynaClient.Scan(input) //Scan es como un getAll
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	item := new([]User)
	if err = dynamodbattribute.UnmarshalListOfMaps(result.Items, item); err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}

	return item, nil
}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (
	*User, error) {
	var u User // Aca vamos a hacer el Unmarshal

	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}
	if !validators.IsEmailValid(u.Email) {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currentUser, _ := FetchUser(u.Email, tableName, dynaClient)
	if currentUser != nil && len(currentUser.Email) != 0 {
		return nil, errors.New(ErrorUserAlreadyExists)
	}

	av, err := dynamodbattribute.MarshalMap(u) // marshalMap para hacer un struct a json, LO PASAMOS a json porque eso entiendo dynamodb
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshallItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	if _, err := dynaClient.PutItem(input); err != nil { // Queda mas clean asi el chequeo del error
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}

	return &u, nil
}

func UpdateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (
	*User, error) {
	var u User
	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currentUser, _ := FetchUser(u.Email, tableName, dynaClient)
	if currentUser != nil && len(currentUser.Email) == 0 {
		return nil, errors.New(ErrorUserDoesNotExists)
	}

	av, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshallItem)
	}

	input := &dynamodb.PutItemInput{ //Lo estarias pisando
		Item:      av,
		TableName: aws.String(tableName),
	}

	if _, err := dynaClient.PutItem(input); err != nil {
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}

	return &u, nil
}

func DeleteUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) error {
	email := req.QueryStringParameters["email"]
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tableName),
	}

	if _, err := dynaClient.DeleteItem(input); err != nil {
		return errors.New(ErrorCouldNotDeleteItem)
	}

	return nil
}
