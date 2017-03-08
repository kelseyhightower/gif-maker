// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/spanner"
	"cloud.google.com/go/storage"
	"cloud.google.com/go/trace"
	"google.golang.org/api/option"
)

var (
	bucket        string
	database      string
	httpAddr      string
	hostname      string
	projectID     string
	traceClient   *trace.Client
	spannerClient *spanner.Client
	storageClient *storage.Client
)

var serviceAccountFile = "/var/run/secret/cloud.google.com/service-account.json"

func main() {
	flag.StringVar(&bucket, "bucket", "", "The Google Cloud Storage bucket to storage images.")
	flag.StringVar(&database, "database", "", "The Spanner database to store events.")
	flag.StringVar(&httpAddr, "http", "0.0.0.0:80", "The HTTP listen address.")
	flag.StringVar(&projectID, "project-id", "", "The Google Cloud project id.")
	flag.Parse()

	log.Println("Starting gif-maker service...")

	var err error
	ctx := context.Background()

	traceClient, err = trace.NewClient(ctx, projectID,
		option.WithServiceAccountFile(serviceAccountFile))
	if err != nil {
		log.Fatal(err)
	}

	p, err := trace.NewLimitedSampler(1, 10)
	if err != nil {
		log.Fatal(err)
	}
	traceClient.SetSamplingPolicy(p)

	storageClient, err = storage.NewClient(ctx,
		option.WithServiceAccountFile(serviceAccountFile))
	if err != nil {
		log.Fatal(err)
	}

	spannerClient, err = spanner.NewClient(ctx, database,
		option.WithServiceAccountFile(serviceAccountFile))

	http.HandleFunc("/", httpHandler)
	http.HandleFunc("/healthz", healthHandler)

	server := http.Server{
		Addr: httpAddr,
	}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	log.Printf("HTTP listener on %s...", httpAddr)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	s := <-signalChan
	log.Println(fmt.Sprintf("Captured %v. Exiting...", s))
	shutdownCtx, _ := context.WithTimeout(context.Background(), 120*time.Second)
	server.Shutdown(shutdownCtx)

	<-shutdownCtx.Done()
	log.Println(shutdownCtx.Err())
	os.Exit(0)
}
