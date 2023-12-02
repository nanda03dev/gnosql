package utils

import (
	"github.com/google/uuid"
	"time"
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
