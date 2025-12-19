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

// InitUsers initialise le fichier users.json
func InitUsers() error {
	// Créer le dossier data
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	// Vérifier si le fichier existe
	if _, err := os.Stat(UsersFile); os.IsNotExist(err) {
		userData := models.UserData{Users: []models.User{}, LastID: 0}
		return saveJSON(UsersFile, userData)
	}

	return nil
}

// saveJSON sauvegarde des données dans un fichier JSON
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

// loadJSON charge des données depuis un fichier JSON
func loadJSON(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(data)
}

// GetAllUsers récupère tous les utilisateurs
func GetAllUsers() ([]models.User, error) {
	userMutex.RLock()
	defer userMutex.RUnlock()

	var userData models.UserData
	if err := loadJSON(UsersFile, &userData); err != nil {
		return nil, err
	}
	return userData.Users, nil
}

// GetUserByID récupère un utilisateur par son ID
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

// GetUserByUsername récupère un utilisateur par son username
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

// CreateUser crée un nouvel utilisateur
func CreateUser(user models.User) (*models.User, error) {
	userMutex.Lock()
	defer userMutex.Unlock()

	var userData models.UserData
	if err := loadJSON(UsersFile, &userData); err != nil {
		return nil, err
	}

	// Vérifier si le username existe déjà
	for _, u := range userData.Users {
		if u.Username == user.Username {
			return nil, errors.New("nom d'utilisateur déjà pris")
		}
		if u.Email == user.Email {
			return nil, errors.New("email déjà utilisé")
		}
	}

	// Créer le nouvel utilisateur
	userData.LastID++
	user.ID = userData.LastID
	userData.Users = append(userData.Users, user)

	if err := saveJSON(UsersFile, userData); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser met à jour un utilisateur
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
