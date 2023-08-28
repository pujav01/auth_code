package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// var client *mongo.Client
var userCollection *mongo.Collection

func main() {
	// Load sensitive information from environment variables
	os.Setenv("DB_URI", "mongodb+srv://pujav01:mongodb@cluster0.dgggrej.mongodb.net/?retryWrites=true&w=majority")
	os.Setenv("DB_NAME", "<dbname>")

	dbURI := os.Getenv("DB_URI")
	dbName := os.Getenv("DB_NAME")
	if dbURI == "" || dbName == "" {
		log.Fatal("DB_URI and DB_NAME environment variables are required")
	}

	// Connect to MongoDB
	opts := options.Client().ApplyURI(dbURI)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	userCollection = client.Database(dbName).Collection("users")
	if userCollection == nil {
		log.Fatal("Failed to initialize user collections")
	}

	// HTTP server setup
	http.HandleFunc("/register", Register)

	port := 65005
	fmt.Printf("Server is running on port %d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("unimplemented method"))
		return
	}

	u := new(User)

	err := json.NewDecoder(r.Body).Decode(u)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	err = u.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user := User{
		User_ID:  primitive.NewObjectID(),
		Name:     u.Name,
		Email:    u.Email,
		Mobile:   u.Mobile,
		Password: string(hashedPass),
		IsAdmin:  u.IsAdmin,
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	result, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "Error inserting user data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
	fmt.Fprintln(w, "User Registered Successfully")
}

type User struct {
	User_ID  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Email    string             `json:"email" bson:"email"`
	Mobile   string             `json:"mobile" bson:"mobile"`
	Password string             `json:"password" bson:"password"`
	IsAdmin  bool               `json:"isAdmin" bson:"isAdmin"`
}

func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("invalid email address field")
	}
	if u.Password == "" {
		return errors.New("invalid password field")
	}
	if u.Mobile == "" {
		return errors.New("invalid mobile number")
	}
	return nil
}
