package xcalistores3

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	drawingStoreBucketName        = "drawing-store"
	credentialsObjectKey          = "credentials"
	sessionsObjectKeyPrefix       = "sessions"
	drawingContentObjectKeyPrefix = "drawing-content"
)

type DrawingStore struct {
	s3Client   *s3.Client
	bucketName string
}

func (store *DrawingStore) GetAllowedCredentials(ctx context.Context) (string, error) {
	input := s3.GetObjectInput{
		Bucket: &store.bucketName,
		Key:    aws.String(credentialsObjectKey),
	}
	response, getErr := store.s3Client.GetObject(ctx, &input)
	if getErr != nil {
		return "", fmt.Errorf("failed to retrieve S3 Object for credentials: %w", getErr)
	}
	content, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read credentials from S3: %w", readErr)
	}
	return string(content), nil
}

func (store *DrawingStore) CreateSession(ctx context.Context) (string, error) {
	deleteErr := store.deleteAllSessions(ctx)
	if deleteErr != nil {
		return "", fmt.Errorf("failed to delete existing sessions while creating a new one: %w", deleteErr)
	}

	body := strings.NewReader("empty")
	sId := SessionId()
	key := fmt.Sprintf("%s/%s", sessionsObjectKeyPrefix, sId)

	input := s3.PutObjectInput{
		Bucket: &store.bucketName,
		Key:    &key,
		Body:   body,
	}
	_, err := store.s3Client.PutObject(ctx, &input)
	if err != nil {
		return "", fmt.Errorf("failed to put object for session %s: %w", key, err)
	}

	return sId, nil
}

func (store *DrawingStore) ListSessions(ctx context.Context) ([]string, error) {
	return store.listObjectKeys(ctx, sessionsObjectKeyPrefix, true)
}

func (store *DrawingStore) deleteAllSessions(ctx context.Context) error {
	keys, listErr := store.listObjectKeys(ctx, sessionsObjectKeyPrefix, false)
	if listErr != nil {
		return fmt.Errorf("failed to list session object keys (omitPrefix=false): %w", listErr)
	}

	if len(keys) == 0 {
		return nil
	}

	var objectIds []types.ObjectIdentifier
	for _, key := range keys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}
	_, deleteErr := store.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(store.bucketName),
		Delete: &types.Delete{Objects: objectIds, Quiet: aws.Bool(true)},
	})
	if deleteErr != nil {
		return fmt.Errorf("failed to delete session object keys: %w", deleteErr)
	}

	return nil
}

func (store *DrawingStore) PutDrawing(ctx context.Context, title string, content io.Reader) error {
	key := fmt.Sprintf("%s/%s", drawingContentObjectKeyPrefix, title)

	input := s3.PutObjectInput{
		Bucket: &store.bucketName,
		Key:    &key,
		Body:   content,
	}
	_, err := store.s3Client.PutObject(ctx, &input)
	if err != nil {
		return fmt.Errorf("failed to put object for drawing %s: %w", key, err)
	}

	return nil
}

func (store *DrawingStore) ListDrawingTitles(ctx context.Context) ([]string, error) {
	return store.listObjectKeys(ctx, drawingContentObjectKeyPrefix, true)
}

func (store *DrawingStore) GetDrawing(ctx context.Context, title string) (string, error) {
	key := fmt.Sprintf("%s/%s", drawingContentObjectKeyPrefix, title)
	input := s3.GetObjectInput{
		Bucket: &store.bucketName,
		Key:    &key,
	}
	output, getObjectErr := store.s3Client.GetObject(ctx, &input)
	if getObjectErr != nil {
		return "", fmt.Errorf("failed to get content object with title '%s': %w", title, getObjectErr)
	}
	content, readBodyErr := io.ReadAll(output.Body)
	if readBodyErr != nil {
		return "", fmt.Errorf("failed to read content body for drawing %s: %w", title, readBodyErr)
	}

	fmt.Printf(">>>>>>>>> content: %#v\n", content)

	return string(content), nil
}

func (store *DrawingStore) listObjectKeys(ctx context.Context, prefix string, omitPrefixFromOutput bool) ([]string, error) {
	var err error
	var output *s3.ListObjectsV2Output
	keys := []string{}
	input := s3.ListObjectsV2Input{
		Bucket: &store.bucketName,
		Prefix: &prefix,
	}
	objectPaginator := s3.NewListObjectsV2Paginator(store.s3Client, &input)
	for objectPaginator.HasMorePages() {
		output, err = objectPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		} else {
			for _, object := range output.Contents {
				keyToOutput := *object.Key
				if omitPrefixFromOutput {
					keyToOutput = string([]rune(*object.Key)[len(prefix)+1:])
				}
				keys = append(keys, keyToOutput)
			}
		}
	}
	return keys, err
}

func SessionId() string {
	buf := make([]byte, 32)

	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x", buf)
}

func NewStore(ctx context.Context, bucketName string) (*DrawingStore, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't load default configuration.")
		fmt.Println(err)
		return nil, fmt.Errorf("failed to load default configuration: %w", err)
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	bucketNameToUse := drawingStoreBucketName
	if len(bucketName) > 0 {
		bucketNameToUse = bucketName
	}
	return &DrawingStore{
		s3Client:   s3Client,
		bucketName: bucketNameToUse,
	}, nil
}
