package unused_gcp_resources

import (
	"context"
	"log"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

type UnusedResourceMetrics struct {
	ResourceIDs []string
	TotalInstancesCount int
	UnusedInstancesCount int
}

func Get_Unused_Disks(projectId string, zone string) UnusedResourceMetrics  {
	ctx := context.Background()

	// Create a new client
	client, err := compute.NewDisksRESTClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	projectId = projectId
	zone = zone
	
	req := &computepb.ListDisksRequest{
		Project: projectId,
		Zone:    zone,
	}

	unused_disks := UnusedResourceMetrics{}
	totalDiskCount := 0
	unusedDiskCount := 0

	it := client.List(ctx, req)
	for {
		disk, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to list disks: %v", err)
		}
		
		totalDiskCount += 1

		// Check if the disk is attached
		if len(disk.GetUsers()) == 0 {
			unusedDiskCount += 1
			diskName := disk.GetName()
			unused_disks.ResourceIDs = append(unused_disks.ResourceIDs, diskName)
		}
	}
	
	unused_disks.TotalInstancesCount = totalDiskCount
	unused_disks.UnusedInstancesCount = unusedDiskCount

	return unused_disks
}

func Get_Unused_IPs(projectID string, region string) UnusedResourceMetrics {
	ctx := context.Background()

	// Create a Compute Service client
	client, err := compute.NewAddressesRESTClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	projectID = projectID
	region = region

	req := &computepb.ListAddressesRequest{
		// TODO: Fill request struct fields.
		// See https://pkg.go.dev/cloud.google.com/go/compute/apiv1/computepb#AggregatedListAddressesRequest.
		Project: projectID,
		Region: region,
	}

	unusedIPs := UnusedResourceMetrics{}
	totalIPCount := 0
	unusedIPCount := 0

	it := client.List(ctx, req)
	for {
		ips, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to list ips: %v", err)
		}
		
		totalIPCount += 1
		// Check if the IP is attached
		if len(ips.GetUsers()) == 0 {
			unusedIPCount += 1
			ipName := ips.GetName()
			unusedIPs.ResourceIDs = append(unusedIPs.ResourceIDs, ipName)
		}
	}

	unusedIPs.TotalInstancesCount = totalIPCount
	unusedIPs.UnusedInstancesCount = unusedIPCount

	return unusedIPs
}