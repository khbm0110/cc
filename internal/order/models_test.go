package order

import "testing"

func TestOrderStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		terminal bool
	}{
		{StatusPending, false},
		{StatusExecuting, false},
		{StatusFilled, true},
		{StatusFailed, true},
		{StatusCanceled, true},
	}

	for _, tt := range tests {
		if tt.status.IsTerminal() != tt.terminal {
			t.Errorf("status %s: expected terminal=%v, got %v", tt.status, tt.terminal, tt.status.IsTerminal())
		}
	}
}

func TestOrderStatus_IsValid(t *testing.T) {
	tests := []struct {
		status OrderStatus
		valid  bool
	}{
		{StatusPending, true},
		{StatusExecuting, true},
		{StatusFilled, true},
		{StatusFailed, true},
		{StatusCanceled, true},
		{OrderStatus("UNKNOWN"), false},
		{OrderStatus(""), false},
	}

	for _, tt := range tests {
		if tt.status.IsValid() != tt.valid {
			t.Errorf("status %q: expected valid=%v, got %v", tt.status, tt.valid, tt.status.IsValid())
		}
	}
}

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		from  OrderStatus
		to    OrderStatus
		valid bool
	}{
		{StatusPending, StatusExecuting, true},
		{StatusPending, StatusCanceled, true},
		{StatusPending, StatusFailed, true},
		{StatusPending, StatusFilled, false},
		{StatusExecuting, StatusFilled, true},
		{StatusExecuting, StatusFailed, true},
		{StatusExecuting, StatusCanceled, true},
		{StatusExecuting, StatusPending, false},
		{StatusFilled, StatusPending, false},
		{StatusFilled, StatusFailed, false},
		{StatusFailed, StatusPending, false},
		{StatusCanceled, StatusPending, false},
	}

	for _, tt := range tests {
		if isValidTransition(tt.from, tt.to) != tt.valid {
			t.Errorf("transition %s -> %s: expected valid=%v", tt.from, tt.to, tt.valid)
		}
	}
}
