package main
import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	//"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	//"github.com/aws/aws-sdk-go/service/codepipeline/codepipelineiface"
	//"github.com/aws/aws-sdk-go/service/codebuild"
)
const (
	githubOwner = "Saba1119"
	githubRepo  = "CRUDapp"
	branch      = "main"
	OAuthtoken   = "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"
)
func main() {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}
	// Create a new CodePipeline client
	codePipelineClient := codepipeline.New(sess)
	// Create a new CodeBuild client
	//codeBuildClient := codebuild.New(sess)
	// Create a new CodeDeploy client
	//codeDeployClient := codedeploy.New(sess)
	// Set up the source stage
	source := &codepipeline.StageDeclaration{
		Name: aws.String("Source"),
		Actions: []*codepipeline.ActionDeclaration{
			&codepipeline.ActionDeclaration{
				Name: aws.String("Source"),
				ActionTypeId: &codepipeline.ActionTypeId{
					Category: aws.String("Source"),
					Owner:    aws.String("ThirdParty"),
					Version:  aws.String("1"),
					Provider: aws.String("GitHub"),
				},
				Configuration: map[string]*string{
					"Owner": aws.String(githubOwner),
					"Repo":  aws.String(githubRepo),
					"Branch": aws.String(branch),
					"OAuthToken":aws.String(OAuthtoken),
				},
				OutputArtifacts: []*codepipeline.OutputArtifact{
					&codepipeline.OutputArtifact{
						Name: aws.String("SourceOutput"),
					},
				},
			},
		},
	}
	// Set up the build stage
	build := &codepipeline.StageDeclaration{
		Name: aws.String("Build"),
		Actions: []*codepipeline.ActionDeclaration{
			&codepipeline.ActionDeclaration{
				Name: aws.String("Build"),
				ActionTypeId: &codepipeline.ActionTypeId{
					Category: aws.String("Build"),
					Owner:    aws.String("AWS"),
					Version:  aws.String("1"),
					Provider: aws.String("CodeBuild"),
				},
				Configuration: map[string]*string{
					"ProjectName": aws.String("sdk-app"),
				},
				InputArtifacts: []*codepipeline.InputArtifact{
					&codepipeline.InputArtifact{
						Name: aws.String("SourceOutput"),
					},
				},
				OutputArtifacts: []*codepipeline.OutputArtifact{
					&codepipeline.OutputArtifact{
						Name: aws.String("BuildOutput"),
					},
				},
			},
		},
	}
	// Set up the deployment stage
	deploy := &codepipeline.StageDeclaration{
		Name: aws.String("Deploy"),
		Actions: []*codepipeline.ActionDeclaration{
			&codepipeline.ActionDeclaration{
				Name: aws.String("Deploy"),
				ActionTypeId: &codepipeline.ActionTypeId{
					Category: aws.String("Deploy"),
					Owner:    aws.String("AWS"),
					Version:  aws.String("1"),
					Provider: aws.String("CodeDeploy"),
				},
				Configuration: map[string]*string{
					"ApplicationName":     aws.String("sdk"),
                    "DeploymentGroupName": aws.String("sdk-test1"),
},
InputArtifacts: []*codepipeline.InputArtifact{
&codepipeline.InputArtifact{
Name: aws.String("BuildOutput"),
},
},
},
},
}
// Set up the S3 artifact store
artifactStore := &codepipeline.ArtifactStore{
    Type: aws.String("S3"),
    Location: aws.String("codepipeline-us-east-1-265635393449"),
 }   
// Create the pipeline
pipeline := &codepipeline.CreatePipelineInput{
Pipeline: &codepipeline.PipelineDeclaration{
Name: aws.String("testing"),
RoleArn: aws.String("arn:aws:iam::554248189203:role/AWSCodePipelineServiceRole-us-east-1-testing"),
ArtifactStore: artifactStore,
Stages: []*codepipeline.StageDeclaration{
source,
build,
deploy,
},
},
}
// Create the pipeline
pipelineOutput, err := codePipelineClient.CreatePipeline(pipeline)
if err != nil {
fmt.Println("Error creating pipeline:", err)
return
}
// Print the ARN of the newly created pipeline
fmt.Println("Created pipeline ARN:", *pipelineOutput)
}




