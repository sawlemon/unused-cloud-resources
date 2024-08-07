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
