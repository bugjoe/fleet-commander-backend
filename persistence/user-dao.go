package persistence

import (
	"errors"
	"fleet-commander-backend-go/models"
	"fmt"
)

// InsertNewUser inserts a new user. When either the Username or the Email of the
// user already exists, the functions returns an error
func InsertNewUser(user *models.User) error {
	passwordHash := user.GetPasswordHash()
	if len(passwordHash) == 0 {
		fmt.Println("ERROR: Invalid password hash", passwordHash)
		return errors.New("")
	}

	user.Password = passwordHash
	fmt.Println("Insert new user:", user)

	database := newArangoDatabase()
	query := "FOR u IN users FILTER LOWER(u.Email) == LOWER(@email) AND LOWER(u.Username) == LOWER(@username) RETURN u"
	bindings := bindingVariables{
		"email":    user.Email,
		"username": user.Username,
	}

	cursor, err := database.Query(nil, query, bindings)
	if err != nil {
		fmt.Println("ERROR: invalid query", err)
		return err
	}

	if cursor.HasMore() {
		fmt.Println("WARN: user already exists")
		return fmt.Errorf("User with username=%s and email=%s already exists", user.Username, user.Email)
	}

	collection, err := database.Collection(nil, "users")
	if err != nil {
		fmt.Println("ERROR: can't open collection", err)
		return err
	}

	_, err = collection.CreateDocument(nil, user)
	if err != nil {
		fmt.Println("ERROR: can't create user", err)
		return err
	}

	return nil
}

// GetUserByEmail returns the user that matches with the given email
func GetUserByEmail(email string) (*models.User, error) {
	fmt.Println("Get user by email:", email)

	database := newArangoDatabase()
	query := "FOR u IN users FILTER LOWER(u.Email) == LOWER(@email) RETURN u"
	bindings := bindingVariables{
		"email": email,
	}
	cursor, err := database.Query(nil, query, bindings)
	if err != nil {
		fmt.Println("ERROR: no user found with email:", email, err)
		return nil, err
	}

	if cursor.Count() > 1 {
		fmt.Printf("WARN: found (%d) users, will use first one", cursor.Count())
	}

	user := new(models.User)
	_, err = cursor.ReadDocument(nil, user)
	if err != nil {
		fmt.Println("ERROR: can't read user from cursor", err)
		return nil, err
	}

	return user, nil
}
