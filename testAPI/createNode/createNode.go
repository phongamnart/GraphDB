package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Node struct {
	Name string `json:"name"`
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

	// Create a new node
	app.Post("/create-node", func(c *fiber.Ctx) error {
		var node Node
		if err := c.BodyParser(&node); err != nil {
			return err
		}

		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			result, err := transaction.Run(
				"CREATE (n:Node {name: $name}) RETURN n",
				map[string]interface{}{"name": node.Name},
			)
			if err != nil {
				return nil, err
			}

			if result.Next() {
				record := result.Record()
				node := record.GetByIndex(0)
				return node, nil
			}

			return nil, result.Err()
		})

		if err != nil {
			return err
		}

		return c.JSON(node)
	})

	log.Fatal(app.Listen(":3000"))
}