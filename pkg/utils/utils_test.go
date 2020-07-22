package utils

import (
	"encoding/base64"
	"reflect"
	"testing"
)



func TestEncodeValue(t *testing.T) {
	type args struct {
		m map[string]interface{}
	}
	v := "value1"
	arg := make(map[string]interface{})
	arg["test1"] = v

	vs := base64.StdEncoding.EncodeToString([]byte(v))
	expected := make(map[string]interface{})
	expected["test1"] = vs

	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "test1",
			args: args{m: arg},
			want: expected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeValue(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkUnicode(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			// test a ASCII charater
			name: "test-true",
			args: args{s: "dsad"},
			want: true,
		},
		{
			// test a unicode character
			name: "test-false",
			args: args{s: "vƝ"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkUnicode(tt.args.s); got != tt.want {
				t.Errorf("checkUnicode() = %v, want %v", got, tt.want)
			}
		})
	}
}
