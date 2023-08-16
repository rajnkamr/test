package ipassigner

import (
        "fmt"
        "net"
        "sync"
        "testing"
        respondertest "antrea.io/antrea/pkg/agent/ipassigner/responder/testing"
        "antrea.io/antrea/pkg/agent/util"
        "bou.ke/monkey"
        "github.com/golang/mock/gomock"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/mock"
        "github.com/stretchr/testify/require"
        "github.com/vishvananda/netlink"
        "k8s.io/apimachinery/pkg/util/sets"
)


func TestIPAssigner_AssignedIPs(t *testing.T) {
	// Create a dummy ipAssigner instance.
	// You may need to mock the data or implement a custom set for testing.
	// For simplicity, we'll assume a.assignedIPs is already set to some values.
	var dummyDevice netlink.Link
	var err error
	//if dummyDeviceName != "" {
	dummyDevice, err = fakeDummyDevice(dummyDeviceName)
	if err != nil {
		fmt.Errorf("error when ensuring dummy device exists: %v", err)
	}
	//      }

	controller := gomock.NewController(t)
	mockResponder := respondertest.NewMockResponder(controller)

	a := &ipAssigner{
		externalInterface: newFakeNetworkInterface(),
		dummyDevice:       dummyDevice,
		assignedIPs:       sets.New[string]("2.2.2.1", "3.3.3.1"), //tt.assignedIPs,
		arpResponder:      mockResponder,
		ndpResponder:      mockResponder,
		mutex:             sync.RWMutex{},
	}

	// Call the AssignedIPs() method.
	ips := a.AssignedIPs()

	// Check the result against the expected values.
	expectedIPs := sets.New[string]("3.3.3.1", "2.2.2.1")
	if !ips.Equal(expectedIPs) {
		t.Errorf("Expected IPs: %v, but got: %v", expectedIPs, ips)
	}
	fmt.Println("already assigned ip list", ips)
}

ifunc TestIPAssigner_AssignIP(t *testing.T) {
	var dummyDevice netlink.Link
	var err error

	dummyDevice, err = fakeDummyDevice(dummyDeviceName)
	if err != nil {
		fmt.Errorf("error when ensuring dummy device exists: %v", err)
	}

	tests := []struct {
		name                string
		ip                  string
		assignedIPs         sets.Set[string]
		expectedError       bool
		expectedAssignedIPs sets.Set[string]
	}{
		{
			name:                "Invalid IP",
			ip:                  "abc",
			assignedIPs:         sets.New[string](),
			expectedError:       true,
			expectedAssignedIPs: sets.New[string](),
		},
		{
			name:                "Assign new IP",
			ip:                  "2.1.1.1",
			assignedIPs:         sets.New[string](),
			expectedAssignedIPs: sets.New[string]("2.1.1.1"),
		},
		{
			name:                "Assign existing IP",
			ip:                  "2.1.1.1",
			assignedIPs:         sets.New[string]("2.1.1.1"),
			expectedError:       true,
			expectedAssignedIPs: sets.New[string]("2.1.1.1"),
		},
		{
			name:                "Add more IP",
			ip:                  "2.2.2.1",
			assignedIPs:         sets.New[string]("2.1.1.1"),
			expectedAssignedIPs: sets.New[string]("2.1.1.1", "2.2.2.1"),
		},
		{
			name:                "Assign IPv6",
			ip:                  "2001:db8::1",
			assignedIPs:         sets.New[string]("2.1.1.1", "2.2.2.1"),
			expectedAssignedIPs: sets.New[string]("2.1.1.1", "2.2.2.1", "2001:db8::1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			controller := gomock.NewController(t)
			mockResponder := respondertest.NewMockResponder(controller)

			a := &ipAssigner{
				externalInterface: newFakeNetworkInterface(),
				dummyDevice:       dummyDevice,
				assignedIPs:       tt.assignedIPs,
				arpResponder:      mockResponder,
				ndpResponder:      mockResponder,
				mutex:             sync.RWMutex{},
			}
			if tt.name != "Invalid IP" && tt.name != "Assign existing IP" {
				mockResponder.EXPECT().AddIP(net.ParseIP(tt.ip)).Return(nil)
			}

			patch := monkey.Patch((*ipAssigner).advertise, func(a *ipAssigner, ip net.IP) {
				fmt.Println("calling patched advertise func")
			})
			defer patch.Unpatch()

			fmt.Println("calling AssignIP", tt.name)
			errr := a.AssignIP(tt.ip, false)
			if tt.expectedError {
				fmt.Println("calling AssignIP inside error case", tt.name)
				assert.Error(t, errr, "Expected an error")
			} else {
				assert.NoError(t, errr)
			}
			assert.Equal(t, tt.expectedAssignedIPs, a.assignedIPs, "Assigned IPs don't match")
		})
	}
}
