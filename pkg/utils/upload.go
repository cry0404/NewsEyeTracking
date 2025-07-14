package utils

import (
	"context"
	"encoding/json"
	
	"fmt"
	"log"
	"os"



	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/joho/godotenv"
)

func UploadCompressedData() {
	err := godotenv.Load()
	if err != nil {
		log.Println("未找到 .env 文件，加载失败")
	}

	accessKeyId := os.Getenv("ACCESS_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY")
	if accessKeyId == "" || accessKeySecret == "" {
		log.Println("请记得设置环境变量")
	}
	region := "cn-shenzhen"
	bucketName := "newseyetrackingtest"
	objectName := "https://oss-cn-shenzhen.aliyuncs.com"
	
	//region := "cn-shenzhen"
	
	provider := credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret)

	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).WithRegion(region)
	client := oss.NewClient(cfg)

	localFile := "./test.txt"
	//bucketName := ""
	//endpoint := ""

	// 创建上传对象的请求
	putRequest := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
		ProgressFn: func(increment, transferred, total int64) {
			fmt.Printf("increment:%v, transferred:%v, total:%v\n", increment, transferred, total)
		}, // 进度回调函数，用于显示上传进度
	}

	//这里显然需要处理可能超时的 context
	result, err := client.PutObjectFromFile(context.TODO(), putRequest, localFile)

	// 发送上传对象的请求

	if err != nil {
		log.Fatalf("failed to put object from file %v", err)
	}

	// 打印上传对象的结果（美化为 JSON）
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("failed to marshal result: %v", err)
	} else {
		fmt.Println(string(jsonResult))
	}
/*
	//测试下载
	downloadFile := "./downloaded.txt"
	getRequest := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
	}
	_, err = client.GetObjectToFile(context.TODO(), getRequest, downloadFile)
	if err != nil {
		log.Fatalf("failed to get object to file: %v", err)
	}
	log.Printf("object downloaded to: %s", downloadFile)
	*/
}