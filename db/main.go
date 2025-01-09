package db

import (
	"context"
	"dolphin_bot/llm"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var mongoClient *mongo.Client

func ConnectMongoDB(uri string) error {
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}

	mongoClient = client
	return nil
}

func SaveMessage(chatId int64, message *llm.Message) error {
	collection := mongoClient.Database("telegram").Collection("chat_history")

	filter := bson.M{"chat_id": chatId}
	update := bson.M{"$push": bson.M{"messages": message}}

	_, err := collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

func GetHistory(chatId int64) ([]llm.Message, error) {
	collection := mongoClient.Database("telegram").Collection("chat_history")

	filter := bson.M{"chat_id": chatId}

	var result ChatHistory
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			prompt, err := GetPrompt(chatId)
			if err != nil {
				log.Printf("Cannot get prompt: %v", err)
			}
			return []llm.Message{
				{
					Role:    "system",
					Content: prompt,
				},
			}, nil
		}
		return nil, err
	}
	return result.Messages, nil
}

func ClearHistory(chatId int64) error {
	collection := mongoClient.Database("telegram").Collection("chat_history")
	filter := bson.M{"chat_id": chatId}
	_, err := collection.DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}

func SavePrompt(chatId int64, prompt string) error {
	collection := mongoClient.Database("telegram").Collection("prompt")
	filter := bson.M{"chat_id": chatId}
	update := bson.M{"$set": bson.M{"prompt": prompt}}
	_, err := collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func GetPrompt(chatId int64) (string, error) {
	collection := mongoClient.Database("telegram").Collection("prompt")
	filter := bson.M{"chat_id": chatId}
	var result Prompt
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "говори по русски", nil
		}
		return "", err
	}
	return result.Prompt, nil
}

func ChangeModel(chatId int64) (string, error) {
	collection := mongoClient.Database("telegram").Collection("model")
	filter := bson.M{"chat_id": chatId}
	var model Model
	err := collection.FindOne(context.Background(), filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			model.Model = "tinyllama"
		} else {
			return "", err
		}
	}
	switch model.Model {
	case "gemma2:2b":
		model.Model = "tinyllama"
	case "tinyllama":
		model.Model = "gemma2:2b"
	default:
		model.Model = "tinyllama"
	}
	update := bson.M{"$set": bson.M{"model": model.Model}}
	_, err = collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return "", err
	}
	return model.Model, nil
}

func GetModel(chatId int64) (string, error) {
	collection := mongoClient.Database("telegram").Collection("model")
	filter := bson.M{"chat_id": chatId}
	var model Model
	err := collection.FindOne(context.Background(), filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "tinyllama", nil
		}
		return "", err
	}
	return model.Model, nil
}
