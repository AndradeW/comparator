package comparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_compareJSON(t *testing.T) {

	type args struct {
		json1       map[string]interface{}
		json2       map[string]interface{}
		prefix      string
		differences map[string][]interface{}
	}

	tests := []struct {
		name                      string
		args                      args
		differencesExpected       map[string][]interface{}
		amountDifferencesExpected int
	}{
		{
			name: "Ok",
			args: args{
				json1:       map[string]interface{}{"a": "1"},
				json2:       map[string]interface{}{"a": "2"},
				prefix:      "test",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"test.a": {"1", "2"}},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok without prefix",
			args: args{
				json1:       map[string]interface{}{"a": "1"},
				json2:       map[string]interface{}{"a": "2"},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"a": {"1", "2"}},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with multiple differences",
			args: args{
				json1:       map[string]interface{}{"a": "1", "b": "2", "c": "3"},
				json2:       map[string]interface{}{"a": "11", "b": "12", "c": "13"},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"a": {"1", "11"}, "b": {"2", "12"}, "c": {"3", "13"}},
			amountDifferencesExpected: 3,
		},
		{
			name: "Key not found in second JSON",
			args: args{
				json1:       map[string]interface{}{"a": "1"},
				json2:       make(map[string]interface{}),
				prefix:      "test",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"test.a": {"1", "key not found in second JSON"}},
			amountDifferencesExpected: 1,
		}, {
			name: "Key not found in first JSON",
			args: args{
				json1:       make(map[string]interface{}),
				json2:       map[string]interface{}{"a": "1"},
				prefix:      "test",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"test.a": {"key not found in first JSON", "1"}},
			amountDifferencesExpected: 1,
		}}
	for _, tt := range tests {

		s := NewComparatorService(nil)

		t.Run(tt.name, func(t *testing.T) {
			s.compareJSON(tt.args.json1, tt.args.json2, tt.args.prefix, tt.args.differences)

			assert.Len(t, tt.args.differences, tt.amountDifferencesExpected)
			assert.Equal(t, tt.differencesExpected, tt.args.differences)
		})
	}
}
