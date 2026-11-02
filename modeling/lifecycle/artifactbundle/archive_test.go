package artifactbundle

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"reflect"
	"strings"
	"testing"
)

func fixture() []File { return []File{{"z/b.bin", []byte{0, 1, 2}}, {"a.txt", []byte("alpha")}} }
func mustBuild(t *testing.T, p string, files []File) []byte {
	t.Helper()
	b, e := Build(p, files)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func reject(t *testing.T, b []byte, p string) {
	t.Helper()
	f, e := Verify(b, p)
	if e == nil || f != nil {
		t.Fatal("invalid archive returned payloads")
	}
}
func TestIndependentZIPVectorAndRepeatability(t *testing.T) {
	raw := mustBuild(t, ReviewProfile, fixture())
	// SHA256 independently constructed with Python struct/zlib/hashlib, including
	// literal ZIP headers and manifest; not computed from the implementation.
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "0adaf3c8f7260cb99db91b21e4e8d39f026177efb23e57266481b9f913d4bce0" {
		t.Fatal("canonical independent ZIP vector mismatch")
	}
	files := fixture()
	files[0], files[1] = files[1], files[0]
	if !bytes.Equal(raw, mustBuild(t, ReviewProfile, files)) {
		t.Fatal("input order changed bytes")
	}
	for i := 0; i < 3; i++ {
		verified, e := Verify(raw, ReviewProfile)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(verified, files) || !bytes.Equal(raw, mustBuild(t, ReviewProfile, verified)) {
			t.Fatal("sorted inventory/rebuild mismatch")
		}
	}
	for _, p := range []string{ReviewProfile, QualificationProfile} {
		b := mustBuild(t, p, nil)
		f, e := Verify(b, p)
		if e != nil || f == nil || len(f) != 0 {
			t.Fatal("empty inventory rejected")
		}
		f, e = Verify(mustBuild(t, p, []File{{"empty", nil}}), p)
		if e != nil || len(f) != 1 || len(f[0].Data) != 0 {
			t.Fatal("empty file rejected")
		}
	}
	reject(t, raw, QualificationProfile)
	reject(t, raw, "unknown/v1")
	if b, e := Build("unknown/v1", nil); e == nil || b != nil {
		t.Fatal("unknown profile build")
	}
}
func TestNamesAndCollisions(t *testing.T) {
	for _, name := range []string{"", "/a", "a/", "a//b", "./a", "a/../b", "a\\b", "C:a", "a.", ".hidden", "a/CON.txt", "NUL", "com9.dat", "a/雪", "manifest.json", "MANIFEST.JSON", strings.Repeat("a", MaxNameBytes+1)} {
		if b, e := Build(ReviewProfile, []File{{name, nil}}); e == nil || b != nil {
			t.Fatal("bad name accepted", name)
		}
	}
	for _, names := range [][]string{{"a", "a"}, {"a", "A"}, {"a", "a/b"}, {"A", "a/b"}, {"a/x", "A/y"}, {"manifest.json/a", "other"}, {"MANIFEST.JSON/a", "other"}} {
		f := []File{}
		for _, n := range names {
			f = append(f, File{n, nil})
		}
		if b, e := Build(ReviewProfile, f); e == nil || b != nil {
			t.Fatal("collision accepted", names)
		}
	}
	name := strings.Repeat("a", MaxNameBytes)
	if _, e := Verify(mustBuild(t, ReviewProfile, []File{{name, nil}}), ReviewProfile); e != nil {
		t.Fatal("maximum name rejected")
	}
}

// rewrite recomputes CRC and sizes independently, so manifest/payload tests do
// not merely exercise the ZIP CRC check. It retains otherwise canonical headers.
func rewrite(t *testing.T, raw []byte, mutate func([]File) []File) []byte {
	t.Helper()
	r, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		t.Fatal(e)
	}
	all := []File{}
	for _, f := range r.File {
		rc, e := f.Open()
		if e != nil {
			t.Fatal(e)
		}
		b, e := io.ReadAll(rc)
		rc.Close()
		if e != nil {
			t.Fatal(e)
		}
		all = append(all, File{f.Name, b})
	}
	all = mutate(all)
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, f := range all {
		h := &zip.FileHeader{Name: f.Name, Method: zip.Store, CreatorVersion: 0x314, ReaderVersion: 20, ModifiedDate: 33, CRC32: crc32.ChecksumIEEE(f.Data), CompressedSize64: uint64(len(f.Data)), UncompressedSize64: uint64(len(f.Data)), ExternalAttrs: 0100644 << 16}
		dst, e := w.CreateRaw(h)
		if e != nil {
			t.Fatal(e)
		}
		dst.Write(f.Data)
	}
	if e := w.Close(); e != nil {
		t.Fatal(e)
	}
	return buf.Bytes()
}
func TestTamperingWithValidZIPCRC(t *testing.T) {
	raw := mustBuild(t, ReviewProfile, fixture())
	mutations := map[string]func([]File) []File{
		"payload": func(f []File) []File { f[1].Data = []byte("other"); return f },
		"manifest size": func(f []File) []File {
			f[0].Data = bytes.Replace(f[0].Data, []byte(`"size":5`), []byte(`"size":4`), 1)
			return f
		},
		"manifest hash": func(f []File) []File {
			f[0].Data = bytes.Replace(f[0].Data, []byte("8ed3f6ad"), []byte("0ed3f6ad"), 1)
			return f
		},
		"inventory digest": func(f []File) []File {
			f[0].Data = bytes.Replace(f[0].Data, []byte("b79df847"), []byte("079df847"), 1)
			return f
		},
		"unknown manifest field":   func(f []File) []File { f[0].Data = append([]byte(`{"extra":true,`), f[0].Data[1:]...); return f },
		"malformed later manifest": func(f []File) []File { f[0].Data = []byte("not json"); return f },
		"missing":                  func(f []File) []File { return f[:2] },
		"extra":                    func(f []File) []File { return append(f, File{"extra", nil}) },
		"duplicate":                func(f []File) []File { return append(f, f[1]) },
		"reordered":                func(f []File) []File { f[1], f[2] = f[2], f[1]; return f },
		"reserved alias":           func(f []File) []File { f[1].Name = "MANIFEST.JSON"; return f },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) { reject(t, rewrite(t, raw, mutate), ReviewProfile) })
	}
}
func TestNoncanonicalMetadataAndRawLimits(t *testing.T) {
	raw := mustBuild(t, ReviewProfile, fixture())
	cd := int(binary.LittleEndian.Uint32(raw[len(raw)-6:]))
	for _, offset := range []int{4, 6, 8, 10, 12, cd + 4, cd + 6, cd + 8, cd + 10, cd + 12, cd + 14, cd + 38} {
		b := append([]byte{}, raw...)
		b[offset] ^= 1
		reject(t, b, ReviewProfile)
	}
	reject(t, append(append([]byte{}, raw...), 0), ReviewProfile)
	reject(t, append([]byte{0}, raw...), ReviewProfile)
	reject(t, raw[:len(raw)-1], ReviewProfile)
	reject(t, make([]byte, MaxArchiveBytes+1), ReviewProfile)
	b := append([]byte{}, raw...)
	binary.LittleEndian.PutUint16(b[len(b)-12:], 65535)
	reject(t, b, ReviewProfile)
	var z bytes.Buffer
	w := zip.NewWriter(&z)
	f, _ := w.Create("manifest.json")
	f.Write([]byte("compressed"))
	w.Close()
	reject(t, z.Bytes(), ReviewProfile)
	oversized := rewrite(t, raw, func(f []File) []File { f[0].Data = make([]byte, MaxManifestBytes+1); return f })
	reject(t, oversized, ReviewProfile)
}
func TestPayloadAndArchiveBoundaries(t *testing.T) {
	files := []File{}
	for i := 0; i < MaxFiles; i++ {
		files = append(files, File{fmt.Sprintf("f%02d", i), nil})
	}
	mustBuild(t, ReviewProfile, files)
	if b, e := Build(ReviewProfile, append(files, File{"overflow", nil})); e == nil || b != nil {
		t.Fatal("file count limit")
	}
	mustBuild(t, ReviewProfile, []File{{"large", make([]byte, MaxFileBytes)}})
	if b, e := Build(ReviewProfile, []File{{"large", make([]byte, MaxFileBytes+1)}}); e == nil || b != nil {
		t.Fatal("file bytes limit")
	}
	files = []File{{"a", make([]byte, MaxFileBytes)}, {"b", make([]byte, MaxFileBytes-4096)}}
	raw := mustBuild(t, ReviewProfile, files)
	files[1].Data = append(files[1].Data, make([]byte, MaxArchiveBytes-len(raw))...)
	raw = mustBuild(t, ReviewProfile, files)
	if len(raw) != MaxArchiveBytes {
		t.Fatal("exact archive bound")
	}
	if _, e := Verify(raw, ReviewProfile); e != nil {
		t.Fatal("maximum archive rejected", e)
	}
	files[1].Data = append(files[1].Data, 0)
	if b, e := Build(ReviewProfile, files); e == nil || b != nil {
		t.Fatal("archive one beyond")
	}
	// Store archives reach their lower archive cap first; this test checks that
	// >32MiB payload declarations also refuse without a partial result.
	large := make([]byte, MaxFileBytes)
	files = []File{{"a", large}, {"b", large}, {"c", large}, {"d", large}, {"e", []byte{0}}}
	if b, e := Build(ReviewProfile, files); e == nil || b != nil {
		t.Fatal("total bytes limit")
	}
}
func TestInputsAndVerifiedResultsDoNotAlias(t *testing.T) {
	files := fixture()
	raw := mustBuild(t, ReviewProfile, files)
	snapshot := append([]byte{}, raw...)
	files[0].Data[0] = 99
	files[0].Name = "changed"
	if !bytes.Equal(raw, snapshot) {
		t.Fatal("Build output aliases input")
	}
	a, e := Verify(raw, ReviewProfile)
	if e != nil {
		t.Fatal(e)
	}
	b, e := Verify(raw, ReviewProfile)
	if e != nil {
		t.Fatal(e)
	}
	a[0].Data[0] = '!'
	a[0].Name = "changed"
	if !bytes.Equal(raw, snapshot) || string(b[0].Data) != "alpha" || b[0].Name != "a.txt" {
		t.Fatal("Verify result aliases source/other result")
	}
}

func TestForgedSmallCountCannotHideManyCentralRecords(t *testing.T) {
	raw := mustBuild(t, ReviewProfile, fixture())
	cd := int(binary.LittleEndian.Uint32(raw[len(raw)-6:]))
	nameLen := int(binary.LittleEndian.Uint16(raw[cd+28:]))
	record := raw[cd : cd+46+nameLen]
	central := bytes.Repeat(record, 65537)
	forged := append([]byte{}, raw[:cd]...)
	forged = append(forged, central...)
	end := append([]byte{}, raw[len(raw)-22:]...)
	binary.LittleEndian.PutUint16(end[8:], 1)
	binary.LittleEndian.PutUint16(end[10:], 1)
	binary.LittleEndian.PutUint32(end[12:], uint32(len(central)))
	binary.LittleEndian.PutUint32(end[16:], uint32(cd))
	forged = append(forged, end...)
	if boundedDirectory(forged, 1) {
		t.Fatal("forged declared count bypassed actual central frontier bound")
	}
	reject(t, forged, ReviewProfile)
}
