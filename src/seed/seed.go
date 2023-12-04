package seed

import (
	"fmt"
	"gnosql/src/in_memory_database"
	"gnosql/src/router"
	"math/rand"

	"github.com/gin-gonic/gin"
)

func SeedData(ginRouter *gin.Engine, db *in_memory_database.Database) {
	UserCollection := in_memory_database.CollectionInput{
		CollectionName: "users",
		IndexKeys:      []string{"city"},
	}

	OrderCollection := in_memory_database.CollectionInput{
		CollectionName: "orders",
		IndexKeys:      []string{"userId", "category"},
	}

	collections := []in_memory_database.CollectionInput{UserCollection, OrderCollection}

	addedCollectionInstance := db.AddCollections(collections)

	router.GenerateEntityRoutes(ginRouter, addedCollectionInstance)

	// List of departments
	cities := []string{"Chennai", "Banglore", "Noida"}
	category := []string{"Food", "Grocery", "Decoration"}

	// Initialize the array with unique usernames and passwords
	for i := 0; i < 10000; i++ {
		user := make(in_memory_database.DocumentInput)
		user["userName"] = fmt.Sprintf("user%d", i+1)
		user["pwd"] = fmt.Sprintf("password%d", i+1)
		user["city"] = cities[rand.Intn(len(cities))]

		userInstance, _ := db.GetCollection(UserCollection.CollectionName)
		userResult := userInstance.Create(user).(in_memory_database.DocumentInput)

		userId := userResult["id"]

		orderInstance, _ := db.GetCollection(OrderCollection.CollectionName)
		for i := 0; i < 2; i++ {
			order := make(in_memory_database.DocumentInput)
			order["userId"] = userId
			order["category"] = category[rand.Intn(len(category))]
			orderInstance.Create(order)
		}

	}

}
