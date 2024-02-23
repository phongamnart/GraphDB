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
	From            string `json:"from"`
	To              string `json:"to"`
	Relationship    string `json:"relationship"`
	Attribute		string `json:"attribute"`
	NewValue		string `json:"newValue"`
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

	// Delete a node
	app.Delete("/delete-node", func(c *fiber.Ctx) error {
		var node Node
		if err := c.BodyParser(&node); err != nil {
			return err
		}

		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			result, err := transaction.Run(
				"MATCH (n:Node {name: $name}) DELETE n",
				map[string]interface{}{"name": node.Name},
			)
			return result, err
		})

		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// Delete all nodes
	app.Delete("/delete-node-all", func(c *fiber.Ctx) error {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			result, err := transaction.Run(
				"MATCH (n) DETACH DELETE n",
				nil,
			)
			return result, err
		})

		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// Update node name
	app.Put("/update-node", func(c *fiber.Ctx) error {
		var node Node
		if err := c.BodyParser(&node); err != nil {
			return err
		}

		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			result, err := transaction.Run(
				"MATCH (n:Node {name: $oldName}) SET n.name = $newName RETURN n",
				map[string]interface{}{"oldName": node.Name, "newName": node.NewName},
			)
			return result, err
		})

		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// Create a relationship
	app.Post("/create-relation", func(c *fiber.Ctx) error {

		var rel Relationship
		if err := c.BodyParser(&rel); err != nil {
			return err
		}

		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			_, err := transaction.Run(
				"MATCH (from:Node {name: $from}), (to:Node {name: $to}) CREATE (from)-[:`"+rel.Relationship+"`]->(to)",
				map[string]interface{}{"from": rel.From, "to": rel.To},
			)
			return nil, err
		})

		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	// Delete a relationship
	app.Delete("/delete-relation", func(c *fiber.Ctx) error {
		var rel Relationship
		if err := c.BodyParser(&rel); err != nil {
			return err
		}

		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			result, err := transaction.Run(
				"MATCH (from:Node {name: $from})-[r]->(to:Node {name: $to}) DELETE r",
				map[string]interface{}{"from": rel.From, "to": rel.To},
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

	app.Put("/update-relationship", func(c *fiber.Ctx) error {
		var rel Relationship
		if err := c.BodyParser(&rel); err != nil {
			return err
		}
	
		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()
	
		_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			result, err := transaction.Run(
				"MATCH (from:Node {name: $from})-[r:`" + rel.Relationship + "`]->(to:Node {name: $to}) SET r." + rel.Attribute + " = $newValue RETURN r",
				map[string]interface{}{"from": rel.From, "to": rel.To, "newValue": rel.NewValue},
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
