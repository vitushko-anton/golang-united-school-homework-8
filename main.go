package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   uint   `json:"age"`
}

func Perform(args Arguments, writer io.Writer) error {
	if _, ok := args["operation"]; !ok || args["operation"] == "" {
		return errors.New("-operation flag has to be specified")
	}
	if _, ok := args["fileName"]; !ok || args["fileName"] == "" {
		return errors.New("-fileName flag has to be specified")
	}
	file, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	switch args["operation"] {
	case "list":
		list(file, writer)
		break
	case "add":
		if _, ok := args["item"]; !ok || args["item"] == "" {
			return errors.New("-item flag has to be specified")
		}
		err = add(args["item"], file, writer)
		if err != nil {
			writer.Write([]byte(err.Error()))
		}
	case "findById":
		if _, ok := args["id"]; !ok || args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}
		user, err := find(args["id"], file)
		if err != nil {
			return err
		}
		if user == nil {
			writer.Write([]byte(""))
		} else {
			b, _ := json.Marshal(user)
			writer.Write(b)
		}
		break
	case "remove":
		if _, ok := args["id"]; !ok || args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}
		user, err := find(args["id"], file)
		if err != nil {
			return err
		}
		if user == nil {
			writer.Write([]byte(fmt.Errorf("Item with id %s not found", args["id"]).Error()))
		} else {
			remove(user, file)
		}
		break
	default:
		return fmt.Errorf("Operation %s not allowed!", args["operation"])
	}
	return nil
}

func list(file *os.File, writer io.Writer) {
	b, _ := ioutil.ReadAll(file)
	writer.Write(b)
	file.Seek(0, 0)
}

func remove(user *User, file *os.File) error {
	defer file.Seek(0, 0)
	b, _ := ioutil.ReadAll(file)
	var users []User
	_ = json.Unmarshal(b, &users)
	for k, u := range users {
		if u.Id == user.Id {
			users = append(users[:k], users[k+1:]...)
		}
	}

	list, _ := json.Marshal(users)
	file.Truncate(0)
	file.Seek(0, 0)
	file.Write(list)

	return nil
}

func add(item string, file *os.File, writer io.Writer) error {
	user := &User{}
	err := json.Unmarshal([]byte(item), user)
	if err != nil {
		return err
	}
	u, err := find(user.Id, file)
	if u != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("Item with id %s already exists", user.Id)
	}
	b, _ := ioutil.ReadAll(file)
	var users []User
	_ = json.Unmarshal(b, &users)
	users = append(users, *user)
	byteUsers, _ := json.Marshal(users)
	file.Write(byteUsers)
	list(file, writer)
	return nil
}

func find(id string, file *os.File) (*User, error) {
	defer file.Seek(0, 0)
	b, _ := ioutil.ReadAll(file)
	var users []User
	err := json.Unmarshal(b, &users)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Id == id {
			return &user, nil
		}
	}
	return nil, nil
}

func parseArgs() Arguments {
	id := flag.String("id", "", "item ID for finding")
	operation := flag.String("operation", "",
		"list - Getting list of Items\n"+
			"add - Adding item to the Collection\n"+
			"findById - Getting Item by ID\n"+
			"remove - Delete Item from the Collection")
	item := flag.String("item", "", "json example:\n"+
		"{\n    id: \"1\",\n    email: \"test@test.com\",\n    age: 31\n}")
	fileName := flag.String("fileName", "", "json filename")
	flag.Parse()
	return Arguments{
		"id":        *id,
		"operation": *operation,
		"item":      *item,
		"fileName":  *fileName,
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
