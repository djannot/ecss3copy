package main

import (

  "log"
  "strings"
  "sync"
  "sync/atomic"
  "time"
  "github.com/djannot/ecss3copy/s3"
  "github.com/jessevdk/go-flags"
  "github.com/mitchellh/goamz/aws"
)

const retries = 3
var c = make(chan KeysToSend)
var s3Client *s3.S3
var ops uint64 = 0
var succeeded uint64 = 0
var failed uint64 = 0

type KeysToSend struct {
  Keys []s3.Key
  Operation string
  Options interface{}
}

type CopyBucketOptions struct {
  SourceBucket string
  TargetBucket string
  Query string
}

var opts struct {
    EndPoint string `short:"e" long:"endpoint" description:"The ECS endpoint" required:"true"`
    ObjectUser string `short:"u" long:"user" description:"The ECS object user" required:"true"`
    Password string `short:"p" long:"password" description:"The ECS object user password" required:"true"`
    SourceBucket string `short:"s" long:"source" description:"The ECS source bucket" required:"true"`
    TargetBucket string `short:"t" long:"target" description:"The ECS target bucket" required:"true"`
    MaxKeys int `short:"m" long:"maxkeys" description:"The number of keys to retrieve simultaneously from the ECS source bucket" default:"100"`
    MetadataSearchQuery string `short:"q" long:"query" description:"The ECS metadata search query to select the objects from the source bucket"`
    Verbose bool `short:"v" long:"verbose" description:"Verbose mode also display the object successfully copies"`
}

func main() {
  _, err := flags.Parse(&opts)
    if err != nil {
      log.Fatal(err)
  }

  s3Auth := aws.Auth{
    AccessKey: opts.ObjectUser,
    SecretKey: opts.Password,
  }

  s3SpecialRegion := aws.Region{
    Name: "Special",
    S3Endpoint: opts.EndPoint,
  }

  s3Client = s3.New(s3Auth, s3SpecialRegion)

  copyBucketOptions := CopyBucketOptions{
    SourceBucket: opts.SourceBucket,
    TargetBucket: opts.TargetBucket,
    Query: opts.MetadataSearchQuery,
  }
  startTime := time.Now()
  copyBucket(copyBucketOptions)
  duration := time.Since(startTime)
  log.Printf("%d operations executed in %f seconds", ops, duration.Seconds())
  log.Printf("%s operations per second", float64(ops) / duration.Seconds())
  log.Printf("%s operations succeeded", succeeded)
  log.Printf("%s operations failed", failed)
}

func listObjects(wg *sync.WaitGroup, c chan KeysToSend, sourceBucket string, operation string, marker string, options interface{}) {
  s3Bucket := s3Client.Bucket(sourceBucket)
  listResp, err := s3Bucket.List("", "", marker, opts.MaxKeys)
  if(err != nil) {
    log.Fatal(err)
  }
  lastKey := ""
  keys := []s3.Key{}
  for _, key := range listResp.Contents {
    lastKey = key.Key
    keys = append(keys, key)
  }

  if(len(keys) > 0) {
    keysToSend := KeysToSend{
      Keys: keys,
      Operation: operation,
      Options: options,
    }
    c <- keysToSend
  }

  wg.Wait()

  if(listResp.IsTruncated) {
    listObjects(wg, c, sourceBucket, operation, lastKey, options)
  }
}

func queryObjects(wg *sync.WaitGroup, c chan KeysToSend, sourceBucket string, query string, operation string, marker string, options interface{}) {
  s3Bucket := s3Client.Bucket(sourceBucket)
  queryResp, err := s3Bucket.Query(query, marker, opts.MaxKeys)

  if(err != nil) {
    log.Fatal(err)
  }

  keys := []s3.Key{}
  for _, item := range queryResp.EntryLists {
    key := s3.Key{
      Key: item.ObjectName,
    }
    keys = append(keys, key)
  }

  if(len(keys) > 0) {
    //wg.Add(1)
    keysToSend := KeysToSend{
      Keys: keys,
      Operation: operation,
      Options: options,
    }
    c <- keysToSend
  }

  wg.Wait()
  if(queryResp.NextMarker != "NO MORE PAGES") {
    queryObjects(wg, c, sourceBucket, query, operation, queryResp.NextMarker, options)
  }
}

func copyBucket(copyBucketOptions CopyBucketOptions) {
  c := make(chan KeysToSend)
  var wg sync.WaitGroup

  go bucketWorker(&wg, c)
  if copyBucketOptions.Query == "" {
    listObjects(&wg, c,  copyBucketOptions.SourceBucket, "CopyObject", "", copyBucketOptions)
  } else {
    queryObjects(&wg, c,  copyBucketOptions.SourceBucket, strings.Replace(copyBucketOptions.Query, " ", "%20", -1), "CopyObject", "", copyBucketOptions)
  }
}

func bucketWorker(wg *sync.WaitGroup, c chan KeysToSend) {
  for {
    keysToSend := <- c
    for _, key := range keysToSend.Keys {
      if(keysToSend.Operation == "CopyObject") {
        wg.Add(1)
        go copyObject(wg, key, keysToSend.Options.(CopyBucketOptions), s3.PublicRead, "REPLACE")
      }
    }
  }
}

func copyObject(wg *sync.WaitGroup, key s3.Key, copyBucketOptions CopyBucketOptions, perm s3.ACL, directive string) {
  s3Bucket := s3Client.Bucket(copyBucketOptions.TargetBucket)
  /*
  Could be implemented to delete the objects in the source bucket
  err := s3Bucket.Del(key.Key)
  if(err != nil) {
    log.Print(err)
  }
  */
  atomic.AddUint64(&ops, 1)
  tried := 0
  for {
    err := s3Bucket.CopyToNewBucket(key.Key, key.Key, copyBucketOptions.SourceBucket, perm, directive)
    if(err != nil) {
      log.Print(err)
      tried++
    } else {
      atomic.AddUint64(&succeeded, 1)
      if opts.Verbose {
        log.Printf("Object %s has been copied from %s to %s", key.Key, copyBucketOptions.SourceBucket, copyBucketOptions.TargetBucket)
      }
      break
    }
    if tried >= retries {
      atomic.AddUint64(&failed, 1)
      log.Printf("Object %s hasn't been copied from %s to %s", key.Key, copyBucketOptions.SourceBucket, copyBucketOptions.TargetBucket)
    }
  }
  wg.Done()
}
