package config

import(
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	DATABASE *mongo.Database
	CLIENT 	 *mongo.Client
)

const DATABASE_URL = "mongodb+srv://admin:W6ptbj7HPS3RJ4cU@cluster0.tgypip5.mongodb.net/"
const DATABASE_NAME = "Zygo"