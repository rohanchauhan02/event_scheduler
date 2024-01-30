package gormpluginv1

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	// "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/jinzhu/gorm"
)

// FinancePlugin is a GORM plugin for triggering finance event updates
type FinancePlugin struct {
	mu           sync.Mutex
	MysqlSess    *gorm.DB
	// SQSClient    *sqs.SQS
	// QueueURI     *string
	// SQSMessage   *string
	OldData      interface{}
	NewData      interface{}
	TriggerEvent bool
}

func (p *FinancePlugin) Update(value interface{}, column string, query string, args ...interface{}) error {
	// Use a mutex to ensure exclusive access to the database
	p.mu.Lock()
	defer p.mu.Unlock()

	// Register the plugin's callback
	db := p.MysqlSess
	db.Callback().Update().Before("gorm:before_update").Register("before_update", func(scope *gorm.Scope) { p.beforeUpdate(scope, column) })
	// Use "gorm:commit_or_rollback_transaction" to catch the commit event
	// db.Callback().Update().After("gorm:commit_or_rollback_transaction").Register("custom_after_commit", p.afterUpdateCommitOrRollback)

	// Perform the update
	if err := db.Where(query, args...).Updates(value).Error; err != nil {
		return err
	}

	return nil
}

func (p *FinancePlugin) Save(value interface{}, column string) error {
	// Use a mutex to ensure exclusive access to the database
	p.mu.Lock()
	defer p.mu.Unlock()

	// Register the plugin's callback
	db := p.MysqlSess
	db.Callback().Update().Before("gorm:before_update").Register("before_update", func(scope *gorm.Scope) { p.beforeUpdate(scope, column) })
	// db.Callback().Update().After("gorm:commit_or_rollback_transaction").Register("custom_after_commit", p.afterUpdateCommitOrRollback)

	// Perform the save
	if err := db.Save(value).Error; err != nil {
		return err
	}
	return nil
}

// beforeUpdate is the callback function to be executed before an update operation
func (p *FinancePlugin) beforeUpdate(scope *gorm.Scope, columnName string) {
	fmt.Println("Before update callback triggered. Finding existing data...")

	// Debugging: Print the SQL query before execution
	db := scope.DB().Debug()

	// Query and store the existing data
	if err := db.First(p.OldData).Error; err != nil {
		// Handle the error if necessary
		fmt.Println("Error finding existing data:", err)
		return
	} else {
		fmt.Println("Existing data found:", p.OldData)
	}

	updatedData := scope.Value
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

// afterUpdateCommitOrRollback is the callback function to be executed after a commit or rollback operation
// func (p *FinancePlugin) afterUpdateCommitOrRollback(scope *gorm.Scope) {
// 	if !scope.HasError() {
// 		p.AfterCommit(scope.DB())
// 	}
// }

// // AfterCommit is the callback function to be executed after a commit operation
// func (p *FinancePlugin) AfterCommit(db *gorm.DB) {
// 	fmt.Println("After commit callback triggered. Pushing message to SQS...")

// 	// Debugging: Print the SQL query before execution
// 	_ = db.Debug()

// 	if p.TriggerEvent {
// 		_, err := p.SQSClient.SendMessage(&sqs.SendMessageInput{
// 			MessageBody: p.SQSMessage,
// 			QueueUrl:    p.QueueURI,
// 		})
// 		if err != nil {
// 			fmt.Println("Failed to publish message to SQS:", err)
// 			return
// 		}
// 		fmt.Println("Message successfully published to SQS.")
// 	}
// }
