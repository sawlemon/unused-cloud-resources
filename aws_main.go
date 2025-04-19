package main

import (
	"fmt"

	aws_unused "github.com/sawlemon/unused-cloud-resources/aws_unused_resources"
)

func main() {
	unused_ebs_data := aws_unused.Get_unused_ebs_volumes("us-east-1")
	fmt.Printf("Unused EBS Volume IDs %s\nTotal Volume Count %d\nUnused Count: %d",
		unused_ebs_data.ResourceIDs,
		unused_ebs_data.TotalInstancesCount,
		unused_ebs_data.UnusedInstancesCount,
	)

	unused_ec2_data := aws_unused.Get_unused_ebs_volumes("us-east-1")
	fmt.Printf("Unused EC2 Instances IDs %s\nTotal Volume Count %d\nUnused Count: %d",
		unused_ec2_data.ResourceIDs,
		unused_ec2_data.TotalInstancesCount,
		unused_ec2_data.UnusedInstancesCount,
	)
}
