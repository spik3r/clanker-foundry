# Canonical artifact bundle v1

This archive checks consistency, not authenticity, approval or domain provenance.
It never extracts payloads. Domain producers must check their exact inventories
and independently rebuild semantic artifacts.

## API and limits

`File{Name string; Data []byte}` is a payload. `Build(profile string, files []File)
([]byte,error)` builds an archive. `Verify(raw []byte, profile string)
([]File,error)` returns sorted owned payloads, excluding the manifest. Inputs
must remain immutable during either call. Empty inventories, empty files and nil
file data are valid; nil data and empty data encode the same zero-byte payload.
Nil input inventory normalizes to an empty inventory; returned inventories are
nonnull empty slices. No filesystem, subprocess, network or clock use occurs.

Supported profiles are exactly `lifecycle-reviewpack/v1` and
`lifecycle-qualification/v1`. Unknown profiles fail. Limits apply together:
64 payloads, 8 MiB per payload, 32 MiB total payload bytes, 16 MiB whole archive,
64 KiB manifest and 192 bytes per payload name. Because storage is uncompressed,
the archive limit can be reached before the total-payload limit. Bounds are
checked before archive emission; errors return no archive or verified payloads.
Errors are fixed `ErrInvalid`, `ErrLimit` or `ErrProfile`, without input text.

## Names and inventories

Names are ASCII relative slash paths. Every segment starts with `[A-Za-z0-9]`,
continues with `[A-Za-z0-9._-]`, and does not end with a dot. Empty segments,
absolute paths, dot/dotdot segments, backslashes, drive letters and Unicode are
rejected. Case-insensitive Windows device stems CON/PRN/AUX/NUL and COM0-COM9 /
LPT0-LPT9, including stems before extensions, are rejected. Exact and
ASCII-case-insensitive duplicate names and file/directory prefix collisions fail.
`manifest.json` is reserved, case-insensitively, including its use as a path
prefix. Shared directory prefixes must use identical ASCII case. Payloads sort by bytewise name;
no case folding or text normalization changes their names or contents.

## Manifest

The first entry is `manifest.json`, UTF-8 compact JSON with one final LF and
fixed field order: formatVersion, profile, files, inventoryDigest. The version is
`lifecycle-artifact-bundle/v1`. Each sorted files record has exactly name, size
(integer byte count), sha256 (lowercase SHA-256 of exact payload bytes).

For inventoryDigest, serialize a compact fixed-order object with only
formatVersion, profile, files, using those same values and record order, without
a final LF. Hash the UTF-8 bytes of `lifecycle-artifact-inventory/v1`, one NUL
byte, then that JSON. This binds the profile and complete inventory without a
self-reference. Digests are lowercase hex. No timestamp or source path enters
identity. Empty files lists encode as arrays, never null.

Verification does not parse a manifest and trust its declarations. It rebuilds
the expected manifest from actual bounded payloads and the requested supported
profile, then requires complete canonical archive equality. Missing, extra,
reordered, unknown, duplicate or malformed manifest fields therefore fail, as do
altered payloads, hashes, sizes and profile claims.

## ZIP byte contract

Manifest first, then sorted payload entries; local and central directory order
are identical. ZIP method Store, flags0, Unix creator version20 (0x0314), reader
version20, DOS date33 (1980-01-01), DOS time0, regular file mode0100644 in external
attributes, internal attributes0, IEEE CRC32 of payload bytes, exact compressed
and uncompressed sizes, and no extras/comments/descriptors/encryption. Each
member has a raw local header with known CRC and size. No directory/symlink
entries, ZIP64, prepended/appended bytes or multiple disks are canonical.

The final 22-byte EOCD has disk0, equal disk/total entry counts, central size and
offset covering exactly the archive, and zero comment length. Verify checks its
bounded declared entry count and pre-scans the complete central-directory
record frontier (bounded names, no extra/comment fields, exact count and region)
before archive/zip allocates member records. It only reads Store
members with equal bounded compressed/uncompressed sizes, checks actual sizes
and CRC, recomputes the manifest, and compares deterministic rebuilt archive
bytes. This rejects alternate valid ZIP encodings and silent extra bytes.
