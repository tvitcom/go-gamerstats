package storage

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)
func NewClient(dsn string) *mongo.Client {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		log.Fatal(err)
	}
	return cl
}

func ShowDbs(cl *mongo.Client) {
	databases, err := cl.ListDatabaseNames(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(databases)
}

// func GetCollection(cl *mongo.Client, dbname, dbtable string) {
// 	col := cl.Database(dbname).Collection(dbtable)
// 	_ = col
// 	fmt.Println(col)
// }