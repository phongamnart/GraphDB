package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Node struct {
	Name string `json:"name"`
}

type Relationship struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Relationship string `json:"relationship"`
}

func main() {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	})

	// Connect to Neo4j
	driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.NoAuth())
	if err != nil {
		log.Fatal(err)
	}
	defer driver.Close()

	// Delete a relationship
	app.Delete("/delete-relation", func(c *fiber.Ctx) error {
		var relation Relationship
		if err := c.BodyParser(&relation); err != nil {
			return err
		}

		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			result, err := transaction.Run(
				"MATCH (from:Node {name: $from})-[r]->(to:Node {name: $to}) DELETE r",
				map[string]interface{}{"from": relation.From, "to": relation.To},
			)
			return result, err
		})

		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// Delete all relationships
	app.Delete("/delete-all-relationships", func(c *fiber.Ctx) error {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			result, err := transaction.Run(
				"MATCH ()-[r]-() DELETE r",
				nil,
			)
			return result, err
		})

		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	log.Fatal(app.Listen(":3000"))

}
