package main

import (
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/klyngen/flightlogger/email"

	"github.com/klyngen/flightlogger/repository"
	"github.com/klyngen/flightlogger/service"

	"github.com/klyngen/flightlogger/configuration"
	"github.com/klyngen/flightlogger/presentation"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("##### STARTING FLIGHTLOG BACKEND ####")
	log.Println("Starting flightlog API")

	// ######## BUILD THE SERVICE ##############

	// Get the configuration - WILL PANIC IF FAILS
	config := configuration.GetConfiguration()

	db := &repository.MySQLRepository{}

	// Create the database connection (DataLayer)
	err := db.CreateConnection(config.DatabaseConfiguration.Username,
		config.DatabaseConfiguration.Password,
		config.DatabaseConfiguration.Database,
		config.DatabaseConfiguration.Port,
		config.DatabaseConfiguration.Hostname)

	if err != nil {
		log.Fatalf("Likely a database misconfiguration: %v", err)
	}


	// Should be enough to add email-support to our application (DataLayer)
	emailService := email.NewEmailService(config.EmailConfiguration)

	if emailService == nil {
		panic("Cannot have non-existing email-service")
	}

	var service common.FlightLogService

	if config.RedisConfiguration.IsEmpty() {
		service = service.NewService(db, emailService, config)
	} else {
		redisPool := createRedisPool(config.RedisConfiguration)
		service = service.NewServiceWithPersistedSession(db, emailService, config, redisstore.New(redisPool))
	}

	// Instantiate our use-case / service-layer

	// Create our presentation layer
	api := presentation.NewService(service, config)
	api.StartAPI()

}

func createRedisPool(config configuration.DatabaseConfig) {
	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp", 
				fmt.Sprintf("%s:%s", config.Hostname, config.Port), 
				redis.DialPassword(config.Password),
				redis.DialClientName(config.Username),
			))
		},
	}
}