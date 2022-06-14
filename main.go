package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	idFlag        = "id"
	itemFlag      = "item"
	operationFlag = "operation"
	fileNameFlag  = "fileName"

	addOperation  		= "add"
	listOperation  		= "list"
	findByIdOperation  	= "findById"
	removeOperation  	= "remove"

	messageItemAlreadyExistsTemplate = "Item with id %s already exists"
	messageItemNotFoundTemplate = "Item with id %s not found"

	errorMissingFlagTemplate = "-%s flag has to be specified"
	errorOperationNotAllowedTemplate = "Operation %s not allowed!"
)

type Arguments map[string]string

type dataRow struct {
	Id string `json:"id"`
	Email string `json:"email"`
	Age int `json:"age"`
}

type dataList []dataRow

func (a Arguments) String() string {
	return fmt.Sprintf("(%s %s) (%s %s) (%s %s) (%s %s)",
		operationFlag, a[operationFlag],
		itemFlag, a[itemFlag],
		idFlag, a[idFlag],
		fileNameFlag, a[fileNameFlag],
	)
}

func (dr dataRow) String() string {
	return "Row: " + dr.toJson()
}

func (dl dataList) String() string {
	return "List: " + dl.toJson()
}

func (dl dataList) indexOfRowWithId(id string) int {
	for i, v := range dl {
		if v.Id == id { return i }
	}
	return -1
}

func (dr dataRow) toJson() string {
	bytes, _ := json.Marshal(dr)
	return string(bytes)
}

func (dl dataList) toJson() string {
	bytes, _ := json.Marshal(dl)
	return string(bytes)
}

func (dr *dataRow) fromJson(textJson string) {
	_ = json.Unmarshal([]byte(textJson), dr)
}

func (dl *dataList) fromJson(textJson string) {
	_ = json.Unmarshal([]byte(textJson), dl)
}

func dataFromFile(filePath string) (dataList, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil { return nil, err }
	byteContainment, err := ioutil.ReadAll(file)
	if err != nil { return nil, err }
	result := dataList{}
	_ = json.Unmarshal(byteContainment, &result)
	return result, nil
}

func dataToFile(filePath string, dataRows dataList) error {
	if file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err == nil {
		if _, err = file.Write([]byte(dataRows.toJson())); err == nil {
			err = file.Close()
			return err
		} else { return err }
	} else { return err }
}

func parseArgs() Arguments {
	result := Arguments{
		operationFlag: *flag.String(operationFlag, "", operationFlag),
		itemFlag: *flag.String(itemFlag, "", itemFlag),
		idFlag: *flag.String(idFlag, "", idFlag),
		fileNameFlag: *flag.String(fileNameFlag, "", fileNameFlag),
	}
	flag.Parse()
	return result
}

func doOperationAdd(args Arguments, writer io.Writer) error {
	if item, ok := args[itemFlag]; !ok || len(item) == 0 {
		return fmt.Errorf(errorMissingFlagTemplate, itemFlag)
	} else {
		if list, err := dataFromFile(args[fileNameFlag]); err == nil {
			dr := dataRow{}
			dr.fromJson(item)
			if list.indexOfRowWithId(dr.Id) != -1 {
				_, err = writer.Write([]byte(fmt.Sprintf(messageItemAlreadyExistsTemplate, dr.Id)))
				return err
			}
			list = append(list, dr)
			if err := dataToFile(args[fileNameFlag], list); err != nil { return err }
			_, _ = writer.Write([]byte(list.toJson()))
			return err
		} else { return err }
	}
}

func doOperationList(args Arguments, writer io.Writer) error {
	if list, err := dataFromFile(args[fileNameFlag]); err == nil {
		_, err := writer.Write([]byte(list.toJson()))
		return err
	} else { return err }
}

func doOperationFindById(args Arguments, writer io.Writer) error {
	if findId, ok := args[idFlag]; !ok || len(findId) == 0 {
		return fmt.Errorf(errorMissingFlagTemplate, idFlag)
	} else {
		if list, err := dataFromFile(args[fileNameFlag]); err == nil {
			for _, v := range list {
				if v.Id == findId {
					_, err = writer.Write([]byte(v.toJson()))
					return err
				}
			}
			_, err = writer.Write([]byte(""))
			return err
		} else { return err }
	}
}

func doOperationRemove(args Arguments, writer io.Writer) error {
	if removeId, ok := args[idFlag]; !ok || len(removeId) == 0 {
		return fmt.Errorf(errorMissingFlagTemplate, idFlag)
	} else {
		if list, err := dataFromFile(args[fileNameFlag]); err == nil {
			if index := list.indexOfRowWithId(removeId); index != -1 {
				newList := make(dataList, 0, len(list)-1)
				newList = append(append(newList, list[0:index]...), list[index+1:]...)
				if err := dataToFile(args[fileNameFlag], newList); err == nil {
					_, _ = writer.Write([]byte(newList.toJson()))
				} else { return err }
			}
			_, _ = writer.Write([]byte(fmt.Sprintf(messageItemNotFoundTemplate, removeId)))
		} else { return err }
	}
	return nil
}

func Perform(args Arguments, writer io.Writer) error {
	if operation, ok := args[operationFlag]; !ok || len(operation) == 0 {
		return fmt.Errorf(errorMissingFlagTemplate, operationFlag)
	} else {
		if fileName, ok := args[fileNameFlag]; !ok || len(fileName) == 0 {
			return fmt.Errorf(errorMissingFlagTemplate, fileNameFlag)
		} else {
			switch operation {
			case addOperation:
				return doOperationAdd(args, writer)
			case listOperation:
				return doOperationList(args, writer)
			case findByIdOperation:
				return doOperationFindById(args, writer)
			case removeOperation:
				return doOperationRemove(args, writer)
			default:
				return fmt.Errorf(errorOperationNotAllowedTemplate, args[operationFlag])
			}
		}
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
