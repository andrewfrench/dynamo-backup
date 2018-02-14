# dynamo-backup

An AWS Lambda-based backup tool using DynamoDB's new backup utility.

## build

Build the tool for Linux so it can be executed by Lambda.

```sh
GOOS=linux go build handler.go
```

## use 

Create a Lambda function to back up your Dynamo tables:

```sh
aws lambda create-function --region $YOUR_REGION --function-name $YOUR_FUNCTION_NAME --memory 128 --role $YOUR_ROLE_ARN --runtime go1.x --zip-file fileb://handler.zip --handler handler
```

Invoke the new Lambda manually or create a CloudWatch event to invoke the Lambda automatically.

