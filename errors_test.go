package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"

	errorspb "github.com/nice20121222/protoc-gen-custom-errors/errors"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestGenerateFileProducesValidGo(t *testing.T) {
	source := generateTestSource(t, []string{"SYSTEM_ERROR"}, "message")
	if _, err := parser.ParseFile(token.NewFileSet(), "test_errors.pb.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source is invalid Go: %v\n%s", err, source)
	}
}

func TestGenerateFileQuotesDefaultMessage(t *testing.T) {
	want := "quoted \"text\" with \\ slash\nand newline"
	source := generateTestSource(t, []string{"SYSTEM_ERROR"}, want)
	file, err := parser.ParseFile(token.NewFileSet(), "test_errors.pb.go", source, parser.AllErrors)
	if err != nil {
		t.Fatalf("generated source is invalid Go: %v\n%s", err, source)
	}
	var got string
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) < 3 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "New" {
			return true
		}
		literal, ok := call.Args[2].(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		got, err = strconv.Unquote(literal.Value)
		if err != nil {
			t.Fatalf("unquote generated message: %v", err)
		}
		return false
	})
	if got != want {
		t.Fatalf("generated message = %q, want %q", got, want)
	}
}

func TestGenerateFileRejectsCamelNameCollision(t *testing.T) {
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("generateFile did not panic for colliding helper names")
		}
		message := fmt.Sprint(got)
		for _, want := range []string{"ErrorReason", "FOO_BAR", "FooBar", "collision"} {
			if !strings.Contains(message, want) {
				t.Fatalf("panic %q does not contain %q", message, want)
			}
		}
	}()
	generateTestSource(t, []string{"FOO_BAR", "FooBar"}, "message")
}

func TestGenerateFileRejectsCrossEnumCamelNameCollision(t *testing.T) {
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("generateFile did not panic for helper names colliding across enums")
		}
		message := fmt.Sprint(got)
		for _, want := range []string{"FirstReason", "SecondReason", "SystemError", "collision"} {
			if !strings.Contains(message, want) {
				t.Fatalf("panic %q does not contain %q", message, want)
			}
		}
	}()
	generateTestSourceForEnums(t, []testEnum{
		{name: "FirstReason", valueNames: []string{"SYSTEM_ERROR"}},
		{name: "SecondReason", valueNames: []string{"SystemError"}},
	}, "message")
}

type testEnum struct {
	name       string
	valueNames []string
}

func generateTestSource(t *testing.T, valueNames []string, message string) string {
	t.Helper()
	return generateTestSourceForEnums(t, []testEnum{{name: "ErrorReason", valueNames: valueNames}}, message)
}

func generateTestSourceForEnums(t *testing.T, enumSpecs []testEnum, message string) string {
	t.Helper()
	enums := make([]*descriptorpb.EnumDescriptorProto, 0, len(enumSpecs))
	for _, enumSpec := range enumSpecs {
		enumOptions := &descriptorpb.EnumOptions{}
		proto.SetExtension(enumOptions, errorspb.E_DefaultCode, int32(500))
		proto.SetExtension(enumOptions, errorspb.E_DefaultMessage, message)
		values := make([]*descriptorpb.EnumValueDescriptorProto, 0, len(enumSpec.valueNames))
		for i, name := range enumSpec.valueNames {
			values = append(values, &descriptorpb.EnumValueDescriptorProto{
				Name:   proto.String(name),
				Number: proto.Int32(int32(i)),
			})
		}
		enums = append(enums, &descriptorpb.EnumDescriptorProto{
			Name:    proto.String(enumSpec.name),
			Options: enumOptions,
			Value:   values,
		})
	}
	request := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{{
			Name:     proto.String("test.proto"),
			Package:  proto.String("test"),
			Syntax:   proto.String("proto3"),
			Options:  &descriptorpb.FileOptions{GoPackage: proto.String("example.com/test;testpb")},
			EnumType: enums,
		}},
	}
	gen, err := (protogen.Options{}).New(request)
	if err != nil {
		t.Fatalf("create generator: %v", err)
	}
	generated := generateFile(gen, gen.Files[0])
	content, err := generated.Content()
	if err != nil {
		t.Fatalf("render generated source: %v", err)
	}
	return string(content)
}

func Test_case2Camel(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "snake1",
			args: args{"SYSTEM_ERROR"},
			want: "SystemError",
		},
		{
			name: "snake2",
			args: args{"System_Error"},
			want: "SystemError",
		},
		{
			name: "snake3",
			args: args{"system_error"},
			want: "SystemError",
		},
		{
			name: "snake4",
			args: args{"System_error"},
			want: "SystemError",
		},
		{
			name: "upper1",
			args: args{"UNKNOWN"},
			want: "Unknown",
		},
		{
			name: "camel1",
			args: args{"SystemError"},
			want: "SystemError",
		},
		{
			name: "camel2",
			args: args{"systemError"},
			want: "SystemError",
		},
		{
			name: "lower1",
			args: args{"system"},
			want: "System",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := case2Camel(tt.args.name); got != tt.want {
				t.Errorf("case2Camel() = %v, want %v", got, tt.want)
			}
		})
	}
}
