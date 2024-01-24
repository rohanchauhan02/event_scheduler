package sqs_plugin

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
	"gorm.io/gorm"
)

// SQSPlugin is a GORM plugin for triggering SQS messages on updates
type SQSPlugin struct {
	SQSClient *sqs.SQS
	MysqlSess *gorm.DB
}

func (p *SQSPlugin) Update(value interface{}) error {
	// Register the plugin's callback
	p.MysqlSess.Callback().Update().Before("gorm:before_update").Register("sqs_plugin:before_update", p.beforeUpdate)
	p.MysqlSess.Callback().Update().After("gorm:after_update").Register("sqs_plugin:after_update", p.afterUpdate)
	// p.MysqlSess.Updates(value)
	return nil
}

// beforeUpdate is the callback function to be executed before an update operation
func (p *SQSPlugin) beforeUpdate(db *gorm.DB) {
	fmt.Println("before update callback triggered. Pushing message to SQS...")
	fmt.Println(db.Statement.SQL.String())
}

// afterUpdate is the callback function to be executed after an update operation
func (p *SQSPlugin) afterUpdate(db *gorm.DB) {
	fmt.Println("After update callback triggered. Query......")
	// Perform your SQS message pushing logic here
	// _, err := p.SQSClient.SendMessage(&sqs.SendMessageInput{
	// 	MessageBody: p.SQSMessage,
	// 	QueueUrl:    p.QueueURI,
	// })
	// if err != nil {
	// 	fmt.Println("Failed to published message to sqs")
	// 	return
	// }
	fmt.Println(db.Statement.SQL.String())
}
