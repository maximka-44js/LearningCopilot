package persistent

import "testing"

func TestNormalizeInsertStageOrder(t *testing.T) {
	tests := []struct {
		name      string
		requested int
		maxOrder  int
		expected  int
	}{
		{name: "append to empty project", requested: 1, maxOrder: 0, expected: 1},
		{name: "insert into middle", requested: 2, maxOrder: 5, expected: 2},
		{name: "append past end", requested: 99, maxOrder: 5, expected: 6},
		{name: "floor at one", requested: 0, maxOrder: 5, expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := normalizeInsertStageOrder(tt.requested, tt.maxOrder)
			if actual != tt.expected {
				t.Fatalf("normalizeInsertStageOrder(%d, %d) = %d, want %d", tt.requested, tt.maxOrder, actual, tt.expected)
			}
		})
	}
}

func TestNormalizeUpdateStageOrder(t *testing.T) {
	tests := []struct {
		name      string
		requested int
		maxOrder  int
		expected  int
	}{
		{name: "keep in range", requested: 3, maxOrder: 5, expected: 3},
		{name: "clamp to end", requested: 99, maxOrder: 5, expected: 5},
		{name: "floor at one", requested: 0, maxOrder: 5, expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := normalizeUpdateStageOrder(tt.requested, tt.maxOrder)
			if actual != tt.expected {
				t.Fatalf("normalizeUpdateStageOrder(%d, %d) = %d, want %d", tt.requested, tt.maxOrder, actual, tt.expected)
			}
		})
	}
}

func TestPlanStageOrderShift(t *testing.T) {
	tests := []struct {
		name         string
		currentOrder int
		targetOrder  int
		fromOrder    int
		toOrder      int
		delta        int
		ok           bool
	}{
		{name: "move up", currentOrder: 4, targetOrder: 2, fromOrder: 2, toOrder: 3, delta: 1, ok: true},
		{name: "move down", currentOrder: 2, targetOrder: 5, fromOrder: 3, toOrder: 5, delta: -1, ok: true},
		{name: "same position", currentOrder: 3, targetOrder: 3, ok: false},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			currentOrder := testCase.currentOrder
			targetOrder := testCase.targetOrder
			expectedFrom := testCase.fromOrder
			expectedTo := testCase.toOrder
			expectedDelta := testCase.delta
			expectedOK := testCase.ok

			fromOrder, toOrder, delta, ok := planStageOrderShift(currentOrder, targetOrder)
			if ok != expectedOK {
				t.Fatalf("planStageOrderShift ok = %v, want %v", ok, expectedOK)
			}
			if fromOrder != expectedFrom || toOrder != expectedTo || delta != expectedDelta {
				t.Fatalf("planStageOrderShift(%d, %d) = (%d, %d, %d, %v), want (%d, %d, %d, %v)", currentOrder, targetOrder, fromOrder, toOrder, delta, ok, expectedFrom, expectedTo, expectedDelta, expectedOK)
			}
		})
	}
}
