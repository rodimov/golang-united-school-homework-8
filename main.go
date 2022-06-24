package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Arguments map[string]string

func Perform(args Arguments, writer io.Writer) error {
	fileName := args["fileName"]
	operation := args["operation"]
	item := args["item"]
	id := args["id"]

	if fileName == "" {
		return errors.New("-fileName flag has to be specified")
	}

	if operation == "" {
		return errors.New("-operation flag has to be specified")
	}

	if !stringInSlice(operation, []string{"list", "add", "remove", "findById"}) {
		return fmt.Errorf("Operation %s not allowed!", operation)
	}

	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(fileName)

		if err != nil {
			log.Fatal(err)
		}

		err = file.Close()

		if err != nil {
			log.Fatal(err)
		}
	}

	var err error = nil

	switch operation {
	case "list":
		printList(fileName, writer)
	case "add":
		err = addItem(item, fileName, writer)
	case "remove":
		err = removeUser(id, fileName, writer)
	case "findById":
		err = findUserById(id, fileName, writer)
	}

	return err
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func parseArgs() Arguments {
	operation := flag.String("operation", "", "operation type")
	item := flag.String("item", "", "item properties")
	fileName := flag.String("fileName", "", "file name")
	id := flag.String("id", "", "id of the user")
	flag.Parse()

	return map[string]string{
		"operation": *operation,
		"item":      *item,
		"fileName":  *fileName,
		"id":        *id,
	}
}

func readAll(fileName string) []byte {
	file, err := os.Open(fileName)

	if err != nil {
		log.Panicf("failed reading file: %s", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	data, err := ioutil.ReadAll(file)

	if err != nil {
		log.Fatal(err)
	}

	return data
}

func readUsersFromJSON(fileName string) []User {
	data := readAll(fileName)

	if string(data) == "" {
		return []User{}
	}

	var users []User
	err := json.Unmarshal(data, &users)

	if err != nil {
		log.Fatal(err)
	}

	return users
}

func writeUsersToJSON(users []User, fileName string) {
	data, err := json.Marshal(users)

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(fileName, data, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func getUserById(id string, users []User) *User {
	for _, user := range users {
		if user.Id == id {
			return &user
		}
	}

	return nil
}

func printList(fileName string, writer io.Writer) {
	_, err := writer.Write(readAll(fileName))

	if err != nil {
		log.Fatal(err)
	}
}

func addItem(item string, fileName string, writer io.Writer) error {
	if item == "" {
		return errors.New("-item flag has to be specified")
	}

	users := readUsersFromJSON(fileName)

	var user User
	err := json.Unmarshal([]byte(item), &user)

	if err != nil {
		log.Fatal(err)
	}

	userById := getUserById(user.Id, users)

	if userById != nil {
		_, err = writer.Write([]byte("Item with id " + user.Id + " already exists"))

		if err != nil {
			log.Fatal(err)
		}

		return nil
	}

	writeUsersToJSON(append(users, user), fileName)

	return nil
}

func removeUser(id, fileName string, writer io.Writer) error {
	if id == "" {
		return errors.New("-id flag has to be specified")
	}

	users := readUsersFromJSON(fileName)
	userById := getUserById(id, users)

	if userById == nil {
		_, err := writer.Write([]byte("Item with id " + id + " not found"))

		if err != nil {
			log.Fatal(err)
		}

		return nil
	}

	var usersToWrite []User

	for _, user := range users {
		if user.Id != id {
			usersToWrite = append(usersToWrite, user)
		}
	}

	writeUsersToJSON(usersToWrite, fileName)

	return nil
}

func findUserById(id, fileName string, writer io.Writer) error {
	if id == "" {
		return errors.New("-id flag has to be specified")
	}

	users := readUsersFromJSON(fileName)

	user := getUserById(id, users)

	if user == nil {
		return nil
	}

	data, err := json.Marshal(*user)

	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write(data)

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)

	if err != nil {
		panic(err)
	}
}
