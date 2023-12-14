package utils

import (
	"bytes"
	"encoding/gob"
	"os"
	"os/user"
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
	uuidObj := uuid.New()
	return uuidObj.String()
}

// extractTimestampFromUUID extracts the timestamp from a version 1 UUID
func ExtractTimestampFromUUID(uuidStr string) time.Time {
	u, err := uuid.Parse(uuidStr)
	if err != nil {
		print(err)
	}
	// Version 1 UUID layout: time_low-time_mid-time_hi_and_version-clock_seq_hi_and_reserved-clock_seq_low-node
	// Extract timestamp from time_low, time_mid, and time_hi_and_version
	timestamp := int64(u[0])<<56 | int64(u[1])<<48 | int64(u[2])<<40 | int64(u[3])<<32 | int64(u[4])<<24 | int64(u[5])<<16 | int64(u[6])<<8 | int64(u[7])
	return time.Unix(0, timestamp)
}

func CreateDatabaseFolder() bool {
	if _, err := CreateFolder(GNOSQLPATH); err == nil {
		return true
	}
	return false
}

func CreateFolder(folderName string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		println("Error while getting user directory %v", err)
		return "", err
	}

	// Construct the full path to the nested folders in the user's home directory
	nestedFolderPath := filepath.Join(usr.HomeDir, folderName)

	// Check if the nested folders already exist
	if _, err := os.Stat(nestedFolderPath); os.IsNotExist(err) {

		// Nested folders do not exist, create them
		err := os.MkdirAll(nestedFolderPath, 0755) // 0755 is the permission mode for the new folders
		if err != nil {
			println("Error while create %s directory %v", folderName, err)
			return "", err
		}
		println("gnosql database folder created successfully")
	} else {
		println("gnosql database folder already exists")
	}

	return nestedFolderPath, nil
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
	return databaseName + "-db.gob"
}

func GetDatabaseFilePath(fileName string) string {
	return filepath.Join(GNOSQLFULLPATH, fileName)
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
	return os.WriteFile(filename, data, 0644)
}

func ReadFromFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
