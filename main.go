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
	database      string
	httpAddr      string
	hostname      string
	projectID     string
	tlsCert       string
	tlsKey        string
	traceClient   *trace.Client
	spannerClient *spanner.Client
	storageClient *storage.Client
)

var serviceAccountFile = "/var/run/secret/cloud.google.com/service-account.json"

func main() {
	flag.StringVar(&database, "database", "", "The Spanner database.")
	flag.StringVar(&httpAddr, "http", ":443", "HTTP Listen address.")
	flag.StringVar(&projectID, "project-id", "", "Google Cloud project id.")
	flag.StringVar(&tlsCert, "tls-cert", "/etc/tls/server.pem", "TLS certificate path")
	flag.StringVar(&tlsKey, "tls-key", "/etc/tls/server.key", "TLS private key path")
	flag.Parse()

	log.Println("Initializing application...")

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	traceClient, err = trace.NewClient(ctx, projectID,
		option.WithServiceAccountFile(serviceAccountFile))
	if err != nil {
		log.Fatal(err)
	}

	storageClient, err = storage.NewClient(ctx,
		option.WithServiceAccountFile(serviceAccountFile))
	if err != nil {
		log.Fatal(err)
	}

	spannerClient, err = spanner.NewClient(ctx, database,
		option.WithServiceAccountFile(serviceAccountFile))

	http.HandleFunc("/", httpHandler)

	server := http.Server{}

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenAndServeTLS(tlsCert, tlsKey)
	}()

	log.Printf("HTTPS listener on %s...", httpAddr)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	shutdownCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	for {
		select {
		case s := <-signalChan:
			log.Println(fmt.Sprintf("Captured %v. Exiting...", s))
			server.Shutdown(shutdownCtx)
		case err := <-errChan:
			log.Fatal(err)
		case <-shutdownCtx.Done():
			log.Println(shutdownCtx.Err())
		}
	}
}
