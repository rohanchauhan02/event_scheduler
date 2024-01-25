package sqs_plugin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
	"gorm.io/gorm"
)

// SQSPlugin is a GORM plugin for triggering SQS messages on updates
type SQSPlugin struct {
	mu           sync.Mutex
	SQSClient    *sqs.SQS
	MysqlSess    *gorm.DB
	QueueURI     *string // Replace with your actual SQS Queue URL
	SQSMessage   *string // Replace with your actual SQS message structure
	OldData      interface{}
	NewData      interface{}
	TriggerEvent bool
}

func (p *SQSPlugin) Update(value interface{}, column string, query string, args ...interface{}) error {
	// Use a mutex to ensure exclusive access to the database
	p.mu.Lock()
	defer p.mu.Unlock()

	// Register the plugin's callback
	db := p.MysqlSess
	db.Callback().Update().Before("gorm:before_update").Register("before_update", func(d *gorm.DB) { p.beforeUpdate(d, column) })
	db.Callback().Update().After("gorm:after_update").Register("after_update", p.afterUpdate)

	// Perform the update
	if err := db.Where(query, args...).Updates(value).Error; err != nil {
		return err
	}

	return nil
}

func (p *SQSPlugin) Save(value interface{}, column string) error {
	// Use a mutex to ensure exclusive access to the database
	p.mu.Lock()
	defer p.mu.Unlock()

	// Register the plugin's callback
	db := p.MysqlSess
	db.Callback().Update().Before("gorm:before_update").Register("before_update", func(d *gorm.DB) { p.beforeUpdate(d, column) })
	db.Callback().Update().After("gorm:after_update").Register("after_update", p.afterUpdate)

	// Perform the save
	if err := db.Save(value).Error; err != nil {
		return err
	}
	return nil
}

// beforeUpdate is the callback function to be executed before an update operation
func (p *SQSPlugin) beforeUpdate(db *gorm.DB, columnName string) {
	fmt.Println("Before update callback triggered. Finding existing data...")

	// Debugging: Print the SQL query before execution
	db = db.Debug()

	// Query and store the existing data
	if err := db.First(p.OldData).Error; err != nil {
		// Handle the error if necessary
		fmt.Println("Error finding existing data:", err)
		return
	} else {
		fmt.Println("Existing data found:", p.OldData)
	}

	updatedData := db.Statement.Dest
	fmt.Println("Updated data:", updatedData)
	jsonData, err := json.Marshal(updatedData)
	if err != nil {
		fmt.Println("Error marshaling data to JSON:", err)
		return
	}

	if err := json.Unmarshal(jsonData, &p.NewData); err != nil {
		fmt.Println("Error unmarshaling JSON to NewData:", err)
		return
	}

	// Use reflection to dynamically access the field value based on the provided column name
	oldValueField := reflect.ValueOf(p.OldData).Elem().FieldByName(columnName)
	newValueField := reflect.ValueOf(p.NewData).Elem().FieldByName(columnName)

	// Check if the field is valid
	if !oldValueField.IsValid() || !newValueField.IsValid() {
		fmt.Println("Invalid column name:", columnName)
		return
	}

	// Check if the field is exportable (i.e., if it starts with an uppercase letter)
	if oldValueField.CanInterface() && newValueField.CanInterface() {
		oldValue := oldValueField.Interface()
		newValue := newValueField.Interface()

		// Check if the types are comparable
		if !reflect.TypeOf(oldValue).Comparable() || !reflect.TypeOf(newValue).Comparable() {
			fmt.Println("Non-comparable field type:", columnName)
			return
		}

		// Example condition to trigger an event if any change in the provided column
		if !reflect.DeepEqual(newValue, oldValue) {
			p.TriggerEvent = true
		} else {
			p.TriggerEvent = false
		}
	} else {
		fmt.Println("Non-exported field:", columnName)
	}
}

// afterUpdate is the callback function to be executed after an update operation
func (p *SQSPlugin) afterUpdate(db *gorm.DB) {
	fmt.Println("After update callback triggered. Pushing message to SQS...")

	// Debugging: Print the SQL query before execution
	_ = db.Debug()

	if p.TriggerEvent {
		_, err := p.SQSClient.SendMessage(&sqs.SendMessageInput{
			MessageBody: p.SQSMessage,
			QueueUrl:    p.QueueURI,
		})
		if err != nil {
			fmt.Println("Failed to publish message to SQS:", err)
			return
		}
		fmt.Println("Message successfully published to SQS.")
	}
}
