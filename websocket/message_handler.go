package websocket

import (
	"encoding/json"
	"github.com/pkg/errors"
	"gitlab.com/fleet-commander/fleet-commander-backend-go/couchbase"
	"gitlab.com/fleet-commander/fleet-commander-backend-go/game"
	"log"
	"strings"
)

func handleMessages(c *connectedPlayer) {
	for {
		message, err := c.NextMessage()
		if err != nil {
			log.Printf("ERROR: %+v", err)
			c.connection.Close()
			break
		}

		switch message.Type {
		case SignInType:
			signIn(message.Payload.(*json.RawMessage), c)

		case SignUpType:
			signUp(message.Payload.(*json.RawMessage), c)
		}
	}
}

func signIn(payload *json.RawMessage, c *connectedPlayer) {
	player := new(game.Player)
	if err := json.Unmarshal(*payload, player); err != nil {
		log.Printf("ERROR: %+v", err)
		c.SendTechnicalErrorMessage()
		return
	}

	playerFromDb, err := couchbase.GetPlayerByEmail(player.Email)
	if err != nil {
		log.Printf("ERROR: %+v", err)
		c.SendTechnicalErrorMessage()
		return
	}

	hash, err := player.PasswordHash()
	if err != nil {
		log.Printf("ERROR: %+v", err)
		c.SendTechnicalErrorMessage()
		return
	}

	if hash != playerFromDb.Password {
		c.SendMessage(NewErrorMessage("Invalid credentials"))
		return
	}

	c.SendMessage(NewSignInMessage())
	c.SendMessage(NewCorrectionMessage(playerFromDb.ActualResources()))
}

func signUp(payload *json.RawMessage, c *connectedPlayer) {
	player := game.NewPlayer()
	if err := json.Unmarshal(*payload, &player); err != nil {
		log.Printf("ERROR: %+v", err)
		c.SendTechnicalErrorMessage()
		return
	}

	exists, err := playerAlreadyExists(player)
	if err != nil {
		log.Printf("ERROR: %+v", err)
		c.SendTechnicalErrorMessage()
		return
	}

	if exists {
		c.SendMessage(NewErrorMessage("Player already exists"))
		return
	}

	err = couchbase.InsertNewPlayer(player)
	if err != nil {
		log.Printf("ERROR: %+v", err)
		c.SendTechnicalErrorMessage()
		return
	}

	c.SendMessage(NewSignUpMessage())
}

func playerAlreadyExists(p game.Player) (bool, error) {
	playerFromDb, err := couchbase.GetPlayerByEmail(p.Email)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return playerFromDb != nil && strings.ToLower(playerFromDb.Email) == strings.ToLower(p.Email) &&
		strings.ToLower(playerFromDb.Name) == strings.ToLower(p.Name), nil
}
