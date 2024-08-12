package main

import (
	"github.com/gin-gonic/gin"
	aws_unused "github.com/sawlemon/unused-cloud-resources/aws_unused_resources"
	gcp_unused "github.com/sawlemon/unused-cloud-resources/gcp_unused_resources"
	"net/http"
)

func getAWSEBSData(c *gin.Context) {
	unused_ebs_data := aws_unused.Get_unused_ebs_volumes("us-east-1", "limited-admin")
	percentage_unused := 100 * unused_ebs_data.UnusedInstancesCount / unused_ebs_data.TotalInstancesCount
	c.JSON(http.StatusOK, gin.H{
		"resource_ids": unused_ebs_data.ResourceIDs,
		"percentage":   percentage_unused,
		"total_count":  unused_ebs_data.TotalInstancesCount,
		"unused_count": unused_ebs_data.UnusedInstancesCount,
	})
}

func getGCPUnusedDisk(c *gin.Context) {
	unused_disk_data := gcp_unused.Get_Unused_Disks("finops-accelerator", "us-central1-a")
	percentage_unused := 100 * unused_disk_data.UnusedInstancesCount / unused_disk_data.TotalInstancesCount
	c.JSON(http.StatusOK, gin.H{
		"resource_ids": unused_disk_data.ResourceIDs,
		"percentage":   percentage_unused,
		"total_count":  unused_disk_data.TotalInstancesCount,
		"unused_count": unused_disk_data.UnusedInstancesCount,
	})
}

func getGCPUnusedIP(c *gin.Context) {
	unusedIPsData := gcp_unused.Get_Unused_IPs("finops-accelerator", "us-central1")
	percentage_unused := 100 * unusedIPsData.UnusedInstancesCount / unusedIPsData.TotalInstancesCount
	c.JSON(http.StatusOK, gin.H{
		"resource_ids": unusedIPsData.ResourceIDs,
		"percentage":   percentage_unused,
		"total_count":  unusedIPsData.TotalInstancesCount,
		"unused_count": unusedIPsData.UnusedInstancesCount,
	})
}

func main() {
	r := gin.Default()

	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "Okay",
		})
	})

	r.GET("/aws/ebs", getAWSEBSData)
	r.GET("/gcp/disks", getGCPUnusedDisk)
	r.GET("/gcp/ips", getGCPUnusedIP)

	r.Run(":9090") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
