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
