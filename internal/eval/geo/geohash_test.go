package geo

import "testing"



func TestGeoHash_Encode(t *testing.T)  {
	tests := []struct{
		name      string
		long      float64
		lat       float64
		precision int
		want      string
		wantErr   bool
	}{
		// Invalid cases
		{"Longitude Out of Range (high)", 200.0, 0.0, 5, "", true},
		{"Longitude Out of Range (low)", -200.0, 0.0, 5, "", true},
		{"Latitude Out of Range (high)", 0.0, 100.0, 5, "", true},
		{"Latitude Out of Range (low)", 0.0, -100.0, 5, "", true},
		{"Precision Too Low", 0.0, 0.0, 0, "", true},
		{"Precision Too High", 0.0, 0.0, 33, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeGeoHash(tt.long, tt.lat, tt.precision)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeGeoHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EncodeGeoHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeoHash_Decode(t *testing.T) {
	
}

func TestGeoHash_Neighbors(t *testing.T) {
	
}