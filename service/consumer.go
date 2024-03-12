package service

import (
	"github.com/adjust/rmq/v5"
	"github.com/tel4vn/fins-microservices/internal/queue"
)

const (
	DOWNLOAD_QUEUE = "download-queue"
	ITEM_QUEUE     = "item-queue"
	RESULT_QUEUE   = "result-queue"
	METADATA_QUEUE = "metadata-queue"
	SYNC_QUEUE     = "sync-queue"
	EXPORT_QUEUE   = "export-queue"
)

func InitConsumerServices() {
	queue.BatchInitQueues(
		queue.NewConsumerConfig(DOWNLOAD_QUEUE, ConsumeDownloadFile, 3),
	)
}

func ConsumeDownloadFile(delivery rmq.Delivery) {
	return
	// Unmarshall message payload from byte to model
	// message, err := queue.UnmarshallMessage[model.DownloadDocumentMessage](delivery)
	// if err != nil {
	// 	return
	// }

	// Create context with timeout
	// ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	// defer cancel()

	// authUser := message.Header.User
	// payload := message.Payload
	// fileName := payload.FileName
	// isUpload := true

	// defer func(ctx context.Context, authUser *model.AuthUser, zipName string) {
	// 	if isUpload {
	// 		_, err := common.UploadImageToStorage(ctx, fileName)
	// 		if err != nil {
	// 			log.Error(err)
	// 		}
	// 	}
	// }(ctx, authUser, fileName)
}
