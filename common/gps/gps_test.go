package gps

import (
	"reflect"
	"testing"
)

func Test_getUnitCircleAngle(t *testing.T) {
	type args struct {
		from GPSRecord
		to   GPSRecord
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "northeast",
			args: args{
				from: GPSRecord{
					UnixMicro: 1,
					Lat:       37.0,
					Long:      -89.0,
				},
				to: GPSRecord{
					UnixMicro: 999_999,
					Lat:       39.0,  // > = to the north
					Long:      -87.0, // > =  to the east
				},
			},
			want: 45.0,
		},
		{
			name: "more east than north",
			args: args{
				from: GPSRecord{
					UnixMicro: 1,
					Lat:       37.0,
					Long:      -89.0,
				},
				to: GPSRecord{
					UnixMicro: 999_999,
					Lat:       38.0,  // > = to the north
					Long:      -87.0, // > = to the east
				},
			},
			want: 26.565051177077986,
		},
		{
			name: "southeast",
			args: args{
				from: GPSRecord{
					UnixMicro: 1,
					Lat:       37.0,
					Long:      -89.0,
				},
				to: GPSRecord{
					UnixMicro: 999_999,
					Lat:       35.0,  // < = to the south
					Long:      -87.0, // > = to the east
				},
			},
			want: 315.0,
		},
		{
			name: "more east than south",
			args: args{
				from: GPSRecord{
					UnixMicro: 1,
					Lat:       37.0,
					Long:      -89.0,
				},
				to: GPSRecord{
					UnixMicro: 999_999,
					Lat:       36.0,  // < = to the north
					Long:      -87.0, // > = to the east
				},
			},
			want: 333.434948822922,
		},
		{
			name: "northwest",
			args: args{
				from: GPSRecord{
					UnixMicro: 1,
					Lat:       37.0,
					Long:      -89.0,
				},
				to: GPSRecord{
					UnixMicro: 999_999,
					Lat:       39.0,  // > = to the north
					Long:      -91.0, // < =  to the west
				},
			},
			want: 135.0,
		},
		{
			name: "more west than north",
			args: args{
				from: GPSRecord{
					UnixMicro: 1,
					Lat:       37.0,
					Long:      -89.0,
				},
				to: GPSRecord{
					UnixMicro: 999_999,
					Lat:       38.0,  // > = to the north
					Long:      -91.0, // < =  to the west
				},
			},
			want: 153.43494882292202,
		},
		{
			name: "southwest",
			args: args{
				from: GPSRecord{
					UnixMicro: 1,
					Lat:       37.0,
					Long:      -89.0,
				},
				to: GPSRecord{
					UnixMicro: 999_999,
					Lat:       35.0,  // < = to the south
					Long:      -91.0, // < =  to the west
				},
			},
			want: 225.0,
		},
		{
			name: "more west than south",
			args: args{
				from: GPSRecord{
					UnixMicro: 1,
					Lat:       37.0,
					Long:      -89.0,
				},
				to: GPSRecord{
					UnixMicro: 999_999,
					Lat:       36.0,  // < = to the south
					Long:      -91.0, // < =  to the west
				},
			},
			want: 206.56505117707798,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUnitCirAngle(tt.args.from, tt.args.to); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVector() = %v, want %v", got, tt.want)
			}
		})
	}
}
