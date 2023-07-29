package address

import "testing"

func TestGetGRPCAddress(t *testing.T) {
	type args struct {
		hostandport string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "localhost:8100",
			args: args{hostandport: "localhost:8100"},
			want: "localhost:8200",
		}, {
			name: "localhost:18100",
			args: args{hostandport: "localhost:18100"},
			want: "localhost:18200",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGRPCAddress(tt.args.hostandport); got != tt.want {
				t.Errorf("GetGRPCAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
