// This file was generated by nomdl/codegen.

package common

import (
	"github.com/attic-labs/noms/chunks"
	"github.com/attic-labs/noms/ref"
	"github.com/attic-labs/noms/types"
)

var __commonPackageInFile_date_CachedRef ref.Ref

// This function builds up a Noms value that describes the type
// package implemented by this file and registers it with the global
// type package definition cache.
func init() {
	p := types.NewPackage([]types.Type{
		types.MakeStructType("Date",
			[]types.Field{
				types.Field{"Unix", types.MakePrimitiveType(types.Int64Kind), false},
			},
			types.Choices{},
		),
	}, []ref.Ref{})
	__commonPackageInFile_date_CachedRef = types.RegisterPackage(&p)
}

// Date

type Date struct {
	_Unix int64

	cs  chunks.ChunkStore
	ref *ref.Ref
}

func NewDate(cs chunks.ChunkStore) Date {
	return Date{
		_Unix: int64(0),

		cs:  cs,
		ref: &ref.Ref{},
	}
}

type DateDef struct {
	Unix int64
}

func (def DateDef) New(cs chunks.ChunkStore) Date {
	return Date{
		_Unix: def.Unix,
		cs:    cs,
		ref:   &ref.Ref{},
	}
}

func (s Date) Def() (d DateDef) {
	d.Unix = s._Unix
	return
}

var __typeForDate types.Type

func (m Date) Type() types.Type {
	return __typeForDate
}

func init() {
	__typeForDate = types.MakeType(__commonPackageInFile_date_CachedRef, 0)
	types.RegisterStruct(__typeForDate, builderForDate, readerForDate)
}

func builderForDate(cs chunks.ChunkStore, values []types.Value) types.Value {
	i := 0
	s := Date{ref: &ref.Ref{}, cs: cs}
	s._Unix = int64(values[i].(types.Int64))
	i++
	return s
}

func readerForDate(v types.Value) []types.Value {
	values := []types.Value{}
	s := v.(Date)
	values = append(values, types.Int64(s._Unix))
	return values
}

func (s Date) Equals(other types.Value) bool {
	return other != nil && __typeForDate.Equals(other.Type()) && s.Ref() == other.Ref()
}

func (s Date) Ref() ref.Ref {
	return types.EnsureRef(s.ref, s)
}

func (s Date) Chunks() (chunks []ref.Ref) {
	chunks = append(chunks, __typeForDate.Chunks()...)
	return
}

func (s Date) ChildValues() (ret []types.Value) {
	ret = append(ret, types.Int64(s._Unix))
	return
}

func (s Date) Unix() int64 {
	return s._Unix
}

func (s Date) SetUnix(val int64) Date {
	s._Unix = val
	s.ref = &ref.Ref{}
	return s
}