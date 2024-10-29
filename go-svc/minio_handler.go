package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

type MinioHandler struct {
	Minioconfig *MinioConfig
	cli         *minio.Client
}

// r := app.Group("/minio")
func (p *MinioHandler) AddRouter(r fiber.Router) error {
	log.Info("MinioHandler AddRouter")

	r.Get("/buckets", p.bucketsHandler)
	r.Get("/bucket/:bucket/objects", p.objectsHandler)
	r.Get("/bucket/:bucket/object-meta/:object", p.objectInfoHandler)
	r.Get("/bucket/:bucket/object/:object", p.objectHandler)
	// r.Head("/bucket/:bucket/object/:object", objectInfoHander)

	return nil
}

// GET /minio/buckets
func (p *MinioHandler) bucketsHandler(c fiber.Ctx) error {
	log.Debug("/minio/buckets")
	if p.cli == nil {
		if err := p.getMinioClient(); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.Minioconfig.Timeout))
	defer cancel()

	// go func() {
	// 	<-ctx.Done()
	// 	log.Error("ListBuckets timeout, ", ctx.Err())
	// }()

	buckets, err := p.cli.ListBuckets(ctx)
	if err != nil {
		log.Errorf("ListBuckets failed: %v", err)
		return err
	}

	c.Context().SetContentType("application/json")
	data, _ := json.MarshalIndent(buckets, "", "  ")
	return c.Send(data)
}

// GET /minio/bucket/:bucket/objects
func (p *MinioHandler) objectsHandler(c fiber.Ctx) error {
	log.Debug("/minio/bucket/:bucket/objects")
	if p.cli == nil {
		if err := p.getMinioClient(); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.Minioconfig.Timeout))
	defer cancel()

	objectCh := p.cli.ListObjects(ctx, c.Params("bucket"),
		minio.ListObjectsOptions{
			Prefix:    "",
			Recursive: true,
		})

	c.Context().SetContentType("application/json")
	c.WriteString("[")
	i := 0
	for object := range objectCh {
		if object.Err != nil {
			log.Errorf("ListObjects failed: %v", object.Err)
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

// GET /minio/bucket/:bucket/object-meta/:object?mime=json|xml
func (p *MinioHandler) objectInfoHandler(c fiber.Ctx) error {
	log.Debug("/minio/bucket/:bucket/object/:object")
	bucket, _ := url.QueryUnescape(c.Params("bucket"))
	object, _ := url.QueryUnescape(c.Params("object"))
	// q := c.Queries()
	// mime := "json"
	// if len(q["mime"]) > 0 {
	// 	mime = q["mime"]
	// }
	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json

	if p.cli == nil {
		if err := p.getMinioClient(); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.Minioconfig.Timeout))
	defer cancel()

	objectInfo, err := p.cli.StatObject(ctx, bucket, object,
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

// GET /minio/bucket/:bucket/object/:object
func (p *MinioHandler) objectHandler(c fiber.Ctx) error {
	log.Debug("/minio/bucket/:bucket/object/:object")
	bucket, _ := url.QueryUnescape(c.Params("bucket"))
	object, _ := url.QueryUnescape(c.Params("object"))

	if p.cli == nil {
		if err := p.getMinioClient(); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.Minioconfig.Timeout))
	defer cancel()

	object_fd, err := p.cli.GetObject(ctx, bucket, object,
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

func (p *MinioHandler) getMinioClient() error {
	// Initialize minio client object.
	minioClient, err := minio.New(p.Minioconfig.Addr,
		&minio.Options{
			Creds:  credentials.NewStaticV4(p.Minioconfig.User, p.Minioconfig.Password, ""),
			Secure: p.Minioconfig.Ssl,
		})
	if err != nil {
		log.Errorf("connect minio '%s' failed: %v", p.Minioconfig.Addr, err)
		return err
	}

	// log.Debugf("minioClient %#v\n", minioClient) // minioClient is now setup
	p.cli = minioClient
	return nil
}
