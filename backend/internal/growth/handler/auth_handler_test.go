package handler

import "testing"

func TestShouldBypassRiskControl(t *testing.T) {
	handler := &AuthHandler{relaxLocalRisk: true}

	cases := []struct {
		name string
		ip   string
		want bool
	}{
		{name: "ipv4 loopback", ip: "127.0.0.1", want: true},
		{name: "ipv6 loopback", ip: "::1", want: true},
		{name: "ipv6 mapped loopback", ip: "::ffff:127.0.0.1", want: true},
		{name: "remote ip", ip: "10.20.30.40", want: false},
		{name: "invalid ip", ip: "localhost", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := handler.shouldBypassRiskControl(tc.ip); got != tc.want {
				t.Fatalf("shouldBypassRiskControl(%q) = %v, want %v", tc.ip, got, tc.want)
			}
		})
	}
}

func TestShouldBypassRiskControlDisabled(t *testing.T) {
	handler := &AuthHandler{relaxLocalRisk: false}
	if handler.shouldBypassRiskControl("127.0.0.1") {
		t.Fatal("expected local risk control bypass to stay disabled")
	}
}
