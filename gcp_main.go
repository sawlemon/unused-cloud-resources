package main

import (
	"fmt"

	gcp_unused "github.com/sawlemon/unused-cloud-resources/gcp_unused_resources"
)

func main() {
	unused_disk_data := gcp_unused.Get_Unused_Disks("root-fort-431811-j3", "us-central1-a")
	fmt.Printf("Unused Disks Name %s\nTotal Disk Count %d\nUnused Disk: %d\n", 
		unused_disk_data.ResourceIDs, 
		unused_disk_data.TotalInstancesCount,
		unused_disk_data.UnusedInstancesCount,
	)
}


// Unused Disks Name [disk-1]
// Total Disk Count 1
// Unused Disk: 1