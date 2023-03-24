package main

import (
    "fmt"
    "os"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    //"github.com/aws/aws-sdk-go/service/ec2"
    "github.com/aws/aws-sdk-go/service/elbv2"
)

func main() {
    // Set up an AWS session
    sess := session.Must(session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    }))

    // Create an EC2 client
    //ec2Svc := ec2.New(sess)

    // Create an ALB client
    elbv2Svc := elbv2.New(sess)

    // Define the EC2 instance ID that the ALB should target
    instanceID := "i-0ee9c00ee93381f3a"

    // Define the target group that the ALB should use
    targetGroupName := "my-target-group"

    // Create the target group
    targetGroup, err := elbv2Svc.CreateTargetGroup(&elbv2.CreateTargetGroupInput{
        Name: aws.String(targetGroupName),
        Port: aws.Int64(8000),
        Protocol: aws.String("HTTP"),
         VpcId: aws.String("vpc-034c17f99ea7c941d"), // add the VPC ID here
    })

    if err != nil {
        fmt.Println("Error creating target group:", err)
        os.Exit(1)
    }

    fmt.Println("Created target group:", *targetGroup.TargetGroups[0].TargetGroupArn)

    // Create the ALB
    alb, err := elbv2Svc.CreateLoadBalancer(&elbv2.CreateLoadBalancerInput{
        Name: aws.String("my-load-balancer-crud"),
        Subnets: []*string{
            aws.String("subnet-00de8f7cf2670b10a"),
            aws.String("subnet-0b92227f097bf4132"),
        },
        SecurityGroups: []*string{
            aws.String("sg-0ead112c434aa04d9"),
        },
        Scheme: aws.String("internet-facing"),
        Type: aws.String("application"),
    })

    if err != nil {
        fmt.Println("Error creating ALB:", err)
        os.Exit(1)
    }

    fmt.Println("Created ALB:", *alb.LoadBalancers[0].LoadBalancerArn)

    // Create a listener for the ALB
    listener, err := elbv2Svc.CreateListener(&elbv2.CreateListenerInput{
        LoadBalancerArn: alb.LoadBalancers[0].LoadBalancerArn,
        Protocol: aws.String("HTTP"),
        Port: aws.Int64(8000),
        DefaultActions: []*elbv2.Action{
            &elbv2.Action{
                Type: aws.String("forward"),
                TargetGroupArn: targetGroup.TargetGroups[0].TargetGroupArn,
            },
        },
    })

    if err != nil {
        fmt.Println("Error creating listener:", err)
        os.Exit(1)
    }

    fmt.Println("Created listener:", *listener.Listeners[0].ListenerArn)

    // Register the EC2 instance with the target group
    _, err = elbv2Svc.RegisterTargets(&elbv2.RegisterTargetsInput{
        TargetGroupArn: targetGroup.TargetGroups[0].TargetGroupArn,
        Targets: []*elbv2.TargetDescription{
            &elbv2.TargetDescription{
                Id: aws.String(instanceID),
                Port: aws.Int64(8000),
            },
        },
    })

    if err != nil {
        fmt.Println("Error registering target:", err)
        os.Exit(1)
    }

    fmt.Println("Registered target:", instanceID)
}
