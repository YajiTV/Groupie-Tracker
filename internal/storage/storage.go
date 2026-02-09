package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/YajiTV/groupie-tracker/internal/models"
)

var (
	userMutex sync.RWMutex
	UsersFile = "data/users.json"
)

// InitUsers initializes the users.json file
func InitUsers() error {
	// Create data folder
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(UsersFile); os.IsNotExist(err) {
		userData := models.UserData{Users: []models.User{}, LastID: 0}
		return saveJSON(UsersFile, userData)
	}

	return nil
}

// saveJSON saves data to a JSON file
func saveJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// loadJSON loads data from a JSON file
func loadJSON(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(data)
}

// GetAllUsers retrieves all users
func GetAllUsers() ([]models.User, error) {
	userMutex.RLock()
	defer userMutex.RUnlock()

	var userData models.UserData
	if err := loadJSON(UsersFile, &userData); err != nil {
		return nil, err
	}
	return userData.Users, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(id int) (*models.User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, errors.New("utilisateur introuvable")
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*models.User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username == username {
			return &user, nil
		}
	}
	return nil, errors.New("utilisateur introuvable")
}

// CreateUser creates a new user
func CreateUser(user models.User) (*models.User, error) {
	userMutex.Lock()
	defer userMutex.Unlock()

	var userData models.UserData
	if err := loadJSON(UsersFile, &userData); err != nil {
		return nil, err
	}

	// Check if username already exists
	for _, u := range userData.Users {
		if u.Username == user.Username {
			return nil, errors.New("nom d'utilisateur déjà pris")
		}
		if u.Email == user.Email {
			return nil, errors.New("email déjà utilisé")
		}
	}

	// Create new user
	userData.LastID++
	user.ID = userData.LastID
	userData.Users = append(userData.Users, user)

	if err := saveJSON(UsersFile, userData); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates a user
func UpdateUser(user models.User) error {
	userMutex.Lock()
	defer userMutex.Unlock()

	var userData models.UserData
	if err := loadJSON(UsersFile, &userData); err != nil {
		return err
	}

	for i, u := range userData.Users {
		if u.ID == user.ID {
			userData.Users[i] = user
			return saveJSON(UsersFile, userData)
		}
	}

	return errors.New("utilisateur introuvable")
}
