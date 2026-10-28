package s3

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/maximka-44js/lc/internal/storage/repo"
)

type S3Config struct {
	Region        string
	Endpoint      string
	Bucket        string
	AccessKey     string
	SecretKey     string
	UsePathStyle  bool
	PresignExpire time.Duration
}

type S3Storage struct {
	client        *s3.Client
	presign       *s3.PresignClient
	bucket        string
	presignExpiry time.Duration
}

func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("s3: region is required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3: bucket is required")
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("s3: access key and secret key are required")
	}

	loadOpts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("s3: load config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	presign := s3.NewPresignClient(client)

	expires := cfg.PresignExpire
	if expires <= 0 {
		expires = 15 * time.Minute
	}

	return &S3Storage{
		client:        client,
		presign:       presign,
		bucket:        cfg.Bucket,
		presignExpiry: expires,
	}, nil
}

func (s *S3Storage) PutObject(ctx context.Context, req repo.PutObjectRequest) error {
	if req.Body == nil {
		return fmt.Errorf("s3: body is required")
	}
	if req.Key == "" {
		return fmt.Errorf("s3: key is required")
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(req.Key),
		Body:        req.Body,
		ContentType: aws.String(req.MimeType),
	}
	if req.Size > 0 {
		input.ContentLength = aws.Int64(req.Size)
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return mapStorageErr(err)
	}

	return nil
}

func (s *S3Storage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("s3: key is required")
	}

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, mapStorageErr(err)
	}

	return result.Body, nil
}

func (s *S3Storage) Copy(ctx context.Context, sourceKey, destinationKey, mimeType string) error {
	sourceKey = strings.TrimSpace(sourceKey)
	destinationKey = strings.TrimSpace(destinationKey)
	if sourceKey == "" || destinationKey == "" {
		return fmt.Errorf("s3: source and destination keys are required")
	}

	copySource := s.bucket + "/" + encodeCopySourceKey(sourceKey)
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(destinationKey),
		CopySource: aws.String(copySource),
	}

	if strings.TrimSpace(mimeType) != "" {
		input.MetadataDirective = s3types.MetadataDirectiveReplace
		input.ContentType = aws.String(strings.TrimSpace(mimeType))
	}

	_, err := s.client.CopyObject(ctx, input)
	if err != nil {
		return mapStorageErr(err)
	}

	return nil
}

func (s *S3Storage) PresignPut(ctx context.Context, req repo.PresignPutRequest) (repo.PresignPutResult, error) {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(req.Key),
		ContentType:   aws.String(req.MimeType),
		ContentLength: aws.Int64(req.Size),
	}

	if req.ChecksumSHA256 != "" {
		if err := validateChecksumSHA256Base64(req.ChecksumSHA256); err != nil {
			return repo.PresignPutResult{}, err
		}
		input.ChecksumSHA256 = aws.String(req.ChecksumSHA256)
		input.ChecksumAlgorithm = s3types.ChecksumAlgorithmSha256
	}

	res, err := s.presign.PresignPutObject(ctx, input, s3.WithPresignExpires(s.presignExpiry))
	if err != nil {
		return repo.PresignPutResult{}, mapStorageErr(err)
	}

	return repo.PresignPutResult{URL: res.URL}, nil
}

func (s *S3Storage) CreateMultipart(ctx context.Context, req repo.CreateMultipartRequest) (repo.CreateMultipartResult, error) {
	input := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(req.Key),
		ContentType: aws.String(req.MimeType),
	}

	res, err := s.client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return repo.CreateMultipartResult{}, mapStorageErr(err)
	}

	return repo.CreateMultipartResult{UploadID: aws.ToString(res.UploadId)}, nil
}

func (s *S3Storage) PresignUploadPart(ctx context.Context, req repo.PresignUploadPartRequest) (repo.PresignUploadPartResult, error) {
	input := &s3.UploadPartInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(req.Key),
		UploadId:      aws.String(req.UploadID),
		PartNumber:    aws.Int32(int32(req.PartNumber)),
		ContentLength: aws.Int64(req.Size),
	}

	if req.ChecksumSHA256 != "" {
		if err := validateChecksumSHA256Base64(req.ChecksumSHA256); err != nil {
			return repo.PresignUploadPartResult{}, err
		}
		input.ChecksumSHA256 = aws.String(req.ChecksumSHA256)
		input.ChecksumAlgorithm = s3types.ChecksumAlgorithmSha256
	}

	res, err := s.presign.PresignUploadPart(ctx, input, s3.WithPresignExpires(s.presignExpiry))
	if err != nil {
		return repo.PresignUploadPartResult{}, mapStorageErr(err)
	}

	return repo.PresignUploadPartResult{URL: res.URL}, nil
}

func (s *S3Storage) CompleteMultipart(ctx context.Context, req repo.CompleteMultipartRequest) error {
	parts := make([]repo.CompletedPart, 0, len(req.CompletedParts))
	parts = append(parts, req.CompletedParts...)

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	completed := make([]s3types.CompletedPart, 0, len(parts))
	for _, part := range parts {
		completed = append(completed, s3types.CompletedPart{
			ETag:       aws.String(part.ETag),
			PartNumber: aws.Int32(int32(part.PartNumber)),
		})
	}

	input := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: completed,
		},
	}

	_, err := s.client.CompleteMultipartUpload(ctx, input)
	if err != nil {
		return mapStorageErr(err)
	}

	return nil
}

func (s *S3Storage) AbortMultipart(ctx context.Context, key, uploadID string) error {
	input := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	}

	_, err := s.client.AbortMultipartUpload(ctx, input)
	if err != nil {
		return mapStorageErr(err)
	}

	return nil
}

func (s *S3Storage) ListMultipartUploads(ctx context.Context, req repo.ListMultipartUploadsRequest) ([]repo.MultipartUpload, error) {
	var (
		markerKey      *string
		markerUploadID *string
		result         []repo.MultipartUpload
	)

	for {
		input := &s3.ListMultipartUploadsInput{
			Bucket:         aws.String(s.bucket),
			Prefix:         aws.String(req.Prefix),
			KeyMarker:      markerKey,
			UploadIdMarker: markerUploadID,
		}

		res, err := s.client.ListMultipartUploads(ctx, input)
		if err != nil {
			return nil, mapStorageErr(err)
		}

		for _, upload := range res.Uploads {
			if upload.Initiated == nil {
				continue
			}
			if !upload.Initiated.Before(req.OlderThan) {
				continue
			}
			result = append(result, repo.MultipartUpload{
				Key:         aws.ToString(upload.Key),
				UploadID:    aws.ToString(upload.UploadId),
				InitiatedAt: aws.ToTime(upload.Initiated),
			})
		}

		if !aws.ToBool(res.IsTruncated) {
			break
		}
		markerKey = res.NextKeyMarker
		markerUploadID = res.NextUploadIdMarker
	}

	return result, nil
}

func (s *S3Storage) Head(ctx context.Context, key string) (repo.ObjectHead, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	res, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return repo.ObjectHead{}, mapStorageErr(err)
	}

	checksum := ""
	if res.ChecksumSHA256 != nil {
		decoded, err := base64.StdEncoding.DecodeString(aws.ToString(res.ChecksumSHA256))
		if err == nil && len(decoded) == sha256.Size {
			checksum = hex.EncodeToString(decoded)
		}
	}

	return repo.ObjectHead{
		Size:         aws.ToInt64(res.ContentLength),
		ETag:         aws.ToString(res.ETag),
		LastModified: aws.ToTime(res.LastModified),
		ContentType:  aws.ToString(res.ContentType),
		Checksum:     checksum,
	}, nil
}

func (s *S3Storage) PresignGet(ctx context.Context, key string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	res, err := s.presign.PresignGetObject(ctx, input, s3.WithPresignExpires(s.presignExpiry))
	if err != nil {
		return "", mapStorageErr(err)
	}

	return res.URL, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return mapStorageErr(err)
	}

	return nil
}

// DeleteByPrefix removes all objects with the provided prefix.
func (s *S3Storage) DeleteByPrefix(ctx context.Context, prefix string) error {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}

	paginator := s3.NewListObjectsV2Paginator(s.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return mapStorageErr(err)
		}
		if len(page.Contents) == 0 {
			continue
		}

		ids := make([]s3types.ObjectIdentifier, 0, len(page.Contents))
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			ids = append(ids, s3types.ObjectIdentifier{Key: obj.Key})
		}
		if len(ids) == 0 {
			continue
		}

		deleteInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &s3types.Delete{Objects: ids, Quiet: aws.Bool(true)},
		}
		res, err := s.client.DeleteObjects(ctx, deleteInput)
		if err != nil {
			return mapStorageErr(err)
		}
		if len(res.Errors) > 0 {
			errInfo := res.Errors[0]
			return fmt.Errorf("s3: delete objects: %s: %s", aws.ToString(errInfo.Key), aws.ToString(errInfo.Message))
		}
	}

	return nil
}

func mapStorageErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return wrapStorageErr(repo.ErrStorageTemporary, err)
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		switch code {
		case "NotFound", "NoSuchKey", "NoSuchBucket":
			return wrapStorageErr(repo.ErrStorageNotFound, err)
		case "NotImplemented":
			return wrapStorageErr(repo.ErrStorageNotSupported, err)
		}
		if apiErr.ErrorFault() == smithy.FaultServer {
			return wrapStorageErr(repo.ErrStorageTemporary, err)
		}
	}

	var respErr *smithyhttp.ResponseError
	if errors.As(err, &respErr) {
		status := respErr.HTTPStatusCode()
		if status >= 500 {
			return wrapStorageErr(repo.ErrStorageTemporary, err)
		}
		if status == 404 {
			return wrapStorageErr(repo.ErrStorageNotFound, err)
		}
		if status == 501 {
			return wrapStorageErr(repo.ErrStorageNotSupported, err)
		}
	}

	return err
}

func wrapStorageErr(kind error, err error) error {
	if err == nil {
		return kind
	}
	return fmt.Errorf("%w: %v", kind, err)
}

func validateChecksumSHA256Base64(value string) error {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return fmt.Errorf("s3: invalid sha256 checksum base64: %w", err)
	}
	if len(decoded) != sha256.Size {
		return fmt.Errorf("s3: invalid sha256 checksum length: %d", len(decoded))
	}
	return nil
}

func encodeCopySourceKey(key string) string {
	parts := strings.Split(key, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}
