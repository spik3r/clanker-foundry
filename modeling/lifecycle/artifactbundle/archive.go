package artifactbundle

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"hash/crc32"
	"io"
	"sort"
	"strings"
)

type inventoryEntry struct {
	Name   string `json:"name"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
}
type inventory struct {
	FormatVersion string           `json:"formatVersion"`
	Profile       string           `json:"profile"`
	Files         []inventoryEntry `json:"files"`
}
type manifest struct {
	FormatVersion   string           `json:"formatVersion"`
	Profile         string           `json:"profile"`
	Files           []inventoryEntry `json:"files"`
	InventoryDigest string           `json:"inventoryDigest"`
}

func profileOK(p string) bool { return p == ReviewProfile || p == QualificationProfile }
func sha(b []byte) string     { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func validName(name string) bool {
	if len(name) == 0 || len(name) > MaxNameBytes || strings.ToLower(name) == ManifestName {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || part[len(part)-1] == '.' {
			return false
		}
		for i, c := range []byte(part) {
			alnum := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
			if !alnum && (i == 0 || c != '.' && c != '_' && c != '-') {
				return false
			}
		}
		stem := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" || len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && stem[3] >= '0' && stem[3] <= '9' {
			return false
		}
	}
	return true
}
func normalizedFiles(files []File) ([]File, error) {
	if len(files) > MaxFiles {
		return nil, ErrLimit
	}
	out := append([]File{}, files...)
	seen := map[string]bool{ManifestName: true}
	prefixes := map[string]string{}
	total := 0
	for _, f := range out {
		if !validName(f.Name) {
			return nil, ErrInvalid
		}
		fold := strings.ToLower(f.Name)
		if seen[fold] {
			return nil, ErrInvalid
		}
		seen[fold] = true
		parts := strings.Split(f.Name, "/")
		for i := 1; i < len(parts); i++ {
			prefix := strings.Join(parts[:i], "/")
			key := strings.ToLower(prefix)
			if prior, ok := prefixes[key]; ok && prior != prefix {
				return nil, ErrInvalid
			}
			prefixes[key] = prefix
		}
		if len(f.Data) > MaxFileBytes {
			return nil, ErrLimit
		}
		total += len(f.Data)
		if total > MaxTotalBytes {
			return nil, ErrLimit
		}
	}
	for _, f := range out {
		parts := strings.Split(strings.ToLower(f.Name), "/")
		for i := 1; i < len(parts); i++ {
			if seen[strings.Join(parts[:i], "/")] {
				return nil, ErrInvalid
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
func manifestBytes(profile string, files []File) ([]byte, error) {
	entries := []inventoryEntry{}
	for _, f := range files {
		entries = append(entries, inventoryEntry{f.Name, len(f.Data), sha(f.Data)})
	}
	canonical, e := json.Marshal(inventory{FormatVersion, profile, entries})
	if e != nil {
		return nil, ErrInvalid
	}
	digest := sha(append([]byte("lifecycle-artifact-inventory/v1\x00"), canonical...))
	b, e := json.Marshal(manifest{FormatVersion, profile, entries, digest})
	if e != nil {
		return nil, ErrInvalid
	}
	b = append(b, '\n')
	if len(b) > MaxManifestBytes {
		return nil, ErrLimit
	}
	return b, nil
}

// Build emits a canonical Store ZIP. It does not write files or run processes.
// Empty inventories and empty payloads are valid; domain inventories are separate.
func Build(profile string, files []File) ([]byte, error) {
	if !profileOK(profile) {
		return nil, ErrProfile
	}
	sorted, e := normalizedFiles(files)
	if e != nil {
		return nil, e
	}
	m, e := manifestBytes(profile, sorted)
	if e != nil {
		return nil, e
	}
	all := make([]File, 0, len(sorted)+1)
	all = append(all, File{ManifestName, m})
	all = append(all, sorted...)
	size := 22
	for _, f := range all {
		size += 76 + 2*len(f.Name) + len(f.Data)
	}
	if size > MaxArchiveBytes {
		return nil, ErrLimit
	}
	var b bytes.Buffer
	b.Grow(size)
	w := zip.NewWriter(&b)
	for _, f := range all {
		h := &zip.FileHeader{Name: f.Name, Method: zip.Store, CreatorVersion: 3<<8 | 20, ReaderVersion: 20, ModifiedDate: 33, CRC32: crc32.ChecksumIEEE(f.Data), CompressedSize64: uint64(len(f.Data)), UncompressedSize64: uint64(len(f.Data)), ExternalAttrs: 0100644 << 16}
		dst, e := w.CreateRaw(h)
		if e != nil {
			return nil, ErrInvalid
		}
		if _, e = dst.Write(f.Data); e != nil {
			return nil, ErrInvalid
		}
	}
	if e = w.Close(); e != nil {
		return nil, ErrInvalid
	}
	if b.Len() != size {
		return nil, ErrInvalid
	}
	return b.Bytes(), nil
}

// boundedDirectory checks the actual central-record frontier before archive/zip
// can allocate File objects. The EOCD count alone is not an allocation bound.
func boundedDirectory(raw []byte, count int) bool {
	end := raw[len(raw)-22:]
	start := uint64(binary.LittleEndian.Uint32(end[16:]))
	size := uint64(binary.LittleEndian.Uint32(end[12:]))
	if start+size != uint64(len(raw)-22) {
		return false
	}
	pos, stop := int(start), len(raw)-22
	for i := 0; i < count; i++ {
		if pos > stop-46 {
			return false
		}
		h := raw[pos : pos+46]
		if binary.LittleEndian.Uint32(h) != 0x02014b50 {
			return false
		}
		names := int(binary.LittleEndian.Uint16(h[28:]))
		extra := int(binary.LittleEndian.Uint16(h[30:]))
		comment := int(binary.LittleEndian.Uint16(h[32:]))
		if names < 1 || names > MaxNameBytes || extra != 0 || comment != 0 || names > stop-pos-46 {
			return false
		}
		pos += 46 + names
	}
	return pos == stop
}

// Verify checks bounded structure, reads actual payloads with CRC verification,
// recomputes all manifest identities and requires exact canonical rebuild bytes.
// It never trusts a claimed hash or extracts a file. Returned data is owned.
func Verify(raw []byte, profile string) ([]File, error) {
	if !profileOK(profile) {
		return nil, ErrProfile
	}
	if len(raw) > MaxArchiveBytes {
		return nil, ErrLimit
	}
	if len(raw) < 22 {
		return nil, ErrInvalid
	}
	// Canonical end record has no comment, multiple disks or ZIP64. Check entry
	// count before archive/zip allocates central-directory File objects.
	end := raw[len(raw)-22:]
	if binary.LittleEndian.Uint32(end) != 0x06054b50 || binary.LittleEndian.Uint16(end[4:]) != 0 || binary.LittleEndian.Uint16(end[6:]) != 0 || binary.LittleEndian.Uint16(end[20:]) != 0 {
		return nil, ErrInvalid
	}
	count := int(binary.LittleEndian.Uint16(end[10:]))
	if count < 1 || count > MaxFiles+1 {
		return nil, ErrLimit
	}
	if int(binary.LittleEndian.Uint16(end[8:])) != count || uint64(binary.LittleEndian.Uint32(end[12:]))+uint64(binary.LittleEndian.Uint32(end[16:]))+22 != uint64(len(raw)) {
		return nil, ErrInvalid
	}
	if !boundedDirectory(raw, count) {
		return nil, ErrInvalid
	}
	r, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil || len(r.File) != count {
		return nil, ErrInvalid
	}
	files := []File{}
	total := 0
	for i, f := range r.File {
		limit := MaxFileBytes
		if i == 0 {
			if f.Name != ManifestName {
				return nil, ErrInvalid
			}
			limit = MaxManifestBytes
		} else {
			if !validName(f.Name) {
				return nil, ErrInvalid
			}
		}
		if f.Method != zip.Store || f.Flags != 0 || !f.Mode().IsRegular() || f.UncompressedSize64 > uint64(limit) || f.CompressedSize64 != f.UncompressedSize64 {
			return nil, ErrInvalid
		}
		rc, e := f.Open()
		if e != nil {
			return nil, ErrInvalid
		}
		data, e := io.ReadAll(io.LimitReader(rc, int64(limit)+1))
		closeErr := rc.Close()
		if e != nil || closeErr != nil || len(data) > limit || uint64(len(data)) != f.UncompressedSize64 {
			return nil, ErrInvalid
		}
		if i > 0 {
			total += len(data)
			if total > MaxTotalBytes {
				return nil, ErrLimit
			}
			files = append(files, File{f.Name, data})
		}
	}
	canonical, e := Build(profile, files)
	if e != nil {
		return nil, e
	}
	if !bytes.Equal(raw, canonical) {
		return nil, ErrInvalid
	}
	// Build sorted a private slice; sort the owned return inventory too.
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return files, nil
}
