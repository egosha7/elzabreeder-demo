package repository

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nfnt/resize"
	"go.uber.org/zap"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"
)

// S3Repository представляет интерфейс для работы с файлами в S3.
type S3Repository interface {
	PutInS3(id string, fileHeaders []*multipart.FileHeader, aspectRatio float64, width, height uint) ([]string, error)
	DeleteFromS3(url string) error
}

// S3Repo представляет репозиторий для работы с S3 Object Storage.
type S3Repo struct {
	logger   *zap.Logger
	S3Client *s3.Client
	bucket   string
}

func (r *S3Repo) PutInS3(id string, fileHeaders []*multipart.FileHeader, aspectRatio float64, width, height uint) ([]string, error) {
	println("Start")
	var urls []string
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Декодируем изображение
		img, format, err := image.Decode(file)
		if err != nil {
			return nil, err
		}

		// Обрезаем изображение до соотношения сторон 3:2
		croppedImg := cropToRatio(img, aspectRatio)

		// Изменяем размер обрезанного изображения до нужных размеров (опционально)
		resizedImg := resize.Resize(width, height, croppedImg, resize.Lanczos3)

		// Загружаем логотип
		logoFile, err := os.Open("cmd/static/logo.png")
		if err != nil {
			return nil, err
		}
		defer logoFile.Close()
		logo, err := png.Decode(logoFile)
		if err != nil {
			return nil, err
		}

		// Добавляем логотип
		finalImg := addWatermark(resizedImg, logo)

		// Кодируем изображение в буфер
		var buf bytes.Buffer
		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, finalImg, nil)
		case "png":
			err = png.Encode(&buf, finalImg)
		default:
			return nil, fmt.Errorf("unsupported image format")
		}
		if err != nil {
			return nil, err
		}

		// Генерируем уникальное имя файла
		filename := fmt.Sprintf(id+"_%d_%s", time.Now().Unix(), strings.TrimSpace(fileHeader.Filename))

		// Загружаем в S3
		_, err = r.S3Client.PutObject(
			context.TODO(), &s3.PutObjectInput{
				Bucket: aws.String(r.bucket),
				Key:    aws.String(filename),
				Body:   bytes.NewReader(buf.Bytes()),
				ACL:    "public-read",
			},
		)
		if err != nil {
			log.Printf(
				"Couldn't upload file %v to %v:%v. Here's why: %v\n",
				filename, r.bucket, filename, err,
			)
			return nil, err
		}

		url := fmt.Sprintf("https://%s.s3.cloud.ru/%s", r.bucket, filename)
		println(url)
		urls = append(urls, url)
	}

	return urls, nil
}

// DeleteFromS3 удаляет объект из S3 по указанному URL.
func (r *S3Repo) DeleteFromS3(url string) error {
	// Извлекаем имя файла из URL
	filename := path.Base(url)

	_, err := r.S3Client.DeleteObject(
		context.TODO(), &s3.DeleteObjectInput{
			Bucket: aws.String(r.bucket),
			Key:    aws.String(filename),
		},
	)
	if err != nil {
		log.Printf("Couldn't delete file %v from %v. Here's why: %v\n", filename, r.bucket, err)
		return err
	}

	return nil
}

// cropToRatio обрезает изображение до указанного соотношения сторон
func cropToRatio(img image.Image, aspectRatio float64) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var newWidth, newHeight int
	if float64(width)/float64(height) > aspectRatio {
		newWidth = int(float64(height) * aspectRatio)
		newHeight = height
	} else {
		newWidth = width
		newHeight = int(float64(width) / aspectRatio)
	}

	// Вычисляем точки обрезки
	x0 := (width - newWidth) / 2
	y0 := (height - newHeight) / 2

	// Создаем новое изображение с нужными размерами
	croppedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.Draw(croppedImg, croppedImg.Bounds(), img, image.Point{x0, y0}, draw.Src)

	return croppedImg
}

// addWatermark добавляет логотип в правый нижний угол изображения
func addWatermark(img, watermark image.Image) image.Image {
	bounds := img.Bounds()
	watermark = resize.Resize(
		uint(bounds.Dx()/6), 0, watermark, resize.Lanczos3,
	) // Логотип размером 1/10 ширины изображения

	offset := image.Pt(
		bounds.Dx()-watermark.Bounds().Dx()-20, bounds.Dy()-watermark.Bounds().Dy()-20,
	) // Отступ 10 пикселей от края
	b := img.Bounds()
	m := image.NewRGBA(b)
	draw.Draw(m, b, img, image.Point{}, draw.Src)
	draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.Point{}, draw.Over)

	return m
}
