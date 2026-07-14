package purchaseservice

import "testing"

func TestMinorUnitCost(t *testing.T) {
	tests := []struct {
		name      string
		price     int64
		quantity  int64
		precision int
		want      int64
		wantErr   bool
	}{
		{name: "single item precision 2", price: 100, quantity: 1, precision: 2, want: 10000},
		{name: "multiple items", price: 500, quantity: 3, precision: 2, want: 150000},
		{name: "zero precision", price: 7, quantity: 2, precision: 0, want: 14},
		{name: "zero quantity rejected", price: 100, quantity: 0, precision: 2, wantErr: true},
		{name: "negative quantity rejected", price: 100, quantity: -5, precision: 2, wantErr: true},
		{name: "zero price rejected", price: 0, quantity: 1, precision: 2, wantErr: true},
		{name: "price*quantity overflow", price: 1 << 40, quantity: 1 << 40, precision: 0, wantErr: true},
		{name: "precision factor overflow", price: 1 << 60, quantity: 1, precision: 2, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := minorUnitCost(tt.price, tt.quantity, tt.precision)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}
