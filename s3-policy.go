package main

import (
    "fmt"
    "os"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/kms"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "github.com/aws/aws-sdk-go/service/s3"
)

func main() {
    // Set up a new AWS session
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1"),
    })
    if err != nil {
        fmt.Println("Error creating session:", err)
        os.Exit(1)
    }

    // Create a new S3 client
    svc := s3.New(sess)

    // Set up the bucket name and policy
    bucketName := "my-public-bucket-sdk-demo"
    policy := `{
        "Version":"2012-10-17",
        "Statement":[{
            "Sid":"PublicReadGetObject",
            "Effect":"Allow",
            "Principal": "*",
            "Action":["s3:GetObject"],
            "Resource":["arn:aws:s3:::` + bucketName + `/*"]
        },{
            "Sid":"PublicWritePutObject",
            "Effect":"Allow",
            "Principal": "*",
            "Action":["s3:PutObject"],
            "Resource":["arn:aws:s3:::` + bucketName + `/*"]
        }]
    }`

    // Create the S3 bucket
    _, err = svc.CreateBucket(&s3.CreateBucketInput{
        Bucket: aws.String(bucketName),
    })
    if err != nil {
        fmt.Println("Error creating bucket:", err)
        if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "BucketAlreadyOwnedByYou" {
            // Bucket already exists, continue
        } else {
            os.Exit(1)
        }
    }

    // Enable versioning for the bucket
    _, err = svc.PutBucketVersioning(&s3.PutBucketVersioningInput{
        Bucket: aws.String(bucketName),
        VersioningConfiguration: &s3.VersioningConfiguration{
            Status: aws.String("Enabled"),
        },
    })
    if err != nil {
        fmt.Println("Error enabling bucket versioning:", err)
        os.Exit(1)
    }

    // Attach the policy to the bucket
    _, err = svc.PutBucketPolicy(&s3.PutBucketPolicyInput{
        Bucket: aws.String(bucketName),
        Policy: aws.String(policy),
    })
    if err != nil {
        fmt.Println("Error attaching policy to bucket:", err)
        os.Exit(1)
    }

    // Create an AWS Key Management Service client
    kmsSvc := kms.New(sess)

    // Create a new KMS key
    keyAlias := "alias/my-kms-key/sdk"
    keyResp, err := kmsSvc.CreateKey(&kms.CreateKeyInput{
        Description: aws.String("My KMS Key"),
    })
    if err != nil {
        fmt.Println("Error creating KMS key:", err)
        os.Exit(1)
    }

    // Attach the KMS key to the bucket
    _, err = svc.PutBucketEncryption(&s3.PutBucketEncryptionInput{
        Bucket: aws.String(bucketName),
        ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{
            Rules: []*s3.ServerSideEncryptionRule{
                {
                    ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
                        KMSMasterKeyID: keyResp.KeyMetadata.Arn,
                        SSEAlgorithm:   aws.String("aws:kms"),
                    },
                },
            },
        },
    })
    if err != nil {
        fmt.Println("Error attaching KMS key to bucket:", err)
        os.Exit(1)
    }

    // Print out the bucket name and KMS key alias
    fmt.Printf("Bucket name: %s\n", bucketName)
    fmt.Printf("KMS key alias: %s\n", keyAlias)
}
