package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func main() {
	// Create a new session using the default AWS configuration
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create an EC2 service client
	svc := ec2.New(sess)


	// Specify the parameters for the new EC2 instance
	params := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-02f3f602d23f1659d"),
		InstanceType: aws.String("t2.micro"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
        {
            DeviceName: aws.String("/dev/sdf"),
            Ebs: &ec2.EbsBlockDevice{
                VolumeSize: aws.Int64(10), // Size of the volume in GB
                VolumeType: aws.String("gp2"), // Type of the volume
                DeleteOnTermination: aws.Bool(true), // Automatically delete the volume when the instance is terminated
            },
        },
    },
}
	// Create the EC2 instance
	result, err := svc.RunInstances(params)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	// Get the instance ID of the newly created EC2 instance
	instanceId := result.Instances[0].InstanceId

	// Specify the parameters for the new security group
	securityGroupName := "my-security-group-saba"
	securityGroupDescription := "My security group description"
	vpcId := "vpc-034c17f99ea7c941d"

	// Create the security group
	createSecurityGroupResult, err := svc.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(securityGroupName),
		Description: aws.String(securityGroupDescription),
		VpcId:       aws.String(vpcId),
	})
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	// Get the ID of the newly created security group
	securityGroupId := createSecurityGroupResult.GroupId

	// Open port 8000 in the new security group
	_, err = svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    securityGroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(8000),
		ToPort:     aws.Int64(8000),
		CidrIp:     aws.String("0.0.0.0/0"),
	})
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	// Open port 8080 in the new security group
	_, err = svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    securityGroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(8080),
		ToPort:     aws.Int64(8080),
		CidrIp:     aws.String("0.0.0.0/0"),
	})
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	// Attach the security group to the instance
	_, err = svc.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
		InstanceId: aws.String(*instanceId),
		Groups:     []*string{aws.String(*securityGroupId)},
	})
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	// Add tags to the instance
	_, err = svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{aws.String(*instanceId)},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("My EC2 Instance-sdk"),
			},
			{
				Key:   aws.String("Environment"),
				Value: aws.String("Development"),
			},
		},
	})
	if err != nil {
		fmt.Println("Error", err)
		return
	} 


	// Create an EC2 service client
	ec2svc := ec2.New(sess)
	
	// Wait for the instance to reach the running state
	err = ec2svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(*instanceId)},
	})
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	// Create an ELBV2 service client
	elbv2Svc := elbv2.New(sess)

	// Create a target group for the ALB
	createTGOutput, err := elbv2Svc.CreateTargetGroup(&elbv2.CreateTargetGroupInput{
		Name:        aws.String("my-target-group-sdk"),
		Protocol:    aws.String("HTTP"),
		Port:        aws.Int64(8000),
		VpcId:       aws.String("vpc-034c17f99ea7c941d"), // replace with your own VPC ID
		TargetType:  aws.String("instance"),
		HealthCheckProtocol: aws.String("HTTP"),
		HealthCheckPath:     aws.String("/healthcheck"),
		HealthCheckIntervalSeconds: aws.Int64(30),
		HealthyThresholdCount:      aws.Int64(2),
		UnhealthyThresholdCount:    aws.Int64(2),
	})
	if err != nil {
		fmt.Println("Error creating target group:", err)
		return
	}
	tgArn := createTGOutput.TargetGroups[0].TargetGroupArn
	fmt.Println("Target group created successfully")
	fmt.Println("ARN:", *tgArn)

	// Create a new ALB
	createLBOutput, err := elbv2Svc.CreateLoadBalancer(&elbv2.CreateLoadBalancerInput{
		Name:           aws.String("my-load-balancer-sdk"),
		Subnets:        []*string{aws.String("subnet-00de8f7cf2670b10a"), aws.String("subnet-0b92227f097bf4132")}, // replace with your own subnet IDs
		SecurityGroups: []*string{aws.String(*securityGroupId)}, // replace with your own security group IDs
		IpAddressType:  aws.String("ipv4"),
	})
	if err != nil {
		fmt.Println("Error creating ALB:", err)
		return
	}
	lbArn := createLBOutput.LoadBalancers[0].LoadBalancerArn
	lbDns := createLBOutput.LoadBalancers[0].DNSName
	fmt.Println("ALB created successfully")
	fmt.Println("ARN:", *lbArn)
	fmt.Println("DNS:", *lbDns)

	// Register the EC2 instance with the target group
	_, err = elbv2Svc.RegisterTargets(&elbv2.RegisterTargetsInput{
		TargetGroupArn: tgArn,
		Targets: []*elbv2.TargetDescription{
			{
				Id: aws.String(*instanceId),
			},
		},
	})
	if err != nil {
		fmt.Println("Error registering target:", err)
		return
	}
	fmt.Println("Target registered successfully")

	// Create a listener for the ALB
	_, err = elbv2Svc.CreateListener(&elbv2.CreateListenerInput{
		DefaultActions: []*elbv2.Action{
			{
				Type: aws.String("forward"),
				TargetGroupArn: tgArn,
			},
		},
		LoadBalancerArn: lbArn,
		Protocol:        aws.String("HTTP"),
		Port:            aws.Int64(8000),
	})
	if err != nil {
		fmt.Println("Error creating listener:", err)
		return
	}
	fmt.Println("Listener created successfully")
}
