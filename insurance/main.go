package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Node struct {
	Name    string `json:"name"`
	NewName string `json:"newName"`
}

type Relationship struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Relationship string `json:"relationship"`
	Attribute    string `json:"attribute"`
	NewValue     string `json:"newValue"`
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

	app.Get("/get-all-data", func(c *fiber.Ctx) error {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		result, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			nodesResult, err := transaction.Run(
				"MATCH (n) RETURN n",
				nil,
			)
			if err != nil {
				return nil, err
			}

			relationshipsResult, err := transaction.Run(
				"MATCH ()-[r]->() RETURN r",
				nil,
			)
			if err != nil {
				return nil, err
			}

			var nodes []interface{}
			for nodesResult.Next() {
				record := nodesResult.Record()
				node := record.GetByIndex(0)
				nodes = append(nodes, node)
			}

			var relationships []interface{}
			for relationshipsResult.Next() {
				record := relationshipsResult.Record()
				relationship := record.GetByIndex(0)
				relationships = append(relationships, relationship)
			}

			data := map[string]interface{}{
				"nodes":         nodes,
				"relationships": relationships,
			}

			return data, nil
		})

		if err != nil {
			return err
		}

		return c.JSON(result)
	})

	log.Fatal(app.Listen(":3000"))
}