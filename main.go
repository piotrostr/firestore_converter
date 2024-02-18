package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	dump = flag.Bool("dump", false, "Dump Firestore into dump.json")
	load = flag.Bool("load", false, "Populate Firestore from dump.json")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	opt := option.WithCredentialsFile(getenv("FIREBASE_SERVICE_ACCOUNT", ""))
	client, err := firestore.NewClientWithDatabase(
		ctx,
		getenv("FIREBASE_PROJECT_ID", ""),
		getenv("FIREBASE_DATABASE_ID", "(default)"),
		opt,
	)
	if err != nil {
		log.Fatalln(err)
	}

	if *dump {
		dumpFirestoreIntoJSON(ctx, client)
	}

	if *load {
		loadJSONIntoFirestore(ctx, client)
	}
}

func loadJSONIntoFirestore(ctx context.Context, client *firestore.Client) {
	defer client.Close()

	f, err := os.Open("dump.json")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	var data map[string]any
	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		log.Fatalln(err)
	}

	// create a staging database in Firestore, don't use default

	for colName, colData := range data {
		colRef := client.Collection(colName)
		for docName, docData := range colData.(map[string]any) {
			_, err := colRef.Doc(docName).Set(ctx, docData)
			if err != nil {
				log.Fatalln(err)
			}
		}

		log.Printf("Collection %s populated", colName)
	}

	log.Println("Firestore populated")
}

func dumpFirestoreIntoJSON(ctx context.Context, client *firestore.Client) {
	defer client.Close()

	log.Println("Dumping Firestore into dump.json")

	cols := client.Collections(ctx)
	allCols, err := cols.GetAll()
	if err != nil {
		log.Fatalln(err)
	}

	if len(allCols) == 0 {
		log.Println("No collections found")
		return
	}

	res := make(map[string]any)

	for len(allCols) > 0 {
		col := allCols[len(allCols)-1]
		allCols = allCols[:len(allCols)-1]

		colRes := make(map[string]any)
		docsIterator := col.Documents(ctx)
		for {
			doc, err := docsIterator.Next()
			if err == iterator.Done {
				break
			}
			colRes[doc.Ref.ID] = doc.Data()
			nestedCollectionsIterator := doc.Ref.Collections(ctx)
			allNestedCols, err := nestedCollectionsIterator.GetAll()
			if err != nil {
				log.Fatalln(err)
			}
			allCols = append(allCols, allNestedCols...)
		}
		res[col.ID] = colRes
	}

	dumpJSON(res)

	log.Println("Firestore dumped")
}

func dumpJSON(data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	f, err := os.Create("dump.json")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	_, err = f.Write(jsonData)
	if err != nil {
		log.Fatalln(err)
	}
}
