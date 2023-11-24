// Generate boilerplate code for setting similar structs from each other

//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
)

// flags
var (
	outputFile = flag.String("o", "", "Output file name, stdout if unset")
)

// globals
var (
	out io.Writer = os.Stdout
)

// genSetFrom generates code to set the public members of a from b
//
// a and b should be pointers to structs
//
// a can be a different type from b
//
// Only the Fields which have the same name and assignable type on a
// and b will be set.
//
// This is useful for copying between almost identical structures that
// are frequently present in auto-generated code for cloud storage
// interfaces.
func genSetFrom(a, b interface{}) {
	name := fmt.Sprintf("setFrom_%T_%T", a, b)
	name = strings.Replace(name, ".", "", -1)
	name = strings.Replace(name, "*", "", -1)
	fmt.Fprintf(out, "\n// %s copies matching elements from a to b\n", name)
	fmt.Fprintf(out, "func %s(a %T, b %T) {\n", name, a, b)
	ta := reflect.TypeOf(a).Elem()
	tb := reflect.TypeOf(b).Elem()
	va := reflect.ValueOf(a).Elem()
	vb := reflect.ValueOf(b).Elem()
	for i := 0; i < tb.NumField(); i++ {
		bField := vb.Field(i)
		tbField := tb.Field(i)
		name := tbField.Name
		aField := va.FieldByName(name)
		taField, found := ta.FieldByName(name)
		if found && aField.IsValid() && bField.IsValid() && aField.CanSet() && tbField.Type.AssignableTo(taField.Type) {
			fmt.Fprintf(out, "\ta.%s = b.%s\n", name, name)
		}
	}
	fmt.Fprintf(out, "}\n")
}

func main() {
	flag.Parse()

	if *outputFile != "" {
		fd, err := os.Create(*outputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			err := fd.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
		out = fd
	}

	fmt.Fprintf(out, `// Code generated by "go run gen_setfrom.go"; DO NOT EDIT.

package s3

import "github.com/aws/aws-sdk-go/service/s3"
`)

	genSetFrom(new(s3.ListObjectsInput), new(s3.ListObjectsV2Input))
	genSetFrom(new(s3.ListObjectsV2Output), new(s3.ListObjectsOutput))
	genSetFrom(new(s3.ListObjectVersionsInput), new(s3.ListObjectsV2Input))
	genSetFrom(new(s3.ObjectVersion), new(s3.DeleteMarkerEntry))
	genSetFrom(new(s3.ListObjectsV2Output), new(s3.ListObjectVersionsOutput))
	genSetFrom(new(s3.Object), new(s3.ObjectVersion))
	genSetFrom(new(s3.CreateMultipartUploadInput), new(s3.HeadObjectOutput))
	genSetFrom(new(s3.CreateMultipartUploadInput), new(s3.CopyObjectInput))
	genSetFrom(new(s3.UploadPartCopyInput), new(s3.CopyObjectInput))
	genSetFrom(new(s3.HeadObjectOutput), new(s3.GetObjectOutput))
	genSetFrom(new(s3.CreateMultipartUploadInput), new(s3.PutObjectInput))
	genSetFrom(new(s3.HeadObjectOutput), new(s3.PutObjectInput))
}
