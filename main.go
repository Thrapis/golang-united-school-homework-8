package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	. "strings"
)

func unFlag(flag string) string {
	return Trim(flag, "-")
}

var (
	idFlag        = "-id"
	itemFlag      = "-item"
	operationFlag = "-operation"
	fileNameFlag  = "-fileName"

	idUnflag        = unFlag(idFlag)
	itemUnflag      = unFlag(itemFlag)
	operationUnflag = unFlag(operationFlag)
	fileNameUnflag  = unFlag(fileNameFlag)

	addOperation  		= "add"
	listOperation  		= "list"
	findByIdOperation  	= "findById"
	removeOperation  	= "remove"

	messageItemAlreadyExistsTemplate = "Item with id %s already exists"
	messageItemNotFoundTemplate = "Item with id %s not found"

	errorMissingFlagTemplate = "%s flag has to be specified"
	errorOperationNotAllowedTemplate = "Operation %s not allowed!"
)

type Arguments map[string]string

func (a Arguments) String() string {
	return fmt.Sprintf("(%s %s) (%s %s) (%s %s) (%s %s)",
		operationFlag, a[operationUnflag],
		itemFlag, a[itemUnflag],
		idFlag, a[idUnflag],
		fileNameFlag, a[fileNameUnflag],
		)
}

type dataRow struct {
	id, email, age string
}

func (dr dataRow) String() string {
	return "Row: " + dr.json()
}

type dataList []dataRow

func (dl dataList) String() string {
	return "List: " + dl.json()
}

func (dl dataList) ContainsWithId(id string) bool {
	for _, v := range dl {
		if v.id == id { return true }
	}
	return false
}

func dataRowFromString(stringRow string) dataRow {
	dr := dataRow{}
	stringRow = Trim(stringRow, "{")
	stringRow = Trim(stringRow, "}")
	for _, pair := range Split(stringRow, ",") {
		kv := Split(pair, ":")
		switch Trim(TrimSpace(kv[0]), "\"")  {
		case "id": dr.id = Trim(TrimSpace(kv[1]), "\"")
		case "email": dr.email = Trim(TrimSpace(kv[1]), "\"")
		case "age": dr.age = Trim(TrimSpace(kv[1]), "\"")
		}
	}
	return dr
}

func dataFromFile(filePath string) (dataList, error) {
	file, err := os.Open(filePath)
	if err != nil { return nil, err }
	byteContainment, err := ioutil.ReadAll(file)
	if err != nil { return nil, err }
	reRow := regexp.MustCompile(`\{(.*?)\}`)
	matches := reRow.FindAllStringSubmatch(string(byteContainment), -1)
	result := make(dataList, len(matches))
	for i:=0; i<len(matches); i++ {
		row := matches[i][1]
		result[i] = dataRowFromString(row)
	}
	return result, nil
}

func (dr dataRow) json() string {
	return fmt.Sprintf("{\"id\":\"%s\",\"email\":\"%s\",\"age\":%s}", dr.id, dr.email, dr.age)
}

func (dl dataList) json() string {
	result := "["
	for i, v := range dl {
		result += v.json()
		if i+1 < len(dl) { result += "," }
	}
	return result + "]"
}

func dataToFile(filePath string, dataRows dataList) error {
	if file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err == nil {
		if _, err = file.Write([]byte(dataRows.json())); err == nil {
			err = file.Close()
			return err
		} else { return err }
	} else { return err }
}

func parseArgs(arrArgs []string) Arguments {
	result := make(Arguments)
	for i := 0; i+1 < len(arrArgs); i++ {
		switch arrArgs[i] {
		case idFlag, itemFlag, operationFlag, fileNameFlag:
			result[unFlag(arrArgs[i])] = arrArgs[i+1]; i++
		}
	}
	return result
}

func doOperationAdd(args Arguments, writer io.Writer) error {
	fileName := args[fileNameUnflag]
	if item, ok := args[itemUnflag]; !ok || len(item) == 0 {
		return fmt.Errorf(errorMissingFlagTemplate, itemFlag)
	} else {
		if list, err := dataFromFile(fileName); err == nil {
			dr := dataRowFromString(item)
			if list.ContainsWithId(dr.id) {
				_, err = writer.Write([]byte(fmt.Sprintf(messageItemAlreadyExistsTemplate, dr.id)))
				return err
			}
			list = append(list, dr)
			if err := dataToFile(fileName, list); err != nil {
				return err
			}
			fmt.Println()
			_, err := writer.Write([]byte(list.json()))
			return err
		} else {
			return err
		}
	}
}

func doOperationList(args Arguments, writer io.Writer) error {
	fileName := args[fileNameUnflag]
	if list, err := dataFromFile(fileName); err == nil {
		_, err := writer.Write([]byte(list.json()))
		return err
	} else { return err }
}

func doOperationFindById(args Arguments, writer io.Writer) error {
	fileName := args[fileNameUnflag]
	if findId, ok := args[idUnflag]; !ok || len(findId) == 0 {
		return fmt.Errorf(errorMissingFlagTemplate, idFlag)
	} else {
		if list, err := dataFromFile(fileName); err == nil {
			for _, v := range list {
				if v.id == findId {
					_, err = writer.Write([]byte(v.json()))
					return err
				}
			}
			_, err = writer.Write([]byte(""))
			return err
		} else { return err }
	}
}

func doOperationRemove(args Arguments, writer io.Writer) error {
	fileName := args[fileNameUnflag]
	if removeId, ok := args[idUnflag]; !ok || len(removeId) == 0 {
		return fmt.Errorf(errorMissingFlagTemplate, idFlag)
	} else {
		if list, err := dataFromFile(fileName); err == nil {
			if !list.ContainsWithId(removeId) {
				_, err = writer.Write([]byte(fmt.Sprintf(messageItemNotFoundTemplate, removeId)))
				return err
			}
			for i, v := range list {
				if v.id == removeId {
					newList := make(dataList, 0, len(list)-1)
					newList = append(newList, list[0:i]...)
					newList = append(newList, list[i+1:]...)
					if err := dataToFile(fileName, newList); err != nil {
						return err
					}
					_, err = writer.Write([]byte(newList.json()))
					return err
				}
			}
			return nil
		} else { return err }
	}
}

func Perform(args Arguments, writer io.Writer) error {
	if operation, ok := args[operationUnflag]; !ok || len(operation) == 0 {
		return fmt.Errorf(errorMissingFlagTemplate, operationFlag)
	} else {
		if fileName, ok := args[fileNameUnflag]; !ok || len(fileName) == 0 {
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
				return fmt.Errorf(errorOperationNotAllowedTemplate, args[operationUnflag])
			}
		}
	}
}

func main() {
	err := Perform(parseArgs(os.Args), os.Stdout)
	if err != nil {
		panic(err)
	}
}
