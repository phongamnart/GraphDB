package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

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

	app.Get("/get-max-relationship-nodes", func(c *fiber.Ctx) error {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		result, err := session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			result, err := transaction.Run(`
				MATCH (a:Product)-[r]->(b:Option)
				WITH MAX(r.value) AS max_value
				MATCH (n:Product)-[r]->(m:Option)
				WHERE r.value = max_value
				RETURN n, m
			`, nil)
			if err != nil {
				return nil, err
			}

			var nodes []interface{}
			for result.Next() {
				record := result.Record()
				nodeN := record.GetByIndex(0)
				nodeM := record.GetByIndex(1)
				nodes = append(nodes, map[string]interface{}{
					"nodeN": nodeN,
					"nodeM": nodeM,
				})
			}

			return nodes, nil
		})

		if err != nil {
			return err
		}

		return c.JSON(result)
	})

	log.Fatal(app.Listen(":3000"))
}
