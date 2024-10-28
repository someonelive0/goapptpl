package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func AddMinioHandler(app *fiber.App) error {
	println("0 ---------------->")

	minio_api := app.Group("/minio")

	minio_api.Get("/buckets", bucketsHandler)
	minio_api.Get("/bucket/:bucket/objects", objectsHandler)
	minio_api.Get("/bucket/:bucket/object-meta/:object", objectInfoHandler)
	minio_api.Get("/bucket/:bucket/object/:object", objectHandler)
	// minio_api.Head("/bucket/:bucket/object/:object", objectInfoHander)

	return nil
}

func bucketsHandler(c fiber.Ctx) error {
	log.Printf("/minio/buckets\n")
	minioClient, err := getMinioClient()
	if err != nil {
		log.Println(err)
	}

	// log.Printf("minioClient %#v\n", minioClient) // minioClient is now setup

	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		log.Printf("ListBuckets failed %s\n", err)
	}
	// log.Printf("buckets %#v\n", buckets) // minioClient is now setup

	c.Context().SetContentType("application/json")
	data, _ := json.MarshalIndent(buckets, "", "  ")
	return c.Send(data)
}

func objectsHandler(c fiber.Ctx) error {
	log.Printf("/minio/bucket/:bucket/objects\n")
	minioClient, err := getMinioClient()
	if err != nil {
		log.Println(err)
	}

	// log.Printf("minioClient %#v\n", minioClient) // minioClient is now setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := minioClient.ListObjects(ctx, c.Params("bucket"),
		minio.ListObjectsOptions{
			Prefix:    "",
			Recursive: true,
		})

	c.Context().SetContentType("application/json")
	c.WriteString("[")
	i := 0
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return c.SendString(object.Err.Error())
		}
		if i > 0 {
			c.WriteString(`, `)
		}
		log.Println(object.Key)
		c.Writef(`"%s" `, strings.Replace(object.Key, `"`, `\"`, -1))
		i++
	}
	c.WriteString(` ]`)

	return nil
}

// GET /minio/bucket/:bucket/object/:object?mime=json|xml
func objectInfoHandler(c fiber.Ctx) error {
	log.Printf("/minio/bucket/:bucket/object/:object\n")
	bucket, _ := url.QueryUnescape(c.Params("bucket"))
	object, _ := url.QueryUnescape(c.Params("object"))
	// q := c.Queries()
	// mime := "json"
	// if len(q["mime"]) > 0 {
	// 	mime = q["mime"]
	// }
	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json

	minioClient, err := getMinioClient()
	if err != nil {
		log.Println(err)
	}

	// log.Printf("minioClient %#v\n", minioClient) // minioClient is now setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectInfo, err := minioClient.StatObject(ctx, bucket, object,
		minio.StatObjectOptions{})
	if err != nil {
		c.Writef("bucket: '%s', object: '%s', ", bucket, object)
		c.WriteString(err.Error())
		return nil
	}

	if mime == "json" {
		c.Context().SetContentType("application/json")
		data, _ := json.MarshalIndent(objectInfo, "", "  ")
		return c.Send(data)
	} else if mime == "xml" {
		data, err := xml.MarshalIndent(objectInfo, "", "  ")
		if err != nil {
			// xml: unsupported type: minio.StringMap
			return c.SendString(err.Error())
		}
		return c.Send(data)
	} else {
		return c.SendString(fmt.Sprintf("mime '%s' not supported", mime))
	}
}

func objectHandler(c fiber.Ctx) error {
	log.Printf("/minio/bucket/:bucket/object/:object\n")
	bucket, _ := url.QueryUnescape(c.Params("bucket"))
	object, _ := url.QueryUnescape(c.Params("object"))
	// object := c.Params("object")

	minioClient, err := getMinioClient()
	if err != nil {
		log.Println(err)
	}

	// log.Printf("minioClient %#v\n", minioClient) // minioClient is now setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	object_fd, err := minioClient.GetObject(ctx, bucket, object,
		minio.GetObjectOptions{})
	if err != nil {
		c.Writef("bucket: '%s', object: '%s', ", bucket, object)
		c.WriteString(err.Error())
		return nil
	}
	defer object_fd.Close()

	// set response header of Attachment
	// => Content-Disposition: attachment; filename=object
	// => Content-Type: image/png
	c.Attachment(object)
	_, err = io.Copy(c, object_fd)
	return err
}

func getMinioClient() (*minio.Client, error) {
	endpoint := "172.16.0.243:9098"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false
	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Println(err)
	}

	return minioClient, err
}

// func AddMinioHandler1(r fiber.Router) error {
// 	println("0 ---------------->")

// 	// minio_api := app.Group("/minio", func(c fiber.Ctx) error {
// 	// 	println("1 ---------------->")
// 	// 	return c.SendString("Minio ")
// 	// })

// 	r.Get("/buckets", func(c fiber.Ctx) error {
// 		println("2 -------- --------  ===>")
// 	})

// 	return nil
// }
