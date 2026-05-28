package test

import (
	"bytes"
	grpcMedia "cinema/gen/media"
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestInitVideoUpload(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run")
	}

	// 1. Подключаемся к сервису
	conn, err := grpc.NewClient("localhost:44046",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := grpcMedia.NewContentClient(conn)
	ctx := context.Background()

	fileName := "sample.mkv"
	contentType := "video/x-matroska"

	// 2. Читаем тестовый файл
	filePath := "testdata/" + fileName
	fileData, err := os.ReadFile(filePath)
	require.NoError(t, err, "положи тестовый файл в "+filePath)

	filmId := "7321ca6d-d20a-401a-b381-b7ee553f656f"
	fileSize := int64(len(fileData))

	// 3. Вызываем InitUpload
	resp, err := client.InitUpload(ctx, &grpcMedia.InitUploadRequest{
		FilmId:      filmId,
		FileName:    fileName,
		ContentType: contentType,
		FileSize:    fileSize,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.UploadId)
	require.NotEmpty(t, resp.PresignedUrls)
	require.Positive(t, resp.ChunkSize)

	t.Logf("upload_id:   %s", resp.UploadId)
	t.Logf("chunk_size:  %d bytes", resp.ChunkSize)
	t.Logf("parts:       %d", len(resp.PresignedUrls))

	//_, _ = client.AbortUpload(ctx, &grpcMedia.AbortUploadRequest{
	//	UploadId: resp.UploadId,
	//	Key:      resp.Key,
	//})

	// 4. Загружаем части напрямую в S3 по presigned URLs
	httpClient := &http.Client{}
	etags := make([]string, len(resp.PresignedUrls))

	for i, url := range resp.PresignedUrls {
		start := int64(i) * resp.ChunkSize
		end := start + resp.ChunkSize
		if end > fileSize {
			end = fileSize
		}
		chunk := fileData[start:end]

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(chunk))
		require.NoError(t, err)
		req.ContentLength = int64(len(chunk))
		req.Header.Set("Content-Type", contentType)

		putResp, err := httpClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, putResp.StatusCode, "part %d upload failed", i+1)
		putResp.Body.Close()

		etag := putResp.Header.Get("ETag")
		require.NotEmpty(t, etag, "part %d: ETag is empty", i+1)
		etags[i] = etag

		t.Logf("part %d uploaded, ETag: %s", i+1, etag)
	}

	t.Logf("all parts uploaded, etags: %v", etags)

	// 5. Завершаем загрузку
	_, err = client.CompleteUpload(ctx, &grpcMedia.CompleteUploadRequest{
		UploadId: resp.UploadId,
		Key:      resp.Key,
		ETags:    etags,
	})
	require.NoError(t, err)

	t.Log("upload completed successfully")
}
