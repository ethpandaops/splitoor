package notifier

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test is for raw_message.go which is imported from the same package
// and contains the RawMessage type used in the Config struct

type TestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestRawMessageUnmarshal(t *testing.T) {
	// Create a test struct
	testStruct := TestStruct{
		Name:  "test",
		Value: 42,
	}

	// Create a RawMessage with custom unmarshal function
	rawMsg := &RawMessage{
		unmarshal: func(v interface{}) error {
			// Cast the input to our target type
			if result, ok := v.(*TestStruct); ok {
				// Set the values directly
				result.Name = testStruct.Name
				result.Value = testStruct.Value
				return nil
			}
			return fmt.Errorf("unsupported type")
		},
	}

	// Unmarshal to a struct
	var result TestStruct
	err := rawMsg.Unmarshal(&result)
	assert.NoError(t, err)

	// Verify the result
	assert.Equal(t, testStruct.Name, result.Name)
	assert.Equal(t, testStruct.Value, result.Value)
}

func TestRawMessageUnmarshalError(t *testing.T) {
	// Create a RawMessage with error-returning unmarshal function
	rawMsg := &RawMessage{
		unmarshal: func(v interface{}) error {
			return fmt.Errorf("mock unmarshal error")
		},
	}

	// Try unmarshaling to a valid target
	var result TestStruct
	err := rawMsg.Unmarshal(&result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock unmarshal error")
}

func TestRawMessageNilValue(t *testing.T) {
	// Test with zero value RawMessage
	var rawMsg RawMessage

	// Unmarshal should return error when unmarshal function is nil
	var result TestStruct
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Expected panic with nil unmarshal function")
	}()
	_ = rawMsg.Unmarshal(&result)
}

func TestRawMessageYAMLUnmarshal(t *testing.T) {
	// Create a RawMessage
	rawMsg := &RawMessage{}
	
	// Create a mock unmarshal function
	called := false
	mockUnmarshal := func(v interface{}) error {
		called = true
		return nil
	}
	
	// Call UnmarshalYAML
	err := rawMsg.UnmarshalYAML(mockUnmarshal)
	assert.NoError(t, err)
	
	// Verify the unmarshal function was set
	assert.NotNil(t, rawMsg.unmarshal)
	
	// Test that the unmarshal function is used
	var result TestStruct
	err = rawMsg.Unmarshal(&result)
	assert.NoError(t, err)
	assert.True(t, called, "The unmarshal function should have been called")
}
