package component

import (
	"reflect"
	"testing"
)

func Test_getCandidate(t *testing.T) {
	type args struct {
		str   string
		width int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "上下左右BABA 5",
			args: args{
				str:   "上下左右BABA",
				width: 5,
			},
			want: []string{
				"上下 ",
				"下左 ",
				"左右B",
				"右BAB",
				"BABA ",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCandidate(tt.args.str, tt.args.width); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCandidate() = %v, want %v", got, tt.want)
			}
		})
	}
}
