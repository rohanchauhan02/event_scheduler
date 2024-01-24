package sqs_plugin

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SQSPlugin is a GORM plugin for triggering SQS messages on updates
type SQSPlugin struct{}

// Name returns the name of the plugin
func (p *SQSPlugin) Name() string {
	return "SQSPlugin"
}

// Initialize initializes the plugin with the GORM DB instance
func (p *SQSPlugin) Initialize(db *gorm.DB) error {
	// Register the plugin's callback
	db.Callback().Update().After("gorm:after_update").Register("sqs_plugin:after_update", p.afterUpdate)
	return nil
}

// afterUpdate is the callback function to be executed after an update operation
func (p *SQSPlugin) afterUpdate(db *gorm.DB) {
	fmt.Println("After update callback triggered. Pushing message to SQS...")

	// Perform your SQS message pushing logic here
	// Example: pushMessageToSQS("Update operation detected!")

	fmt.Println("Message pushed to SQS successfully.")
}

type CustomModel struct {
	ID        uint `gorm:"primarykey"`
	Name      string
	UpdatedAt time.Time
}
