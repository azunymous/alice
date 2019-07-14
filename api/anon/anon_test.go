package anon

import "testing"

func TestGenerateUsername(t *testing.T) {
	type args struct {
		seed int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with seed of 1",
			args: args{
				seed: 1,
			},
			want: "anonymous_of_Virgin_Islands,_U.S.",
		},
		{
			name: "Seed of 100 is different",
			args: args{
				seed: 100,
			},
			want: "anon-san_of_Benin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateUsername(tt.args.seed); got != tt.want {
				t.Errorf("GenerateUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}
