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
			name: "Ok with Arrays without differences",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				json2:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       make(map[string][]interface{}),
			amountDifferencesExpected: 0,
		},
		{
			name: "Ok with Arrays",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				json2:       map[string]interface{}{"a": []interface{}{"13", "22", "33"}},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"a[0]": {"1", "13"}, "a[1]": {"2", "22"}, "a[2]": {"3", "33"}},
			amountDifferencesExpected: 3,
		},
		{
			name: "Ok with Arrays, second without array",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				json2:       map[string]interface{}{"a": "13"},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"a": {[]interface{}{"1", "2", "3"}, "13"}},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with Arrays, length difference",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
				json2:       map[string]interface{}{"a": []interface{}{}},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"a": {"different lengths", 3, 0}},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with Arrays of Json",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "1"}}},
				json2:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "1"}}},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       make(map[string][]interface{}),
			amountDifferencesExpected: 0,
		},
		{
			name: "Ok with Arrays of Json",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "1", "c": "2"}}},
				json2:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "11", "c": "12"}}},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"a[0].b": {"1", "11"}, "a[0].c": {"2", "12"}},
			amountDifferencesExpected: 2,
		},
		{
			name: "Ok with Arrays of Json, second without Json",
			args: args{
				json1:       map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": "1", "c": "2"}}},
				json2:       map[string]interface{}{"a": []interface{}{"11"}},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"a[0]": {map[string]interface{}{"b": "1", "c": "2"}, "11"}},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with Json of Json",
			args: args{
				json1:       map[string]interface{}{"a": map[string]interface{}{"b": "1"}},
				json2:       map[string]interface{}{"a": map[string]interface{}{"b": "1"}},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       make(map[string][]interface{}),
			amountDifferencesExpected: 0,
		},
		{
			name: "Ok with Json of Json",
			args: args{
				json1:       map[string]interface{}{"a": map[string]interface{}{"b": "1"}},
				json2:       map[string]interface{}{"a": map[string]interface{}{"b": "12"}},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"a.b": {"1", "12"}},
			amountDifferencesExpected: 1,
		},
		{
			name: "Ok with Json of Json, second without json",
			args: args{
				json1:       map[string]interface{}{"a": map[string]interface{}{"b": "1"}},
				json2:       map[string]interface{}{"a": "12"},
				prefix:      "",
				differences: make(map[string][]interface{}),
			},
			differencesExpected:       map[string][]interface{}{"a": {map[string]interface{}{"b": "1"}, "12"}},
			amountDifferencesExpected: 1,
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
