package main

import (
	"fmt"

	gcp_unused "github.com/sawlemon/unused-cloud-resources/gcp_unused_resources"
)

func main() {
	unused_disk_data := gcp_unused.Get_Unused_Disks("finops-accelerator", "us-central1-a")

	fmt.Printf("Unused Disks Name %s\nTotal Disk Count %d\nUnused Disk: %d\n",
		unused_disk_data.ResourceIDs,
		unused_disk_data.TotalInstancesCount,
		unused_disk_data.UnusedInstancesCount,
	)

	unusedIPsData := gcp_unused.Get_Unused_IPs("finops-accelerator", "us-central1")

	fmt.Printf("Unused IPs Name %s\nTotal IPs Count %d\nUnused IPs Count: %d\n",
		unusedIPsData.ResourceIDs,
		unusedIPsData.TotalInstancesCount,
		unusedIPsData.UnusedInstancesCount,
		// unusedIPsData.UnusedInstancesCount / unusedIPsData.TotalInstancesCount * 100,
	)
}

// Unused Disks Name [disk-1]
// Total Disk Count 1
// Unused Disk: 1
