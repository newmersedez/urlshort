package main

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeCommentGroup builds an *ast.CommentGroup with a single comment line.
func makeCommentGroup(text string) *ast.CommentGroup {
	return &ast.CommentGroup{
		List: []*ast.Comment{{Slash: token.NoPos, Text: text}},
	}
}

// writeSrc creates a Go source file in dir with the given content.
func writeSrc(t *testing.T, dir, filename, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644)
	require.NoError(t, err)
}

// readGen returns the content of reset.gen.go from dir.
func readGen(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, outputFile))
	require.NoError(t, err, "reset.gen.go was not created")
	return string(data)
}

func TestProcessDir_Primitives(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package foo

// generate:reset
type Counter struct {
	count   int
	name    string
	enabled bool
	ratio   float64
}
`)
	require.NoError(t, processDir(dir))

	gen := readGen(t, dir)
	assert.Contains(t, gen, "func (c *Counter) Reset()")
	assert.Contains(t, gen, "c.count = 0")
	assert.Contains(t, gen, `c.name = ""`)
	assert.Contains(t, gen, "c.enabled = false")
	assert.Contains(t, gen, "c.ratio = 0")
}

func TestProcessDir_SliceAndMap(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package foo

// generate:reset
type Container struct {
	items []string
	index map[string]int
}
`)
	require.NoError(t, processDir(dir))

	gen := readGen(t, dir)
	assert.Contains(t, gen, "c.items = c.items[:0]")
	assert.Contains(t, gen, "clear(c.index)")
}

func TestProcessDir_PointerToPrimitive(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package foo

// generate:reset
type Config struct {
	timeout *int
	label   *string
}
`)
	require.NoError(t, processDir(dir))

	gen := readGen(t, dir)
	assert.Contains(t, gen, "if c.timeout != nil")
	assert.Contains(t, gen, "*c.timeout = 0")
	assert.Contains(t, gen, "if c.label != nil")
	assert.Contains(t, gen, `*c.label = ""`)
}

func TestProcessDir_PointerToStruct(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package foo

// generate:reset
type Node struct {
	value int
	child *Node
}
`)
	require.NoError(t, processDir(dir))

	gen := readGen(t, dir)
	assert.Contains(t, gen, "n.value = 0")
	// Pointer-to-struct uses interface assertion, nil guard must be present.
	assert.Contains(t, gen, "any(n.child).(interface{ Reset() })")
	assert.Contains(t, gen, "n.child != nil")
}

func TestProcessDir_NonPointerStruct(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package foo

// generate:reset
type Outer struct {
	count int
	inner Inner
}

type Inner struct{}
`)
	require.NoError(t, processDir(dir))

	gen := readGen(t, dir)
	assert.Contains(t, gen, "o.count = 0")
	// Non-pointer struct uses &field so pointer-receiver Reset() is reachable.
	assert.Contains(t, gen, "any(&o.inner).(interface{ Reset() })")
}

func TestProcessDir_MultipleStructs(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package foo

// generate:reset
type A struct {
	x int
}

// generate:reset
type B struct {
	y string
}
`)
	require.NoError(t, processDir(dir))

	gen := readGen(t, dir)
	assert.Contains(t, gen, "func (a *A) Reset()")
	assert.Contains(t, gen, "func (b *B) Reset()")
}

func TestProcessDir_SkipsUnannotatedStructs(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package foo

type Plain struct {
	x int
}

// generate:reset
type Marked struct {
	y int
}
`)
	require.NoError(t, processDir(dir))

	gen := readGen(t, dir)
	assert.NotContains(t, gen, "func (p *Plain) Reset()")
	assert.Contains(t, gen, "func (m *Marked) Reset()")
}

func TestProcessDir_NoAnnotations_NoFileCreated(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package foo

type Plain struct {
	x int
}
`)
	require.NoError(t, processDir(dir))

	_, err := os.Stat(filepath.Join(dir, outputFile))
	assert.True(t, os.IsNotExist(err), "reset.gen.go should not be created when there are no annotated structs")
}

func TestProcessDir_GroupedTypeDeclaration(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package foo

type (
	// generate:reset
	Stats struct {
		hits int
		misses int
	}

	Other struct {
		x int
	}
)
`)
	require.NoError(t, processDir(dir))

	gen := readGen(t, dir)
	assert.Contains(t, gen, "func (s *Stats) Reset()")
	assert.NotContains(t, gen, "func (o *Other) Reset()")
}

func TestGeneratedFileHeader(t *testing.T) {
	dir := t.TempDir()
	writeSrc(t, dir, "types.go", `package mypackage

// generate:reset
type Thing struct {
	n int
}
`)
	require.NoError(t, processDir(dir))

	gen := readGen(t, dir)
	assert.True(t, strings.HasPrefix(gen, "// Code generated"), "file must start with generated marker")
	assert.Contains(t, gen, "package mypackage")
}

func TestRun_WalksSubdirectories(t *testing.T) {
	root := t.TempDir()
	subDir := filepath.Join(root, "sub")
	require.NoError(t, os.Mkdir(subDir, 0o755))

	writeSrc(t, subDir, "types.go", `package sub

// generate:reset
type Item struct {
	id int
}
`)
	require.NoError(t, run(root))

	gen := readGen(t, subDir)
	assert.Contains(t, gen, "func (i *Item) Reset()")
}

func TestRun_SkipsVendorAndTestdata(t *testing.T) {
	root := t.TempDir()
	for _, skip := range []string{"vendor", "testdata"} {
		d := filepath.Join(root, skip)
		require.NoError(t, os.Mkdir(d, 0o755))
		writeSrc(t, d, "types.go", `package skip

// generate:reset
type Skipped struct{ x int }
`)
	}

	require.NoError(t, run(root))

	for _, skip := range []string{"vendor", "testdata"} {
		_, err := os.Stat(filepath.Join(root, skip, outputFile))
		assert.True(t, os.IsNotExist(err), "%s should be skipped", skip)
	}
}

func TestHasGenerateComment(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"exact match", "// generate:reset", true},
		{"with spaces", "//  generate:reset", false},
		{"wrong keyword", "// generate:something", false},
		{"empty comment", "//", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := makeCommentGroup(tt.text)
			assert.Equal(t, tt.expected, hasGenerateComment(cg))
		})
	}
}

func TestPrimitiveZero(t *testing.T) {
	cases := map[string]string{
		"int": "0", "int8": "0", "int16": "0", "int32": "0", "int64": "0",
		"uint": "0", "uint8": "0", "uint16": "0", "uint32": "0", "uint64": "0",
		"uintptr": "0", "byte": "0", "rune": "0",
		"float32": "0", "float64": "0",
		"complex64": "0", "complex128": "0",
		"bool":   "false",
		"string": `""`,
	}
	for typeName, want := range cases {
		got, ok := primitiveZero(typeName)
		assert.True(t, ok, "expected %s to be a primitive", typeName)
		assert.Equal(t, want, got)
	}

	_, ok := primitiveZero("MyStruct")
	assert.False(t, ok)
}
