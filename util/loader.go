package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	ctrl "sigs.k8s.io/controller-runtime"
)

// // Scheme is the basic k8s scheme
// var Scheme *runtime.Scheme
//
// func init() {
// 	Scheme = scheme.Scheme //runtime.NewScheme()
// 	apiextensions.AddToScheme(Scheme)
// 	cmapi.AddToScheme(Scheme)
// }

const yamlSeparator = "\n---"
const separator = "---"

type ApplyObject struct {
	Raw     []byte
	Unstruc *unstructured.Unstructured
	Runtime runtime.Object
}

type ObjectLoader interface {
	LoadObjects() ([]ApplyObject, error)
}

type objectLoader struct {
	scheme *runtime.Scheme
	log    logr.Logger
}

func NewObjectLoader(scheme *runtime.Scheme) ObjectLoader {
	return &objectLoader{
		scheme: scheme,
		log:    ctrl.Log.WithName("loader"),
	}
}
func (o *objectLoader) LoadObjects() ([]ApplyObject, error) {

	objects := []ApplyObject{}
	manifests := Boxed.List()
	for _, m := range manifests {
		o.log.Info("Loading", "manifest", m)
		data := Boxed.Get(m)

		scanner := bufio.NewScanner(bytes.NewReader(data))
		buf := make([]byte, 8*1024)
		scanner.Buffer(buf, 512*1024)

		scanner.Split(splitYAMLDocument)

		for scanner.Scan() {
			decoder := serializer.NewCodecFactory(o.scheme, serializer.EnableStrict).UniversalDeserializer()
			a := ApplyObject{
				Raw: make([]byte, len(scanner.Bytes())),
			}
			copy(a.Raw, scanner.Bytes())
			obj, _ /* gvk */, err := decoder.Decode(a.Raw, nil, nil)
			if err != nil {
				return nil, err
			}
			a.Runtime = obj

			// convert the runtime.Object to unstructured.Unstructured
			unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(a.Runtime)
			if err != nil {
				return nil, err
			}
			a.Unstruc = &unstructured.Unstructured{
				Object: unstructuredData,
			}

			objects = append(objects, a)
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Invalid input: %s", err)
		}
	}
	return objects, nil

}

func (o *objectLoader) ReadObjects(filename string) ([]runtime.Object, error) {

	objects := []runtime.Object{}

	o.log.Info("Loading", "file", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(splitYAMLDocument)

	for scanner.Scan() {
		decoder := serializer.NewCodecFactory(o.scheme, serializer.EnableStrict).UniversalDeserializer()
		obj, _ /* gvk */, err := decoder.Decode(scanner.Bytes(), nil, nil)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Invalid input: %s", err)
	}
	return objects, nil
}

// FROM: https://github.com/kubernetes/apimachinery/blob/a98ff070d70e1d5c58428a86787e7a05a38cabe8/pkg/util/yaml/decoder.go#L142
// splitYAMLDocument is a bufio.SplitFunc for splitting YAML streams into individual documents.
func splitYAMLDocument(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	sep := len([]byte(yamlSeparator))
	if i := bytes.Index(data, []byte(yamlSeparator)); i >= 0 {
		// We have a potential document terminator
		i += sep
		after := data[i:]
		if len(after) == 0 {
			// we can't read any more characters
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
