// Package designchange assembles consistency-checked offline model change projects.
// Supplied contents may be private. Suites are unvalidated and unexecuted, and
// their binding is not inspected. No archive grants review or execution authority.
package designchange

import (
	"encoding/json"
	"errors"
)

const (
	Profile             = "lifecycle-designchange/v1"
	FormatVersion       = "lifecycle-designchange-json/v1"
	VerificationVersion = "lifecycle-designchange-verification/v1"
	MaxRequestBytes     = 3 << 20
	MaxSuiteBytes       = 1 << 20
	MaxSourceBytes      = 1 << 20
	MaxSourcesBytes     = 4 << 20
	MaxInnerBytes       = 8 << 20
	MaxArchiveBytes     = 16 << 20
	MaxCarrierBytes     = 32 << 20
	MaxCarrierDepth     = 16
	maxModelBytes       = 1 << 20
	maxMetadataBytes    = 8 << 10
	maxGuideBytes       = 4 << 10
)

var (
	ErrInvalid = errors.New("invalid design-change project")
	ErrLimit   = errors.New("design-change limit exceeded")
)

// Expected contains independent caller-held identities, never archive assertions.
// All fields are mandatory lowercase SHA-256 hex strings.
type Expected struct {
	RequestSHA256        string `json:"requestSHA256"`
	BeforeSourceSHA256   string `json:"beforeSourceSHA256"`
	BeforeModelDigest    string `json:"beforeModelDigest"`
	CandidateSHA256      string `json:"candidateSHA256"`
	CandidateModelDigest string `json:"candidateModelDigest"`
	SuiteSHA256          string `json:"suiteSHA256"`
}

type sourceDTO struct {
	Name       string `json:"name"`
	DataBase64 string `json:"dataBase64"`
}
type carrierDTO struct {
	FormatVersion string       `json:"formatVersion"`
	Operation     string       `json:"operation"`
	RequestBase64 *string      `json:"requestBase64"`
	SuiteBase64   *string      `json:"suiteBase64"`
	Sources       *[]sourceDTO `json:"sources"`
	ArchiveBase64 *string      `json:"archiveBase64"`
	Expected      *Expected    `json:"expected"`
}
type plannerResultDTO struct {
	ReportVersion  string          `json:"reportVersion"`
	ProfileVersion string          `json:"profileVersion"`
	Mode           string          `json:"mode"`
	Authority      string          `json:"authority"`
	BeforeDigest   string          `json:"beforeDigest"`
	AfterDigest    string          `json:"afterDigest"`
	After          json.RawMessage `json:"after"`
	Review         json.RawMessage `json:"review"`
}
type metadataDTO struct {
	FormatVersion               string `json:"formatVersion"`
	Mode                        string `json:"mode"`
	Authority                   string `json:"authority"`
	RequestSHA256               string `json:"requestSHA256"`
	BeforeSourceSHA256          string `json:"beforeSourceSHA256"`
	BeforeModelDigest           string `json:"beforeModelDigest"`
	CandidateSHA256             string `json:"candidateSHA256"`
	CandidateModelDigest        string `json:"candidateModelDigest"`
	SuiteSHA256                 string `json:"suiteSHA256"`
	PlannerResultSHA256         string `json:"plannerResultSHA256"`
	QualificationSHA256         string `json:"qualificationSHA256"`
	ModelFormatVersion          string `json:"modelFormatVersion"`
	ProfileVersion              string `json:"profileVersion"`
	GeneratorVersion            string `json:"generatorVersion"`
	ToolchainVersion            string `json:"toolchainVersion"`
	QualificationSourceRevision string `json:"qualificationSourceRevision"`
	SuiteStatus                 string `json:"suiteStatus"`
	SuiteBindingStatus          string `json:"suiteBindingStatus"`
}
type verificationDTO struct {
	FormatVersion      string `json:"formatVersion"`
	Consistent         bool   `json:"consistent"`
	Mode               string `json:"mode"`
	Authority          string `json:"authority"`
	SuiteStatus        string `json:"suiteStatus"`
	SuiteBindingStatus string `json:"suiteBindingStatus"`
}
