package seed

import (
	"fmt"
	"gnosql/src/in_memory_database"
	"math/rand"
	"strconv"
)

func SeedData(gnoSQL *in_memory_database.GnoSQL) *in_memory_database.Database {
	testDBName := "test"

	UserCollectionInput := in_memory_database.CollectionInput{
		CollectionName: "users",
		IndexKeys:      []string{"city", "pincode"},
	}

	OrderCollectionInput := in_memory_database.CollectionInput{
		CollectionName: "orders",
		IndexKeys:      []string{"userId", "category"},
	}

	collectionsInput := []in_memory_database.CollectionInput{UserCollectionInput, OrderCollectionInput}

	if dbExists := gnoSQL.GetDB(testDBName); dbExists != nil {
		fmt.Printf("\nSeed %s database already exists\n", testDBName)
		return nil
	}

	db := gnoSQL.CreateDB(testDBName, collectionsInput)

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
	for i := 0; i < 100; i++ {
		user := make(in_memory_database.Document)
		user["userName"] = fmt.Sprintf("user%d", i+1)
		user["pwd"] = fmt.Sprintf("password%d", i+1)

		var city City = cities[rand.Intn(len(cities))]
		var cityPincodeDetails Pincode = city["pincodeDetails"].(Pincode)

		user["city"] = city["cityName"]
		pincode := rand.Intn(cityPincodeDetails["pincodeEnd"]-cityPincodeDetails["pincodeStart"]+1) + cityPincodeDetails["pincodeStart"]

		user["pincode"] = strconv.Itoa(pincode)

		UserCollection := db.GetColl(UserCollectionInput.CollectionName)
		newUser := UserCollection.Create(user)

		userId := newUser["id"]

		OrderCollection := db.GetColl(OrderCollectionInput.CollectionName)

		for i := 0; i < 2; i++ {
			order := make(in_memory_database.Document)
			order["userId"] = userId
			order["category"] = category[rand.Intn(len(category))]
			OrderCollection.Create(order)
		}

	}
	// manually write seed test database to disk
	go db.SaveDatabaseToFile()

	return db

}
