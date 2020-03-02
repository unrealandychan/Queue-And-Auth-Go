package controllers

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
		"log"
	"fmt"
	"os"
	"../models"
	uuid2 "github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"strconv"
	"github.com/dgrijalva/jwt-go"
	"time"
	"github.com/pkg/errors"
)

type DynamodbDriver struct{}


// Create a Table with All Users/Log records.
// HASH Key would be UUID for every individual event
// Other Attribute of Table :
//    - JWT
//    - CreationTime
//    - Access Time
func (d *DynamodbDriver) CreateUserTable()(interface{},error){
	sess := session.Must(session.NewSession(&aws.Config{
		Region:aws.String("ap-southeast-1"),
	}))
	tableName :=  os.Getenv("TABLE_NAME")
	svc := dynamodb.New(sess)

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions:[]*dynamodb.AttributeDefinition{

			{
				AttributeName:aws.String("Ticket"),
				AttributeType:aws.String("N"),
			},

			{
				AttributeName:aws.String("UUID"),
				AttributeType:aws.String("B"),
			},

		},KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("UUID"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("Ticket"),
				KeyType:       aws.String("RANGE"),
			},

		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableName),
	}

	_ , err := svc.CreateTable(input)

	if err!=nil{
		fmt.Println(err)
		return nil,err
	}
	fmt.Printf("Table name %s have been created \n",tableName)
	return true,err
}


func (d *DynamodbDriver) CreateCountTable()(interface{},error){
	sess := session.Must(session.NewSession(&aws.Config{
		Region:aws.String("ap-southeast-1"),
	}))
	tableName :=  "EventCount"
	svc := dynamodb.New(sess)

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions:[]*dynamodb.AttributeDefinition{

			{
				AttributeName:aws.String("Event"),
				AttributeType:aws.String("S"),
			},

		},KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Event"),
				KeyType:       aws.String("HASH"),
			},


		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableName),
	}

	_ , err := svc.CreateTable(input)

	if err!=nil{
		fmt.Println(err)
		return nil ,err
	}
	fmt.Printf("Table name %s have been created \n",tableName)
	return true,err
}

func (d *DynamodbDriver) NewEvent(event string)(interface{},error){
	sess := session.Must(session.NewSession(&aws.Config{
		Region:aws.String("ap-southeast-1"),
	}))
	tableName :=  "EventCount"
	svc := dynamodb.New(sess)

	Event := models.Event{
		event,
		0,
		0,
	}

	av, err := dynamodbattribute.MarshalMap(Event)
	if err != nil{
		log.Fatal(err)
	}

	input := &dynamodb.PutItemInput{
		Item:av,
		TableName:aws.String(tableName),
	}
	_ , err = svc.PutItem(input)
	if err != nil{
		return nil , err
	}
	fmt.Printf("Event name %s has been Create \n",event)
	return true, err
}

func (d *DynamodbDriver) GetToken(event string)(interface{},error){
	sess := session.Must(session.NewSession(&aws.Config{
		Region:aws.String("ap-southeast-1"),
	}))
	tableName :=  os.Getenv("TABLE_NAME")
	svc := dynamodb.New(sess)

	input:=&dynamodb.UpdateItemInput{
		TableName:aws.String("EventCount"),
		Key: map[string]*dynamodb.AttributeValue{
			"Event":{S:aws.String(event)},
		},
		ExpressionAttributeNames:map[string]*string{
			"#Count":aws.String("QueueCount"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":q":{
				N:aws.String("1"),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
		UpdateExpression:aws.String("SET #Count = #Count +:q" ),
	}


	result , err := svc.UpdateItem(input)

	countNumber ,err := strconv.Atoi(*result.Attributes["QueueCount"].N)

	uuid, err := uuid2.NewUUID()
	if err !=nil{
		log.Fatal(err)
	}

	Item := models.Ticket{
		uuid,
		event,
		countNumber,
		false,
		time.Now().Unix(),
	}


	av, err := dynamodbattribute.MarshalMap(Item)
	if err != nil{
		return nil,err
	}

	item_input := &dynamodb.PutItemInput{
		Item:av,
		TableName:aws.String(tableName),
	}
	_ , err = svc.PutItem(item_input)
	if err!=nil{
		return nil,err
	}
	fmt.Printf("User with Ticker %v have been issue \n",countNumber)


	claims := &models.Claims{
		uuid,
		event,
		countNumber,
		false,
		jwt.StandardClaims{
			IssuedAt:time.Now().Unix(),
			Issuer:"Oneshop",
			ExpiresAt:time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)

	tokenString ,err := token.SignedString([]byte(os.Getenv("TOKEN")))

	if err != nil{
		return nil , err
	}


	return tokenString ,err
	}

func (D *DynamodbDriver) ValidateToken(tokenString string)(*models.Claims,error){
	JWT_TOKEN := []byte(os.Getenv("TOKEN"))
	claims := &models.Claims{}

	token , err := jwt.ParseWithClaims(tokenString,claims,func(token *jwt.Token)(interface{},error){
		return JWT_TOKEN, nil
	})

	if err !=nil{
		return claims  , err
	}
	if !token.Valid{
		fmt.Println("Token Invalid")
	return nil , err
	}
	return claims,nil
}

func (D *DynamodbDriver)CheckQueue(claims *models.Claims )(interface{},error){

	sess := session.Must(session.NewSession(&aws.Config{
		Region:aws.String("ap-southeast-1"),
	}))
	tableName :=  "EventCount"
	svc := dynamodb.New(sess)

	input := & dynamodb.GetItemInput{

		Key: map[string]*dynamodb.AttributeValue{
			"Event":{S:aws.String(claims.Event),},
		},

		TableName:aws.String(tableName),
	}

	result , err := svc.GetItem(input)
	item :=models.Event{}
	err = dynamodbattribute.UnmarshalMap(result.Item,&item)

	if err !=nil{
		return nil, err
	}

	if claims.Ticket > item.CurrentCount{
		//Success and Return a JWT Token with claim == true
		claims.Access=true
		claims.ExpiresAt =time.Now().Add(time.Minute * 15).Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)


		input := &dynamodb.UpdateItemInput{
			TableName:aws.String(claims.Event),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":u":{
					N:aws.String(strconv.Itoa(int(time.Now().Unix()))),
				},

			},
			Key: map[string]*dynamodb.AttributeValue{
				"UUID":{
					B:claims.UUID[:],
				},

				"Ticket":{
					N:aws.String(strconv.Itoa(claims.Ticket)),
				},
			},
			UpdateExpression:aws.String("set UpdateTime = :u"),
		}
		_ ,err := svc.UpdateItem(input)

		if err != nil{
			return nil ,err
		}

		tokenString ,err := token.SignedString([]byte(os.Getenv("TOKEN")))
		if err != nil{
			return nil ,err
		}

		return tokenString,nil

	}else{
		//Fail , Have to Wait Again
		return nil,err
	}

}

func (D *DynamodbDriver)CheckAuth(claims *models.Claims)(interface{},error){
	if claims.Access == true{
		return true ,nil
	}else{
		return false, errors.New("Auth Fail")
	}
}