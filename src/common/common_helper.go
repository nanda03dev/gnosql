package common

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"gnosql/src/global_constants"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func DeleteElement(array []string, elementToDelete string) []string {
	// Find the index of the element
	indexToDelete := -1
	for i, v := range array {
		if v == elementToDelete {
			indexToDelete = i
			break
		}
	}

	// If the element is found, remove it
	if indexToDelete != -1 {
		array = append(array[:indexToDelete], array[indexToDelete+1:]...)
	}
	return array
}

func Generate16DigitUUID() string {
	id, _ := uuid.NewUUID()
	return id.String()
}

func ExtractTimestampFromUUIDString(uuidStr string) time.Time {
	uuid := uuid.MustParse(uuidStr)

	t := uuid.Time()
	sec, nsec := t.UnixTime()
	timeStamp := time.Unix(sec, nsec)
	return timeStamp
}

func UuidStringToTimeString(uuidStr string) string {
	uuid := uuid.MustParse(uuidStr)

	t := uuid.Time()
	sec, nsec := t.UnixTime()
	timeStamp := time.Unix(sec, nsec)
	return TimeToString(timeStamp)
}

func TimeToString(time time.Time) string {
	return time.UTC().Format("2006-01-02T15:04:05Z07:00")
}

func CreateDatabaseFolder() bool {
	if _, err := CreateFolder(global_constants.GNOSQL_FULL_PATH); err == nil {
		return true
	}
	return false
}

func CreateFolder(nestedFolderPath string) (string, error) {

	// Check if the nested folders already exist
	if _, err := os.Stat(nestedFolderPath); os.IsNotExist(err) {

		// Nested folders do not exist, create them
		err := os.MkdirAll(nestedFolderPath, 0755) // 0755 is the permission mode for the new folders
		if err != nil {
			println("Error while create %s directory %v", nestedFolderPath, err)
			return "", err
		}
		println("GnoSQL database folder created successfully")
	} else {
		println("GnoSQL database folder already exists")
	}

	return nestedFolderPath, nil
}

func ReadFoldersInDirectory(directoryPath string) ([]string, error) {
	var fileNames []string

	// Read the directory
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		println("database file names reading, Error %v", err)
		return nil, err
	}

	// Iterate over the files
	for _, file := range files {
		// Check if the entry is a file (not a directory)
		if file.IsDir() {
			filePath := filepath.Join(directoryPath, file.Name())

			// Append the file path to the slice
			fileNames = append(fileNames, filePath)
		}

		// Construct the full path to the file

	}

	return fileNames, nil
}

func ReadFileNamesInDirectory(directoryPath string) ([]string, error) {
	var fileNames []string

	// Read the directory
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		println("database file names reading, Error %v", err)
		return nil, err
	}

	// Iterate over the files
	for _, file := range files {
		// Check if the entry is a file (not a directory)
		if file.IsDir() {
			continue
		}

		// Construct the full path to the file
		filePath := filepath.Join(directoryPath, file.Name())

		// Append the file path to the slice
		fileNames = append(fileNames, filePath)
	}

	return fileNames, nil
}

func ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func GetDatabaseFileName(databaseName string) string {
	return databaseName + global_constants.DB_EXTENSION
}
func GetDatabaseFolderPath(databaseName string) string {
	return filepath.Join(global_constants.GNOSQL_FULL_PATH, databaseName)
}
func GetDatabaseFilePath(databaseName, fileName string) string {
	return filepath.Join(global_constants.GNOSQL_FULL_PATH, databaseName+"/"+fileName)
}

func GetCollectionFileName(collectionName string) string {
	return collectionName + global_constants.COLLECTION_EXTENSION
}

func GetCollectionBatchIdFileName() string {
	return Generate16DigitUUID() + global_constants.COLLECTION_BATCH_EXTENSION
}
func GetCollectionFolderPath(databaseName string, collectionName string) string {
	return filepath.Join(global_constants.GNOSQL_FULL_PATH, databaseName+"/"+collectionName)
}
func GetCollectionFilePath(databaseName string, collectionName string, fileName string) string {
	return GetCollectionFolderPath(databaseName, collectionName) + "/" + fileName
}

func DeleteFolder(filePath string) bool {
	err := os.RemoveAll(filePath)
	if err != nil {
		fmt.Println("Error deleting file or directory:", err)
		return false
	}

	return true
}

func DeleteFile(filePath string) bool {
	err := os.Remove(filePath)
	if err != nil {
		println("Error deleting file:", err)
		return false
	}

	return true
}

func EncodeGob(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	return buf.Bytes(), err
}

func DecodeGob(data []byte, target interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(target)
}

func SaveToFile(filename string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		fmt.Printf("\n err %v ", err)

		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func ReadFromFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func WriteGobDataToDisk(filePath string, data []byte) {
	err := SaveToFile(filePath, data)

	if err != nil {
		fmt.Printf("\n filePath: %v Error saving collection GOB to file:", err)
	}
}

func EncodeGOBAndWriteFile[T any](filePath string, data T) (error, error) {
	gobData, encodeErr := EncodeGob(data)
	var writeErr error

	if encodeErr == nil {
		writeErr = SaveToFile(filePath, gobData)
	}

	fmt.Printf("\n Error while write file: %v \t EncodeError: %v \t WriteFileError: %v ", filePath, encodeErr, writeErr)

	return encodeErr, writeErr
}

func ReadFileAndDecodeGOB[T any](filePath string) (T, error) {
	var data T

	fileData, err := os.ReadFile(filePath)

	if err != nil {
		fmt.Printf("\n Reading file %s, Error %v", filePath, err)
		return data, err
	}

	err = DecodeGob(fileData, &data)

	if err != nil {
		fmt.Printf("\n Decoding file %s, Error %v", filePath, err)

		return data, err
	}

	return data, nil
}
