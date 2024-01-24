package sqs_plugin

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
	"gorm.io/gorm"
)

// SQSPlugin is a GORM plugin for triggering SQS messages on updates
type SQSPlugin struct {
	SQSClient  *sqs.SQS
	SQSMessage *string
	QueueURI   *string
}

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
	_, err := p.SQSClient.SendMessage(&sqs.SendMessageInput{
		MessageBody: p.SQSMessage,
		QueueUrl:    p.QueueURI,
	})
	if err != nil {
		fmt.Println("Failed to published message to sqs")
		return
	}
	fmt.Println("Message pushed to SQS successfully.")
}
