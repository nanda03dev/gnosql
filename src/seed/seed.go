package seed

import (
	"fmt"
	"gnosql/src/in_memory_database"
	"math/rand"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SeedData(ginRouter *gin.Engine, gnoSQL *in_memory_database.GnoSQL) *in_memory_database.Database {
	testDBName := "test"

	if dbExists := gnoSQL.GetDatabase(testDBName); dbExists != nil {
		fmt.Printf("\nSeed %s database already exists\n", testDBName)
		return nil
	}

	db := gnoSQL.CreateDatabase(testDBName)

	UserCollection := in_memory_database.CollectionInput{
		CollectionName: "users",
		IndexKeys:      []string{"city", "pincode"},
	}

	OrderCollection := in_memory_database.CollectionInput{
		CollectionName: "orders",
		IndexKeys:      []string{"userId", "category"},
	}

	collections := []in_memory_database.CollectionInput{UserCollection, OrderCollection}

	db.CreateCollections(collections)

	type City map[string]interface{}
	type Pincode map[string]int

	cities := []City{
		{
			"cityName": "Chennai",
			"pincodeDetails": Pincode{
				"pincodeStart": 600000,
				"pincodeEnd":   600010,
			},
		},
		{
			"cityName": "Bangalore",
			"pincodeDetails": Pincode{
				"pincodeStart": 500000,
				"pincodeEnd":   500010,
			},
		},
		{
			"cityName": "Noida",
			"pincodeDetails": Pincode{
				"pincodeStart": 110025,
				"pincodeEnd":   110035,
			},
		},
	}

	// List of departments
	category := []string{"Food", "Grocery", "Decoration"}

	// Initialize the array with unique usernames and passwords
	for i := 0; i < 100000; i++ {
		user := make(in_memory_database.Document)
		user["userName"] = fmt.Sprintf("user%d", i+1)
		user["pwd"] = fmt.Sprintf("password%d", i+1)

		var city City = cities[rand.Intn(len(cities))]
		var cityDetails Pincode = city["pincodeDetails"].(Pincode)

		user["city"] = city["cityName"]
		pincode := rand.Intn(cityDetails["pincodeEnd"]-cityDetails["pincodeStart"]+1) + cityDetails["pincodeStart"]

		user["pincode"] = strconv.Itoa(pincode)

		userInstance := db.GetCollection(UserCollection.CollectionName)
		userResult := userInstance.Create(user).(in_memory_database.Document)

		userId := userResult["id"]

		orderInstance := db.GetCollection(OrderCollection.CollectionName)
		for i := 0; i < 2; i++ {
			order := make(in_memory_database.Document)
			order["userId"] = userId
			order["category"] = category[rand.Intn(len(category))]
			orderInstance.Create(order)
		}

	}

	return db

}
