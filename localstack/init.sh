#!/bin/sh

export AWS_REGION=ap-northeast-1
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy

# バケット作成
awslocal s3api create-bucket --bucket test-bucket

# お試しファイルを作成してアップロード
echo "Hello, LocalStack!" > /tmp/test.txt
awslocal s3api put-object --bucket test-bucket --key test.txt --body /tmp/test.txt

# Lambda関数の作成
awslocal lambda create-function \
  --region $AWS_REGION \
  --function-name presign-function \
  --runtime provided.al2023 \
  --role arn:aws:iam::000000000000:role/lambda-role \
  --handler bootstrap \
  --zip-file fileb:///lambda/lambda.zip \
  --environment "Variables={BUCKET_NAME=test-bucket,AWS_REGION=$AWS_REGION}"
